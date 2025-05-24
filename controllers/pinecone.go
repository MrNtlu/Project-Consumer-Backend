package controllers

import (
	"app/db"
	"app/models"
	"app/responses"
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/structpb"
)

type PineconeController struct {
	Database      *db.MongoDB
	Pinecone      *pinecone.Client
	PineconeIndex *pinecone.IndexConnection
	UserListModel *models.UserListModel
}

func NewPineconeController(
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
) PineconeController {
	return PineconeController{
		Database:      mongoDB,
		Pinecone:      pinecone,
		PineconeIndex: pineconeIndex,
		UserListModel: models.NewUserListModel(mongoDB),
	}
}

// GetRecommendationsByType gets recommendations for a user based on their content lists
func (pc *PineconeController) GetRecommendationsByType(userID string, topK int) (map[string]interface{}, error) {
	// Get user content lists using the new aggregation function
	userLists, err := pc.UserListModel.GetUserListIDForSuggestion(userID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"userID": userID,
			"error":  err,
		}).Error("failed to get user list for recommendation")
		return nil, fmt.Errorf("failed to get user list for recommendation: %w", err)
	}

	// Create content ID sets for checking if user already has content
	movieIDSet := make(map[string]bool)
	for _, movie := range userLists.MovieIDList {
		movieIDSet[movie.ID] = true
	}

	tvIDSet := make(map[string]bool)
	for _, tv := range userLists.TVIDList {
		tvIDSet[tv.ID] = true
	}

	animeIDSet := make(map[string]bool)
	for _, anime := range userLists.AnimeIDList {
		animeIDSet[anime.ID] = true
	}

	gameIDSet := make(map[string]bool)
	for _, game := range userLists.GameIDList {
		gameIDSet[game.ID] = true
	}

	// Fetch content details for recommendations
	movieDetails, err := pc.fetchContentDetails(userLists.MovieIDList, "movies")
	if err != nil {
		logrus.WithField("error", err).Error("failed to fetch movie details")
	}

	tvDetails, err := pc.fetchContentDetails(userLists.TVIDList, "tv-series")
	if err != nil {
		logrus.WithField("error", err).Error("failed to fetch TV series details")
	}

	animeDetails, err := pc.fetchContentDetails(userLists.AnimeIDList, "animes")
	if err != nil {
		logrus.WithField("error", err).Error("failed to fetch anime details")
	}

	gameDetails, err := pc.fetchContentDetails(userLists.GameIDList, "games")
	if err != nil {
		logrus.WithField("error", err).Error("failed to fetch game details")
	}

	// Log content counts for debugging
	logrus.WithFields(logrus.Fields{
		"movieCount": len(movieDetails),
		"tvCount":    len(tvDetails),
		"animeCount": len(animeDetails),
		"gameCount":  len(gameDetails),
		"totalItems": len(movieDetails) + len(tvDetails) + len(animeDetails) + len(gameDetails),
	}).Info("content details retrieved for recommendations")

	// Get recommendations for each content type
	movieRecs := pc.getRecommendationsForContentType(movieDetails, "movie", topK, movieIDSet)
	tvRecs := pc.getRecommendationsForContentType(tvDetails, "tvseries", topK, tvIDSet)
	animeRecs := pc.getRecommendationsForContentType(animeDetails, "anime", topK, animeIDSet)
	gameRecs := pc.getRecommendationsForContentType(gameDetails, "game", topK, gameIDSet)

	// Combine all recommendations
	allRecs := append(append(append(movieRecs, tvRecs...), animeRecs...), gameRecs...)

	return map[string]interface{}{
		"movies":   movieRecs,
		"tvSeries": tvRecs,
		"animes":   animeRecs,
		"games":    gameRecs,
		"all":      allRecs,
	}, nil
}

// fetchContentDetails fetches content details from MongoDB for the given IDs
func (pc *PineconeController) fetchContentDetails(contentIDs []responses.UserListAISuggestionID, collectionName string) ([]bson.M, error) {
	if len(contentIDs) == 0 {
		return []bson.M{}, nil
	}

	objectIDQueries := make([]bson.M, 0, len(contentIDs))
	stringIDQueries := make([]bson.M, 0, len(contentIDs))

	for _, idObj := range contentIDs {
		if idObj.ID == "" {
			continue
		}

		if objectID, err := primitive.ObjectIDFromHex(idObj.ID); err == nil {
			objectIDQueries = append(objectIDQueries, bson.M{"_id": objectID})
		} else {
			stringIDQueries = append(stringIDQueries, bson.M{"_id": idObj.ID})
		}
	}

	if len(objectIDQueries) == 0 && len(stringIDQueries) == 0 {
		return []bson.M{}, nil
	}

	query := bson.M{
		"$or": append(objectIDQueries, stringIDQueries...),
	}

	collection := pc.Database.Database.Collection(collectionName)
	cursor, err := collection.Find(context.TODO(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to find content in collection %s: %w", collectionName, err)
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, fmt.Errorf("failed to decode content results: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"collection": collectionName,
		"found":      len(results),
		"requested":  len(contentIDs),
	}).Info("fetched content details")

	return results, nil
}

// getRecommendationsForContentType gets recommendations for a specific content type
func (pc *PineconeController) getRecommendationsForContentType(contentData []bson.M, contentType string, limit int, userContentSet map[string]bool) []map[string]interface{} {
	if len(contentData) == 0 {
		return []map[string]interface{}{}
	}

	// Get a sample of the user's content (up to 3 items) for generating recommendations
	sampleSize := 3
	if len(contentData) < sampleSize {
		sampleSize = len(contentData)
	}

	recommendations := []map[string]interface{}{}

	// For each sample item, get recommendations
	for i := 0; i < sampleSize; i++ {
		contentID := ""

		// Extract the ID from the content item
		if id, ok := contentData[i]["_id"].(primitive.ObjectID); ok {
			contentID = id.Hex()
		} else if id, ok := contentData[i]["_id"].(string); ok {
			contentID = id
		}

		if contentID == "" {
			continue
		}

		// Get similar items from Pinecone
		similarItems, err := pc.GetSimilarItemsFromPinecone(contentID, contentType, limit/sampleSize*2)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"contentID": contentID,
				"error":     err,
			}).Error("failed to get similar items from Pinecone")
			continue
		}

		// Filter out items the user already has
		for _, item := range similarItems {
			itemID, ok := item["id"].(string)
			if !ok || itemID == "" || userContentSet[itemID] {
				continue
			}

			// Check if this item is already in our recommendations
			isDuplicate := false
			for _, rec := range recommendations {
				if recID, ok := rec["id"].(string); ok && recID == itemID {
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				recommendations = append(recommendations, item)

				// Break if we have enough recommendations
				if len(recommendations) >= limit {
					break
				}
			}
		}

		// Break if we have enough recommendations
		if len(recommendations) >= limit {
			break
		}
	}

	// Return the recommendations with the right content type
	for i := range recommendations {
		if recommendations[i]["type"] == nil {
			recommendations[i]["type"] = contentType
		}
	}

	return recommendations
}

// GetRecommendationsForType gets recommendations for a specific content type
func (pc *PineconeController) GetRecommendationsForType(contentData []bson.M, contentType string, perTypeCount int, userContentSet map[string]bool) ([]map[string]interface{}, error) {
	if len(contentData) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Step 1: Find sequels and series first
	sequelRecs, err := pc.FindSequelsAndSeries(contentData, contentType, userContentSet)
	if err != nil {
		logrus.WithField("error", err).Error("error finding sequels and series")
	}

	// If we have enough sequel recommendations, return them
	if len(sequelRecs) >= perTypeCount {
		return sequelRecs[:perTypeCount], nil
	}

	// Step 2: Get similarity-based recommendations for remaining slots
	remainingCount := perTypeCount - len(sequelRecs)

	// Use a sample of user's content for recommendations (up to 3 items)
	sampleSize := int(math.Min(float64(len(contentData)), 3))
	similarityRecs := []map[string]interface{}{}

	for i := 0; i < sampleSize; i++ {
		contentID := ""
		if id, ok := contentData[i]["_id"].(primitive.ObjectID); ok {
			contentID = id.Hex()
		} else if id, ok := contentData[i]["_id"].(string); ok {
			contentID = id
		}

		if contentID == "" {
			continue
		}

		// Get vector similarity recommendations from Pinecone
		// This would need to be implemented based on your Pinecone setup
		similarItems, err := pc.getContentSimilarityRecommendations(contentID, contentType, int(math.Ceil(float64(remainingCount)/float64(sampleSize))*2))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"contentID":   contentID,
				"contentType": contentType,
				"error":       err,
			}).Error("failed to get similarity recommendations")
			continue
		}

		// Filter recommendations to avoid duplicates and user's existing content
		for _, item := range similarItems {
			itemID := ""
			if id, ok := item["id"].(string); ok {
				itemID = id
			}

			if itemID == "" || userContentSet[itemID] {
				continue
			}

			// Check if already in sequel recommendations
			alreadyInSequels := false
			for _, seqRec := range sequelRecs {
				if seqID, ok := seqRec["id"].(string); ok && seqID == itemID {
					alreadyInSequels = true
					break
				}
			}

			if !alreadyInSequels {
				similarityRecs = append(similarityRecs, item)
				if len(similarityRecs) >= remainingCount {
					break
				}
			}
		}
	}

	// Combine sequel and similarity recommendations
	return append(sequelRecs, similarityRecs...), nil
}

// FindSequelsAndSeries finds sequels and series content not yet consumed by user
func (pc *PineconeController) FindSequelsAndSeries(contentData []bson.M, contentType string, userContentSet map[string]bool) ([]map[string]interface{}, error) {
	if len(contentData) == 0 {
		return []map[string]interface{}{}, nil
	}

	recommendations := []map[string]interface{}{}
	collection := pc.GetCollectionForContentType(contentType)

	for _, content := range contentData {
		// Get the title to search for
		title := ""
		if val, ok := content["title_original"].(string); ok && val != "" {
			title = val
		} else if val, ok := content["title_en"].(string); ok && val != "" {
			title = val
		} else if val, ok := content["title"].(string); ok && val != "" {
			title = val
		}

		if title == "" {
			continue
		}

		// Extract series information
		seriesInfo := pc.ExtractSeriesInfo(title)
		if seriesInfo["seriesName"] == "" {
			continue
		}

		// Escape special characters for regex
		escapedSeriesName := regexp.QuoteMeta(seriesInfo["seriesName"])

		// Search for other content in the same series
		query := bson.M{
			"$and": []bson.M{
				{
					"$or": []bson.M{
						{"title": bson.M{"$regex": escapedSeriesName, "$options": "i"}},
						{"title_en": bson.M{"$regex": escapedSeriesName, "$options": "i"}},
						{"title_original": bson.M{"$regex": escapedSeriesName, "$options": "i"}},
					},
				},
				{
					"_id": bson.M{"$ne": content["_id"]},
				},
			},
		}

		opts := options.Find().SetLimit(5)
		cursor, err := collection.Find(context.TODO(), query, opts)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"contentType": contentType,
				"title":       title,
				"error":       err,
			}).Error("failed to find series content")
			continue
		}

		var potentialSeriesContent []bson.M
		if err = cursor.All(context.TODO(), &potentialSeriesContent); err != nil {
			cursor.Close(context.TODO())
			continue
		}
		cursor.Close(context.TODO())

		// Filter out content the user already has
		var filteredSeriesContent []bson.M
		for _, item := range potentialSeriesContent {
			contentIDStr := ""
			if contentID, ok := item["_id"].(primitive.ObjectID); ok {
				contentIDStr = contentID.Hex()
			} else if contentID, ok := item["_id"].(string); ok {
				contentIDStr = contentID
			}

			if contentIDStr != "" && !userContentSet[contentIDStr] {
				filteredSeriesContent = append(filteredSeriesContent, item)
			}
		}

		// Score and sort by metadata similarity
		type scoredContent struct {
			item  bson.M
			score float64
		}

		var scoredItems []scoredContent
		for _, item := range filteredSeriesContent {
			score := pc.CalculateMetadataSimilarity(content, item, contentType)
			scoredItems = append(scoredItems, scoredContent{item: item, score: score})
		}

		// Sort by score (higher first)
		for i := 0; i < len(scoredItems); i++ {
			for j := i + 1; j < len(scoredItems); j++ {
				if scoredItems[i].score < scoredItems[j].score {
					scoredItems[i], scoredItems[j] = scoredItems[j], scoredItems[i]
				}
			}
		}

		// Add to recommendations
		for _, scoredItem := range scoredItems {
			contentIDStr := ""
			if contentID, ok := scoredItem.item["_id"].(primitive.ObjectID); ok {
				contentIDStr = contentID.Hex()
			} else if contentID, ok := scoredItem.item["_id"].(string); ok {
				contentIDStr = contentID
			}

			if contentIDStr != "" {
				recommendations = append(recommendations, map[string]interface{}{
					"id":    contentIDStr,
					"score": scoredItem.score,
					"type":  contentType,
					"data":  scoredItem.item,
				})

				// Limit to 3 series recommendations per content
				if len(recommendations) >= 3 {
					break
				}
			}
		}

		// Limit total sequel recommendations
		if len(recommendations) >= 10 {
			break
		}
	}

	return recommendations, nil
}

// ExtractSeriesInfo extracts series name and number/season from title
func (pc *PineconeController) ExtractSeriesInfo(title string) map[string]string {
	// Patterns to match series titles
	numberPattern := regexp.MustCompile(`^(.*?)(?:\s+|-|:)(?:\d+|[IVXLCDM]+)(?:\s+|-|:|$)`)
	seasonPattern := regexp.MustCompile(`^(.*?)(?:\s+|-|:)(?:season|part|chapter)(?:\s+|-|:)(?:\d+|[IVXLCDM]+)(?:\s+|-|:|$)`)
	franchisePattern := regexp.MustCompile(`^(.*?)(?:\s+|-|:)(?:the|a|an|origins|returns|rises|forever|begins)(?:\s+|-|:|$)`)

	seriesInfo := map[string]string{
		"seriesName":   "",
		"seriesNumber": "",
	}

	// Try each pattern
	var seriesName string
	if match := numberPattern.FindStringSubmatch(title); len(match) > 1 {
		seriesName = strings.TrimSpace(match[1])
	} else if match := seasonPattern.FindStringSubmatch(title); len(match) > 1 {
		seriesName = strings.TrimSpace(match[1])
	} else if match := franchisePattern.FindStringSubmatch(title); len(match) > 1 {
		seriesName = strings.TrimSpace(match[1])
	} else if len(title) > 4 {
		// If no pattern matched but title is long enough, use whole title as series name
		seriesName = title
	}

	if seriesName != "" {
		seriesInfo["seriesName"] = seriesName

		// Try to extract number if present
		numberMatch := regexp.MustCompile(`\d+|[IVXLCDM]+`).FindString(title)
		if numberMatch != "" {
			seriesInfo["seriesNumber"] = numberMatch
		}
	}

	return seriesInfo
}

// CalculateMetadataSimilarity calculates similarity score between content items based on metadata
func (pc *PineconeController) CalculateMetadataSimilarity(content1, content2 bson.M, contentType string) float64 {
	score := 0.5 // Base similarity score

	switch contentType {
	case "anime":
		// Compare genres
		genres1 := pc.GetFieldArrayValues(content1, "genres", "name")
		genres2 := pc.GetFieldArrayValues(content2, "genres", "name")
		score += pc.CompareArrays(genres1, genres2) * 0.15

		// Compare demographics
		demo1 := pc.GetFieldArrayValues(content1, "demographics", "name")
		demo2 := pc.GetFieldArrayValues(content2, "demographics", "name")
		score += pc.CompareArrays(demo1, demo2) * 0.15

		// Compare themes
		themes1 := pc.GetFieldArrayValues(content1, "themes", "name")
		themes2 := pc.GetFieldArrayValues(content2, "themes", "name")
		score += pc.CompareArrays(themes1, themes2) * 0.10

		// Compare studios
		studios1 := pc.GetFieldArrayValues(content1, "studios", "name")
		studios2 := pc.GetFieldArrayValues(content2, "studios", "name")
		score += pc.CompareArrays(studios1, studios2) * 0.10

	case "movie":
		// Compare genres
		var genres1, genres2 []string
		if g, ok := content1["genres"].(primitive.A); ok {
			for _, genre := range g {
				if genreStr, ok := genre.(string); ok {
					genres1 = append(genres1, genreStr)
				}
			}
		}
		if g, ok := content2["genres"].(primitive.A); ok {
			for _, genre := range g {
				if genreStr, ok := genre.(string); ok {
					genres2 = append(genres2, genreStr)
				}
			}
		}
		score += pc.CompareArrays(genres1, genres2) * 0.20

		// Compare production companies
		prodComp1 := pc.GetFieldArrayValues(content1, "production_companies", "name")
		prodComp2 := pc.GetFieldArrayValues(content2, "production_companies", "name")
		score += pc.CompareArrays(prodComp1, prodComp2) * 0.10

		// Compare actors
		actors1 := pc.GetFieldArrayValues(content1, "actors", "name")
		actors2 := pc.GetFieldArrayValues(content2, "actors", "name")
		if len(actors1) > 3 {
			actors1 = actors1[:3]
		}
		if len(actors2) > 3 {
			actors2 = actors2[:3]
		}
		score += pc.CompareArrays(actors1, actors2) * 0.20

	case "tvseries":
		// Similar to movie case but with TV-specific fields
		var genres1, genres2 []string
		if g, ok := content1["genres"].(primitive.A); ok {
			for _, genre := range g {
				if genreStr, ok := genre.(string); ok {
					genres1 = append(genres1, genreStr)
				}
			}
		}
		if g, ok := content2["genres"].(primitive.A); ok {
			for _, genre := range g {
				if genreStr, ok := genre.(string); ok {
					genres2 = append(genres2, genreStr)
				}
			}
		}
		score += pc.CompareArrays(genres1, genres2) * 0.20

		// Compare networks
		networks1 := pc.GetFieldArrayValues(content1, "networks", "name")
		networks2 := pc.GetFieldArrayValues(content2, "networks", "name")
		score += pc.CompareArrays(networks1, networks2) * 0.15

	case "game":
		// Compare genres
		var genres1, genres2 []string
		if g, ok := content1["genres"].(primitive.A); ok {
			for _, genre := range g {
				if genreStr, ok := genre.(string); ok {
					genres1 = append(genres1, genreStr)
				}
			}
		}
		if g, ok := content2["genres"].(primitive.A); ok {
			for _, genre := range g {
				if genreStr, ok := genre.(string); ok {
					genres2 = append(genres2, genreStr)
				}
			}
		}
		score += pc.CompareArrays(genres1, genres2) * 0.20

		// Compare platforms, developers, publishers
		platforms1 := pc.GetFieldArrayValues(content1, "platforms", "")
		platforms2 := pc.GetFieldArrayValues(content2, "platforms", "")
		score += pc.CompareArrays(platforms1, platforms2) * 0.15

		developers1 := pc.GetFieldArrayValues(content1, "developers", "")
		developers2 := pc.GetFieldArrayValues(content2, "developers", "")
		score += pc.CompareArrays(developers1, developers2) * 0.10
	}

	// Cap at 0.99 to avoid exact match confusion
	if score > 0.99 {
		return 0.99
	}
	return score
}

// CompareArrays calculates similarity between two arrays
func (pc *PineconeController) CompareArrays(array1, array2 []string) float64 {
	if len(array1) == 0 || len(array2) == 0 {
		return 0
	}

	// Create sets
	set1 := make(map[string]bool)
	for _, item := range array1 {
		set1[item] = true
	}

	set2 := make(map[string]bool)
	for _, item := range array2 {
		set2[item] = true
	}

	// Calculate intersection
	intersectionCount := 0
	for item := range set1 {
		if set2[item] {
			intersectionCount++
		}
	}

	// Calculate union
	unionCount := len(set1) + len(set2) - intersectionCount

	if unionCount == 0 {
		return 0
	}
	return float64(intersectionCount) / float64(unionCount)
}

// GetFieldArrayValues extracts array field values
func (pc *PineconeController) GetFieldArrayValues(obj bson.M, fieldName string, subField string) []string {
	var result []string

	if array, ok := obj[fieldName].(primitive.A); ok {
		for _, item := range array {
			if subField == "" {
				if itemStr, ok := item.(string); ok {
					result = append(result, itemStr)
				}
			} else if itemMap, ok := item.(bson.M); ok {
				if itemValue, ok := itemMap[subField].(string); ok {
					result = append(result, itemValue)
				}
			}
		}
	}

	return result
}

// GetCollectionForContentType returns the appropriate collection for the content type
func (pc *PineconeController) GetCollectionForContentType(contentType string) *mongo.Collection {
	switch contentType {
	case "movie":
		return pc.Database.Database.Collection("movies")
	case "tvseries":
		return pc.Database.Database.Collection("tv-series")
	case "anime":
		return pc.Database.Database.Collection("animes")
	case "game":
		return pc.Database.Database.Collection("games")
	default:
		return nil
	}
}

func (pc *PineconeController) getContentSimilarityRecommendations(contentID, contentType string, limit int) ([]map[string]interface{}, error) {
	ctx := context.Background()

	// build a MetadataFilter from a simple map
	filterMap := map[string]interface{}{"type": contentType}
	mf, err := structpb.NewStruct(filterMap)
	if err != nil {
		return nil, fmt.Errorf("failed to build metadata filter: %w", err)
	}

	req := &pinecone.QueryByVectorIdRequest{
		VectorId:        contentID,
		TopK:            uint32(limit),
		MetadataFilter:  mf,
		IncludeValues:   false,
		IncludeMetadata: true,
	}

	res, err := pc.PineconeIndex.QueryByVectorId(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("pinecone query-by-vector-id error: %w", err)
	}

	// Format the results
	results := make([]map[string]interface{}, 0, len(res.Matches))
	for _, m := range res.Matches {
		var id string
		if m.Vector != nil {
			id = m.Vector.Id
		}
		rec := map[string]interface{}{
			"id":    id,
			"score": m.Score,
			"type":  contentType,
		}
		// override type if metadata contains it
		if m.Vector != nil && m.Vector.Metadata != nil {
			if val, ok := m.Vector.Metadata.Fields["type"]; ok {
				rec["type"] = val.GetStringValue()
			}
		}
		results = append(results, rec)
	}

	logrus.WithFields(logrus.Fields{
		"contentID":   contentID,
		"contentType": contentType,
		"returned":    len(results),
	}).Info("Pinecone similarity query complete")

	return results, nil
}

// GetSimilarItemsFromPinecone delegates to getContentSimilarityRecommendations
func (pc *PineconeController) GetSimilarItemsFromPinecone(contentID, contentType string, limit int) ([]map[string]interface{}, error) {
	return pc.getContentSimilarityRecommendations(contentID, contentType, limit)
}
