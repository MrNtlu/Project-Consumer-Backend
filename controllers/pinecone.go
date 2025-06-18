package controllers

import (
	"app/db"
	"app/models"
	"app/responses"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/structpb"
)

// Cache TTL constants - optimized for weekly content updates
const (
	EmbeddingCacheTTL = 7 * 24 * time.Hour // 7 days - embeddings rarely change
	SearchCacheTTL    = 2 * time.Hour      // 2 hours - balance freshness with performance
)

// Cache key prefixes
const (
	EmbeddingCachePrefix = "embedding:"
	SearchCachePrefix    = "search:"
)

type PineconeController struct {
	Database      *db.MongoDB
	Pinecone      *pinecone.Client
	PineconeIndex *pinecone.IndexConnection
	RedisClient   *redis.Client
	UserListModel *models.UserListModel
}

func NewPineconeController(
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
	redisClient *redis.Client,
) PineconeController {
	return PineconeController{
		Database:      mongoDB,
		Pinecone:      pinecone,
		PineconeIndex: pineconeIndex,
		RedisClient:   redisClient,
		UserListModel: models.NewUserListModel(mongoDB),
	}
}

// startCacheCleanup runs a background goroutine to clean expired cache entries

// generateCacheKey creates a consistent cache key from components
func (pc *PineconeController) generateCacheKey(prefix string, parts ...string) string {
	return prefix + strings.Join(parts, "|")
}

// GetCacheStats returns cache statistics for monitoring
func (pc *PineconeController) GetCacheStats() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get embedding cache stats
	embeddingKeys, err := pc.RedisClient.Keys(ctx, EmbeddingCachePrefix+"*").Result()
	embeddingCount := 0
	if err == nil {
		embeddingCount = len(embeddingKeys)
	}

	// Get search cache stats
	searchKeys, err := pc.RedisClient.Keys(ctx, SearchCachePrefix+"*").Result()
	searchCount := 0
	if err == nil {
		searchCount = len(searchKeys)
	}

	// Get Redis memory info
	memInfo, err := pc.RedisClient.Info(ctx, "memory").Result()
	memoryUsed := "unknown"
	if err == nil {
		// Parse used_memory from Redis INFO output
		lines := strings.Split(memInfo, "\r\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "used_memory_human:") {
				memoryUsed = strings.TrimPrefix(line, "used_memory_human:")
				break
			}
		}
	}

	return map[string]interface{}{
		"embedding_cache_count": embeddingCount,
		"search_cache_count":    searchCount,
		"embedding_cache_ttl":   EmbeddingCacheTTL.String(),
		"search_cache_ttl":      SearchCacheTTL.String(),
		"redis_memory_used":     memoryUsed,
		"cache_type":            "redis",
	}
}

// ClearCache clears all caches (useful for testing or manual cache invalidation)
func (pc *PineconeController) ClearCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear embedding cache
	embeddingKeys, err := pc.RedisClient.Keys(ctx, EmbeddingCachePrefix+"*").Result()
	if err == nil && len(embeddingKeys) > 0 {
		pc.RedisClient.Del(ctx, embeddingKeys...)
	}

	// Clear search cache
	searchKeys, err := pc.RedisClient.Keys(ctx, SearchCachePrefix+"*").Result()
	if err == nil && len(searchKeys) > 0 {
		pc.RedisClient.Del(ctx, searchKeys...)
	}

	logrus.Info("cache cleared successfully")
}

// GetRecommendationsByType gets recommendations for a user based on their content lists
func (pc *PineconeController) GetRecommendationsByType(userID string, topK int) (map[string]interface{}, error) {
	// 1. Fetch user content lists and not interested list concurrently
	var (
		userLists         responses.UserListAISuggestion
		notInterestedList []models.NotInterested
		errUserLists      error
		errNotInterested  error
	)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		userLists, errUserLists = pc.UserListModel.GetUserListIDForSuggestion(userID)
		if errUserLists != nil {
			logrus.WithFields(logrus.Fields{"userID": userID, "error": errUserLists}).Error("failed to get user list for recommendation")
		}
	}()

	go func() {
		defer wg.Done()
		aiSuggestionsModel := models.NewAISuggestionsModel(pc.Database)
		notInterestedList, errNotInterested = aiSuggestionsModel.GetAllNotInterestedByUserID(userID)
		if errNotInterested != nil {
			logrus.WithFields(logrus.Fields{"userID": userID, "error": errNotInterested}).Error("failed to get not interested content")
		}
	}()
	wg.Wait()

	if errUserLists != nil {
		return nil, fmt.Errorf("failed to get user list for recommendation: %w", errUserLists)
	}
	if errNotInterested != nil {
		return nil, fmt.Errorf("failed to get not interested content: %w", errNotInterested)
	}

	// 2. Create not interested content IDs list for Pinecone filtering
	notInterestedIDs := make([]string, len(notInterestedList))
	for i, notInterested := range notInterestedList {
		notInterestedIDs[i] = notInterested.ContentID
	}

	// 3. Build user content ID sets for duplicate avoidance
	makeSet := func(ids []responses.UserListAISuggestionID) map[string]bool {
		set := make(map[string]bool, len(ids))
		for _, c := range ids {
			set[c.ID] = true
		}
		return set
	}
	movieSet := makeSet(userLists.MovieIDList)
	tvSet := makeSet(userLists.TVIDList)
	animeSet := makeSet(userLists.AnimeIDList)
	gameSet := makeSet(userLists.GameIDList)

	// 4. Fetch ALL content details concurrently (no sampling)
	var (
		movieDetails []bson.M
		tvDetails    []bson.M
		animeDetails []bson.M
		gameDetails  []bson.M
		errMovie     error
		errTV        error
		errAnime     error
		errGame      error
	)
	var wgFetch sync.WaitGroup
	wgFetch.Add(4)

	go func() {
		defer wgFetch.Done()
		movieDetails, errMovie = pc.fetchContentDetails(userLists.MovieIDList, "movies")
		if errMovie != nil {
			logrus.WithError(errMovie).Error("failed to fetch movie details")
		}
	}()
	go func() {
		defer wgFetch.Done()
		tvDetails, errTV = pc.fetchContentDetails(userLists.TVIDList, "tv-series")
		if errTV != nil {
			logrus.WithError(errTV).Error("failed to fetch TV series details")
		}
	}()
	go func() {
		defer wgFetch.Done()
		animeDetails, errAnime = pc.fetchContentDetails(userLists.AnimeIDList, "animes")
		if errAnime != nil {
			logrus.WithError(errAnime).Error("failed to fetch anime details")
		}
	}()
	go func() {
		defer wgFetch.Done()
		gameDetails, errGame = pc.fetchContentDetails(userLists.GameIDList, "games")
		if errGame != nil {
			logrus.WithError(errGame).Error("failed to fetch game details")
		}
	}()
	wgFetch.Wait()

	// 5. Log content counts
	logrus.WithFields(logrus.Fields{
		"movieCount":         len(movieDetails),
		"tvCount":            len(tvDetails),
		"animeCount":         len(animeDetails),
		"gameCount":          len(gameDetails),
		"totalItems":         len(movieDetails) + len(tvDetails) + len(animeDetails) + len(gameDetails),
		"notInterestedCount": len(notInterestedIDs),
	}).Info("content details and not interested list retrieved for recommendations")

	// 6. Generate recommendations concurrently using ALL content and Pinecone-level filtering
	recs := make(map[string][]map[string]interface{})
	var wgRecs sync.WaitGroup
	wgRecs.Add(4)

	go func() {
		defer wgRecs.Done()
		recs["movies"] = pc.getAccurateRecommendationsWithPineconeFiltering(movieDetails, "movie", 10, movieSet, notInterestedIDs)
	}()
	go func() {
		defer wgRecs.Done()
		recs["tvSeries"] = pc.getAccurateRecommendationsWithPineconeFiltering(tvDetails, "tvseries", 10, tvSet, notInterestedIDs)
	}()
	go func() {
		defer wgRecs.Done()
		recs["animes"] = pc.getAccurateRecommendationsWithPineconeFiltering(animeDetails, "anime", 10, animeSet, notInterestedIDs)
	}()
	go func() {
		defer wgRecs.Done()
		recs["games"] = pc.getAccurateRecommendationsWithPineconeFiltering(gameDetails, "game", 10, gameSet, notInterestedIDs)
	}()
	wgRecs.Wait()

	// 7. Combine all recommendations
	all := make([]map[string]interface{}, 0,
		len(recs["movies"])+len(recs["tvSeries"])+len(recs["animes"])+len(recs["games"]))
	for _, key := range []string{"movies", "tvSeries", "animes", "games"} {
		all = append(all, recs[key]...)
	}

	// 8. Return structured results
	return map[string]interface{}{
		"movies":   recs["movies"],
		"tvSeries": recs["tvSeries"],
		"animes":   recs["animes"],
		"games":    recs["games"],
		"all":      all,
	}, nil
}

// getRecommendationsForContentTypeWithHybrid gets recommendations with hybrid scoring
func (pc *PineconeController) getRecommendationsForContentTypeWithHybrid(
	contentData []bson.M,
	contentType string,
	limit int,
	userContentSet map[string]bool,
	notInterestedSet map[string]bool,
) []map[string]interface{} {
	if len(contentData) == 0 {
		return nil
	}

	// Get a sample of the user's content (up to 3 items) for generating recommendations
	sampleSize := 3
	if len(contentData) < sampleSize {
		sampleSize = len(contentData)
	}

	// Channels and sync for concurrent fetch
	itemCh := make(chan map[string]interface{}, sampleSize*limit*2) // Request more for hybrid filtering
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(sampleSize)

	// Concurrently fetch similar items for each sample with hybrid scoring
	for i := 0; i < sampleSize; i++ {
		go func(idx int) {
			defer wg.Done()
			// Extract contentID
			var contentID string
			if id, ok := contentData[idx]["_id"].(primitive.ObjectID); ok {
				contentID = id.Hex()
			} else if id, ok := contentData[idx]["_id"].(string); ok {
				contentID = id
			}
			if contentID == "" {
				return
			}

			// Fetch similar items with hybrid scoring
			sims, err := pc.GetSimilarItemsWithHybridScoring(contentID, contentType, contentData[idx], (limit/sampleSize)*3)
			if err != nil {
				logrus.WithFields(logrus.Fields{"contentID": contentID, "error": err}).Error("failed to get hybrid similar from Pinecone")
				return
			}
			// Send items to channel or exit if done
			for _, item := range sims {
				select {
				case itemCh <- item:
				case <-done:
					return
				}
			}
		}(i)
	}
	// Close channel when fetchers finish
	go func() {
		wg.Wait()
		close(itemCh)
	}()

	// Collect and filter recommendations
	var (
		recommendations []map[string]interface{}
		seen            = make(map[string]struct{}, limit)
	)
Loop:
	for item := range itemCh {
		// Extract ID
		idVal, ok := item["id"].(string)
		if !ok || idVal == "" || userContentSet[idVal] || notInterestedSet[idVal] {
			continue
		}
		// Dedupe
		if _, exists := seen[idVal]; exists {
			continue
		}
		seen[idVal] = struct{}{}
		// Assign type if missing
		if item["type"] == nil {
			item["type"] = contentType
		}
		recommendations = append(recommendations, item)
		// Stop when limit reached
		if len(recommendations) >= limit {
			close(done)
			break Loop
		}
	}

	// Sort by hybrid score (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		scoreI, okI := recommendations[i]["hybrid_score"].(float64)
		scoreJ, okJ := recommendations[j]["hybrid_score"].(float64)
		if !okI {
			scoreI = 0
		}
		if !okJ {
			scoreJ = 0
		}
		return scoreI > scoreJ
	})

	return recommendations
}

// GetSimilarItemsWithHybridScoring combines vector similarity with metadata similarity
func (pc *PineconeController) GetSimilarItemsWithHybridScoring(contentID, contentType string, sourceContent bson.M, limit int) ([]map[string]interface{}, error) {
	// Get vector similarity results from Pinecone (request more than needed for hybrid filtering)
	vectorResults, err := pc.getContentSimilarityRecommendations(contentID, contentType, limit*2)
	if err != nil {
		return nil, err
	}

	// Process results with hybrid scoring
	var enhancedResults []map[string]interface{}
	for _, result := range vectorResults {
		resultID, ok := result["id"].(string)
		if !ok || resultID == "" {
			continue
		}

		// Fetch full content data for metadata comparison
		resultContent, err := pc.getContentByID(resultID, contentType)
		if err != nil {
			logrus.WithError(err).Warnf("failed to fetch content for ID: %s", resultID)
			// Skip this result but continue with others
			continue
		}

		// Calculate metadata similarity
		metadataSimilarity := pc.CalculateMetadataSimilarity(sourceContent, resultContent, contentType)

		// Combine vector similarity with metadata similarity
		vectorScore, ok := result["score"].(float32)
		if !ok {
			vectorScore = 0
		}

		// Hybrid scoring: 70% vector similarity + 30% metadata similarity
		hybridScore := (float64(vectorScore) * 0.7) + (metadataSimilarity * 0.3)

		enhancedResult := map[string]interface{}{
			"id":             resultID,
			"vector_score":   vectorScore,
			"metadata_score": metadataSimilarity,
			"hybrid_score":   hybridScore,
			"type":           contentType,
			"data":           resultContent,
		}
		enhancedResults = append(enhancedResults, enhancedResult)
	}

	// Sort by hybrid score (highest first)
	sort.Slice(enhancedResults, func(i, j int) bool {
		return enhancedResults[i]["hybrid_score"].(float64) > enhancedResults[j]["hybrid_score"].(float64)
	})

	// Return top results
	if len(enhancedResults) > limit {
		enhancedResults = enhancedResults[:limit]
	}

	logrus.WithFields(logrus.Fields{
		"contentID":     contentID,
		"contentType":   contentType,
		"vectorResults": len(vectorResults),
		"hybridResults": len(enhancedResults),
	}).Info("hybrid scoring complete")

	return enhancedResults, nil
}

// Helper function to get content by ID
func (pc *PineconeController) getContentByID(contentID, contentType string) (bson.M, error) {
	collection := pc.GetCollectionForContentType(contentType)
	if collection == nil {
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	var query bson.M
	if objectID, err := primitive.ObjectIDFromHex(contentID); err == nil {
		query = bson.M{"_id": objectID}
	} else {
		query = bson.M{"_id": contentID}
	}

	var result bson.M
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find content with ID %s: %w", contentID, err)
	}

	return result, nil
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

// getRecommendationsForContentType gets recommendations for a specific content type (original method kept for fallback)
func (pc *PineconeController) getRecommendationsForContentType(contentData []bson.M, contentType string, limit int, userContentSet map[string]bool, notInterestedSet map[string]bool) []map[string]interface{} {
	if len(contentData) == 0 {
		return nil
	}

	// Get a sample of the user's content (up to 3 items) for generating recommendations
	sampleSize := 3
	if len(contentData) < sampleSize {
		sampleSize = len(contentData)
	}

	// Channels and sync for concurrent fetch
	itemCh := make(chan map[string]interface{}, sampleSize*limit)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(sampleSize)

	// Concurrently fetch similar items for each sample
	for i := 0; i < sampleSize; i++ {
		go func(idx int) {
			defer wg.Done()
			// Extract contentID
			var contentID string
			if id, ok := contentData[idx]["_id"].(primitive.ObjectID); ok {
				contentID = id.Hex()
			} else if id, ok := contentData[idx]["_id"].(string); ok {
				contentID = id
			}
			if contentID == "" {
				return
			}

			// Fetch similar items
			sims, err := pc.GetSimilarItemsFromPinecone(contentID, contentType, (limit/sampleSize)*2)
			if err != nil {
				logrus.WithFields(logrus.Fields{"contentID": contentID, "error": err}).Error("failed to get similar from Pinecone")
				return
			}
			// Send items to channel or exit if done
			for _, item := range sims {
				select {
				case itemCh <- item:
				case <-done:
					return
				}
			}
		}(i)
	}
	// Close channel when fetchers finish
	go func() {
		wg.Wait()
		close(itemCh)
	}()

	// Collect and filter recommendations
	var (
		recommendations []map[string]interface{}
		seen            = make(map[string]struct{}, limit)
	)
Loop:
	for item := range itemCh {
		// Extract ID
		idVal, ok := item["id"].(string)
		if !ok || idVal == "" || userContentSet[idVal] || notInterestedSet[idVal] {
			continue
		}
		// Dedupe
		if _, exists := seen[idVal]; exists {
			continue
		}
		seen[idVal] = struct{}{}
		// Assign type if missing
		if item["type"] == nil {
			item["type"] = contentType
		}
		recommendations = append(recommendations, item)
		// Stop when limit reached
		if len(recommendations) >= limit {
			close(done)
			break Loop
		}
	}

	return recommendations
}

// GetRecommendationsForType gets recommendations for a specific content type
func (pc *PineconeController) GetRecommendationsForType(contentData []bson.M, contentType string, perTypeCount int, userContentSet map[string]bool, notInterestedSet map[string]bool) ([]map[string]interface{}, error) {
	if len(contentData) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Step 1: Find sequels and series first
	sequelRecs, err := pc.FindSequelsAndSeries(contentData, contentType, userContentSet, notInterestedSet)
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

			if itemID == "" || userContentSet[itemID] || notInterestedSet[itemID] {
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
func (pc *PineconeController) FindSequelsAndSeries(contentData []bson.M, contentType string, userContentSet map[string]bool, notInterestedSet map[string]bool) ([]map[string]interface{}, error) {
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

		opts := options.Find().SetLimit(3)
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

			if contentIDStr != "" && !userContentSet[contentIDStr] && !notInterestedSet[contentIDStr] {
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
func (pc *PineconeController) CalculateMetadataSimilarity(
	content1, content2 bson.M,
	contentType string,
) float64 {
	score := 0.5 // base similarity

	switch contentType {
	case "anime":
		genres1 := pc.GetFieldArrayValues(content1, "genres", "name")
		genres2 := pc.GetFieldArrayValues(content2, "genres", "name")
		score += pc.CompareArrays(genres1, genres2) * 0.2

		demo1 := pc.GetFieldArrayValues(content1, "demographics", "name")
		demo2 := pc.GetFieldArrayValues(content2, "demographics", "name")
		score += pc.CompareArrays(demo1, demo2) * 0.2

		themes1 := pc.GetFieldArrayValues(content1, "themes", "name")
		themes2 := pc.GetFieldArrayValues(content2, "themes", "name")
		score += pc.CompareArrays(themes1, themes2) * 0.05

		studios1 := pc.GetFieldArrayValues(content1, "studios", "name")
		studios2 := pc.GetFieldArrayValues(content2, "studios", "name")
		score += pc.CompareArrays(studios1, studios2) * 0.05

	case "movie":
		// genres
		var g1, g2 []string
		if arr, ok := content1["genres"].(primitive.A); ok {
			for _, v := range arr {
				if s, ok2 := v.(string); ok2 {
					g1 = append(g1, s)
				}
			}
		}
		if arr, ok := content2["genres"].(primitive.A); ok {
			for _, v := range arr {
				if s, ok2 := v.(string); ok2 {
					g2 = append(g2, s)
				}
			}
		}
		score += pc.CompareArrays(g1, g2) * 0.25

		pcList1 := pc.GetFieldArrayValues(content1, "production_companies", "name")
		pcList2 := pc.GetFieldArrayValues(content2, "production_companies", "name")
		score += pc.CompareArrays(pcList1, pcList2) * 0.1

		actors1 := pc.GetFieldArrayValues(content1, "actors", "name")
		actors2 := pc.GetFieldArrayValues(content2, "actors", "name")
		if len(actors1) > 3 {
			actors1 = actors1[:3]
		}
		if len(actors2) > 3 {
			actors2 = actors2[:3]
		}
		score += pc.CompareArrays(actors1, actors2) * 0.15

	case "tvseries":
		var tg1, tg2 []string
		if arr, ok := content1["genres"].(primitive.A); ok {
			for _, v := range arr {
				if s, ok2 := v.(string); ok2 {
					tg1 = append(tg1, s)
				}
			}
		}
		if arr, ok := content2["genres"].(primitive.A); ok {
			for _, v := range arr {
				if s, ok2 := v.(string); ok2 {
					tg2 = append(tg2, s)
				}
			}
		}
		score += pc.CompareArrays(tg1, tg2) * 0.25

		nets1 := pc.GetFieldArrayValues(content1, "networks", "name")
		nets2 := pc.GetFieldArrayValues(content2, "networks", "name")
		score += pc.CompareArrays(nets1, nets2) * 0.1

	case "game":
		var gg1, gg2 []string
		if arr, ok := content1["genres"].(primitive.A); ok {
			for _, v := range arr {
				if s, ok2 := v.(string); ok2 {
					gg1 = append(gg1, s)
				}
			}
		}
		if arr, ok := content2["genres"].(primitive.A); ok {
			for _, v := range arr {
				if s, ok2 := v.(string); ok2 {
					gg2 = append(gg2, s)
				}
			}
		}
		score += pc.CompareArrays(gg1, gg2) * 0.25

		pls1 := pc.GetFieldArrayValues(content1, "platforms", "")
		pls2 := pc.GetFieldArrayValues(content2, "platforms", "")
		score += pc.CompareArrays(pls1, pls2) * 0.2

		dev1 := pc.GetFieldArrayValues(content1, "developers", "")
		dev2 := pc.GetFieldArrayValues(content2, "developers", "")
		score += pc.CompareArrays(dev1, dev2) * 0.10
	}

	// incorporate popularity if present in metadata
	if pop, ok := content2["score"].(float64); ok {
		score += pop * 0.1
	}

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

// getAccurateRecommendationsWithPineconeFiltering gets recommendations using ALL user content with aggressive parallel processing
func (pc *PineconeController) getAccurateRecommendationsWithPineconeFiltering(
	contentData []bson.M,
	contentType string,
	limit int,
	userContentSet map[string]bool,
	notInterestedIDs []string,
) []map[string]interface{} {
	if len(contentData) == 0 {
		return []map[string]interface{}{}
	}

	logrus.WithFields(logrus.Fields{
		"contentType":        contentType,
		"userContentCount":   len(contentData),
		"notInterestedCount": len(notInterestedIDs),
		"limit":              limit,
	}).Info("starting aggressive parallel processing with ALL user content")

	// Use ALL user content for maximum accuracy - no sampling!
	// Aggressive concurrency settings based on Pinecone best practices
	maxConcurrency := 25 // Pinecone supports up to 30+ concurrent requests
	if len(contentData) < maxConcurrency {
		maxConcurrency = len(contentData)
	}

	// Channels for parallel processing with larger buffer
	type recommendationResult struct {
		recommendations []map[string]interface{}
		err             error
	}

	resultChan := make(chan recommendationResult, len(contentData))
	semaphore := make(chan struct{}, maxConcurrency) // Semaphore to control concurrency
	var wg sync.WaitGroup

	// Launch parallel Pinecone queries for ALL content with controlled concurrency
	for _, content := range contentData {
		wg.Add(1)
		go func(content bson.M) {
			defer wg.Done()

			// Acquire semaphore - this controls max concurrent API calls
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Extract contentID
			var contentID string
			if id, ok := content["_id"].(primitive.ObjectID); ok {
				contentID = id.Hex()
			} else if id, ok := content["_id"].(string); ok {
				contentID = id
			}

			if contentID == "" {
				resultChan <- recommendationResult{recommendations: nil, err: nil}
				return
			}

			// Request more recommendations per query to ensure we have enough candidates
			// This compensates for the filtering that happens later
			queryLimit := limit * 2 // Request 2x more per query for better results
			if queryLimit > 100 {   // Don't go crazy with API limits
				queryLimit = 100
			}

			recommendations, err := pc.getContentSimilarityRecommendationsWithFilter(
				contentID,
				contentType,
				queryLimit,
				notInterestedIDs,
			)

			resultChan <- recommendationResult{recommendations: recommendations, err: err}
		}(content)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all recommendations from parallel queries with smart deduplication
	allRecommendations := make([]map[string]interface{}, 0, len(contentData)*limit)
	seen := make(map[string]struct{})
	successfulQueries := 0
	failedQueries := 0

	for result := range resultChan {
		if result.err != nil {
			failedQueries++
			logrus.WithError(result.err).Error("failed to get filtered recommendations from Pinecone")
			continue
		}
		successfulQueries++

		// Add recommendations, avoiding duplicates and user's existing content
		for _, rec := range result.recommendations {
			itemID, ok := rec["id"].(string)
			if !ok || itemID == "" {
				continue
			}

			// Skip if user already has this content
			if userContentSet[itemID] {
				continue
			}

			// Skip if already added (deduplication)
			if _, exists := seen[itemID]; exists {
				continue
			}

			seen[itemID] = struct{}{}
			rec["type"] = contentType
			allRecommendations = append(allRecommendations, rec)
		}
	}

	// Sort by score (highest first) - this ensures we return the best recommendations
	sort.Slice(allRecommendations, func(i, j int) bool {
		scoreI, okI := allRecommendations[i]["score"].(float32)
		scoreJ, okJ := allRecommendations[j]["score"].(float32)
		if !okI {
			scoreI = 0
		}
		if !okJ {
			scoreJ = 0
		}
		return scoreI > scoreJ
	})

	// Return top results
	if len(allRecommendations) > limit {
		allRecommendations = allRecommendations[:limit]
	}

	logrus.WithFields(logrus.Fields{
		"contentType":       contentType,
		"userContentUsed":   len(contentData),
		"concurrentQueries": maxConcurrency,
		"successfulQueries": successfulQueries,
		"failedQueries":     failedQueries,
		"totalCandidates":   len(allRecommendations),
		"requestedLimit":    limit,
		"actualReturned":    len(allRecommendations),
		"deduplicatedItems": len(seen),
	}).Info("aggressive parallel processing with ALL content complete")

	return allRecommendations
}

// getContentSimilarityRecommendationsWithFilter gets similar content with not interested filtering at Pinecone level
func (pc *PineconeController) getContentSimilarityRecommendationsWithFilter(contentID, contentType string, limit int, notInterestedIDs []string) ([]map[string]interface{}, error) {
	ctx := context.Background()

	// Build metadata filter for content type only
	// Note: Pinecone doesn't support complex operators like $nin in metadata filters
	// We'll filter out not interested content after getting results from Pinecone
	filterMap := map[string]interface{}{
		"type": contentType,
	}

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

	// Create a set for fast lookup of not interested IDs
	notInterestedSet := make(map[string]bool)
	for _, id := range notInterestedIDs {
		notInterestedSet[id] = true
	}

	// Format the results and filter out not interested content
	results := make([]map[string]interface{}, 0, len(res.Matches))
	filteredCount := 0
	for _, m := range res.Matches {
		var id string
		if m.Vector != nil {
			id = m.Vector.Id
		}

		// Skip if this content is in the not interested list
		if len(notInterestedIDs) > 0 && notInterestedSet[id] {
			filteredCount++
			continue
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
		"contentID":          contentID,
		"contentType":        contentType,
		"notInterestedCount": len(notInterestedIDs),
		"filteredOut":        filteredCount,
		"returned":           len(results),
		"filterApplied":      len(notInterestedIDs) > 0,
	}).Info("Pinecone similarity query with filtering complete")

	return results, nil
}

// SearchContentByQuery performs semantic search using Pinecone and returns paginated results
func (pc *PineconeController) SearchContentByQuery(query string, contentType string, page int, limit int) ([]map[string]interface{}, int, error) {
	// Check search cache first
	if results, totalResults, found := pc.getSearchFromCache(query, contentType, page, limit); found {
		logrus.WithFields(logrus.Fields{
			"query":       query,
			"contentType": contentType,
			"page":        page,
			"limit":       limit,
		}).Debug("search cache hit")
		return results, totalResults, nil
	}

	// Cache miss - perform search
	logrus.WithFields(logrus.Fields{
		"query":       query,
		"contentType": contentType,
		"page":        page,
		"limit":       limit,
	}).Debug("search cache miss - performing full search")

	// 1. Preprocess query to handle common typos and variations
	processedQuery := pc.preprocessSearchQuery(query)

	// 2. Get embedding for the search query (with caching)
	embedding, err := pc.getQueryEmbedding(processedQuery)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"query": query,
			"error": err,
		}).Error("failed to get query embedding")
		return nil, 0, fmt.Errorf("failed to generate embedding for query: %w", err)
	}

	// 3. Calculate pagination parameters
	offset := (page - 1) * limit
	searchLimit := limit + 10 // Get slightly more results for better ranking

	// 4. Search Pinecone with the embedding
	pineconeResults, err := pc.searchPineconeByVector(embedding, contentType, searchLimit)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"query":       query,
			"contentType": contentType,
			"error":       err,
		}).Error("failed to search Pinecone")
		return nil, 0, fmt.Errorf("failed to search Pinecone: %w", err)
	}

	// 5. Apply hybrid scoring and get MongoDB data
	enrichedResults, err := pc.enrichWithMongoDBData(pineconeResults, query, contentType)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"query":       query,
			"contentType": contentType,
			"error":       err,
		}).Error("failed to enrich with MongoDB data")
		return nil, 0, fmt.Errorf("failed to enrich results: %w", err)
	}

	// 6. Sort by hybrid score
	sort.Slice(enrichedResults, func(i, j int) bool {
		scoreI, okI := enrichedResults[i]["hybrid_score"].(float64)
		scoreJ, okJ := enrichedResults[j]["hybrid_score"].(float64)
		if !okI {
			scoreI = 0
		}
		if !okJ {
			scoreJ = 0
		}
		return scoreI > scoreJ
	})

	// 7. Apply pagination
	totalResults := len(enrichedResults)
	start := offset
	end := offset + limit

	if start >= totalResults {
		return []map[string]interface{}{}, totalResults, nil
	}

	if end > totalResults {
		end = totalResults
	}

	paginatedResults := enrichedResults[start:end]

	// Cache the search results
	go pc.setSearchCache(query, contentType, page, limit, paginatedResults, totalResults)

	logrus.WithFields(logrus.Fields{
		"query":           query,
		"contentType":     contentType,
		"page":            page,
		"limit":           limit,
		"totalResults":    totalResults,
		"returnedResults": len(paginatedResults),
		"cached":          true,
	}).Info("search completed successfully and cached")

	return paginatedResults, totalResults, nil
}

// preprocessSearchQuery handles common typos and query variations with optimized lookup
func (pc *PineconeController) preprocessSearchQuery(query string) string {
	// Early return for empty queries
	if query == "" {
		return query
	}

	// Comprehensive typo corrections for entertainment content (optimized map)
	typoMap := map[string]string{
		// Pirates of the Caribbean variations
		"carribiean": "caribbean",
		"carribbean": "caribbean",
		"carribean":  "caribbean",
		"caribean":   "caribbean",
		"caribian":   "caribbean",
		"caribbian":  "caribbean",

		// Popular franchises
		"batman":               "dark knight batman",
		"spiderman":            "spider-man",
		"superman":             "man of steel superman",
		"xmen":                 "x-men",
		"starwars":             "star wars",
		"star wars":            "star wars jedi sith force",
		"lordoftherings":       "lord of the rings",
		"harrypotter":          "harry potter",
		"breakingbad":          "breaking bad",
		"walkingdead":          "walking dead",
		"fastfurious":          "fast and furious",
		"jurassicpark":         "jurassic park",
		"transformers":         "transformers autobots",
		"missionimpossible":    "mission impossible",
		"johnwick":             "john wick",
		"avengers":             "avengers marvel",
		"ironman":              "iron man",
		"captainamerica":       "captain america",
		"wonderwoman":          "wonder woman",
		"blackpanther":         "black panther",
		"guardiansofthegalaxy": "guardians of the galaxy",
		"sherlockholmes":       "sherlock holmes",

		// Anime specific
		"dragonball":         "dragon ball",
		"dragonballz":        "dragon ball z",
		"onepiece":           "one piece",
		"naruto":             "naruto",
		"bleach":             "bleach",
		"attackontitan":      "attack on titan",
		"deathnote":          "death note",
		"fullmetalalchemist": "fullmetal alchemist",
		"onepunchman":        "one punch man",
		"myheroacademia":     "my hero academia",
		"demonslayer":        "demon slayer",
		"jujutsukaisen":      "jujutsu kaisen",
		"tokyoghoul":         "tokyo ghoul",
		"mobpsycho":          "mob psycho",
		"hunterxhunter":      "hunter x hunter",
		"cowboybebop":        "cowboy bebop",
		"evangelion":         "evangelion",
		"studioghibli":       "studio ghibli",
		"pokemon":            "pokemon",
		"sailormoon":         "sailor moon",

		// Gaming
		"callofduty":        "call of duty",
		"grandtheftauto":    "grand theft auto",
		"assassinscreed":    "assassins creed",
		"finalfantasy":      "final fantasy",
		"worldofwarcraft":   "world of warcraft",
		"elderscrolls":      "elder scrolls",
		"godofwar":          "god of war",
		"lastofus":          "last of us",
		"uncharted":         "uncharted",
		"residentevil":      "resident evil",
		"silenthill":        "silent hill",
		"metalgear":         "metal gear",
		"streetfighter":     "street fighter",
		"mortalkombat":      "mortal kombat",
		"supersmashbros":    "super smash bros",
		"legendofzelda":     "legend of zelda",
		"supermario":        "super mario",
		"minecraft":         "minecraft",
		"fortnite":          "fortnite",
		"amongus":           "among us",
		"fallout":           "fallout",
		"skyrim":            "skyrim",
		"witcher":           "witcher",
		"cyberpunk":         "cyberpunk",
		"reddeadredemption": "red dead redemption",
		"gtav":              "grand theft auto v",
		"gta5":              "grand theft auto v",
		"cod":               "call of duty",
		"wow":               "world of warcraft",
		"lol":               "league of legends",
		"dota":              "dota",
		"csgo":              "counter strike",
		"valorant":          "valorant",
		"overwatch":         "overwatch",
		"apex":              "apex legends",
		"pubg":              "pubg",

		// TV Shows
		"friendstv":           "friends",
		"theofficeus":         "the office",
		"gameofthrones":       "game of thrones westeros",
		"strangerthings":      "stranger things",
		"thewitcher":          "the witcher",
		"mandalorian":         "mandalorian",
		"houseofcards":        "house of cards",
		"orangeisthenewblack": "orange is the new black",
		"13reasonswhy":        "13 reasons why",
		"blackmirror":         "black mirror",
		"thecrownnetflix":     "the crown",
		"moneyheist":          "money heist",
		"squidgame":           "squid game",
		"bridgerton":          "bridgerton",
		"theumbrellaacademy":  "umbrella academy",
		"lockeandkey":         "locke and key",
		"thecrown":            "the crown",
		"peakyblinders":       "peaky blinders",
		"sherlockbbc":         "sherlock",
		"doctorwho":           "doctor who time travel",
		"westworld":           "westworld",
		"houseofthedragon":    "house of the dragon",
		"ringsofpower":        "rings of power",

		// Common misspellings
		"marvell":     "marvel",
		"dc comics":   "dc",
		"disneyplus":  "disney",
		"netflix":     "netflix",
		"hbo":         "hbo",
		"amazon":      "amazon prime",
		"hulu":        "hulu",
		"appletv":     "apple tv",
		"paramount":   "paramount",
		"peacock":     "peacock",
		"crunchyroll": "crunchyroll",
		"funimation":  "funimation",

		// Genre expansions
		"horror":      "horror scary thriller suspense",
		"comedy":      "comedy funny humor",
		"action":      "action adventure",
		"romance":     "romance love",
		"scifi":       "science fiction",
		"sci-fi":      "science fiction",
		"fantasy":     "fantasy magic",
		"thriller":    "thriller suspense",
		"drama":       "drama",
		"documentary": "documentary",
		"animation":   "animation animated",
		"superhero":   "superhero marvel dc",
		"zombie":      "zombie undead",
		"vampire":     "vampire",
		"werewolf":    "werewolf",
		"alien":       "alien space",
		"robot":       "robot ai artificial intelligence",
		"war":         "war military",
		"crime":       "crime detective",
		"mystery":     "mystery detective",
		"western":     "western cowboy",
		"musical":     "musical music",
		"sports":      "sports",
		"family":      "family kids children",
		"kids":        "kids children family",
		"teen":        "teen teenager young adult",
		"adult":       "adult mature",

		// Actor/Director name corrections
		"tom cruise":         "tom cruise",
		"leonardo dicaprio":  "leonardo dicaprio",
		"brad pitt":          "brad pitt",
		"angelina jolie":     "angelina jolie",
		"will smith":         "will smith",
		"johnny depp":        "johnny depp",
		"robert downey":      "robert downey jr",
		"chris evans":        "chris evans",
		"chris hemsworth":    "chris hemsworth",
		"scarlett johansson": "scarlett johansson",
		"jennifer lawrence":  "jennifer lawrence",
		"emma stone":         "emma stone",
		"ryan reynolds":      "ryan reynolds",
		"dwayne johnson":     "dwayne johnson rock",
		"kevin hart":         "kevin hart",
		"samuel jackson":     "samuel l jackson",
		"morgan freeman":     "morgan freeman",
		"denzel washington":  "denzel washington",
		"christopher nolan":  "christopher nolan",
		"quentin tarantino":  "quentin tarantino",
		"martin scorsese":    "martin scorsese",
		"steven spielberg":   "steven spielberg",
		"james cameron":      "james cameron",
		"ridley scott":       "ridley scott",
		"tim burton":         "tim burton",
		"jj abrams":          "j j abrams",
		"michael bay":        "michael bay",
		"zack snyder":        "zack snyder",
		"russo brothers":     "russo brothers",
		"kevin feige":        "kevin feige marvel",
	}

	originalQuery := query
	query = strings.ToLower(strings.TrimSpace(query))

	// Apply typo corrections
	for typo, correction := range typoMap {
		if strings.Contains(query, typo) {
			query = strings.ReplaceAll(query, typo, correction)
		}
	}

	// Handle partial matches and word boundaries
	queryWords := strings.Fields(query)
	for i, word := range queryWords {
		for typo, correction := range typoMap {
			if strings.Contains(word, typo) {
				queryWords[i] = strings.ReplaceAll(word, typo, correction)
			}
		}
	}
	query = strings.Join(queryWords, " ")

	// Smart query expansion based on context
	query = pc.expandSearchQuery(query, originalQuery)

	// Remove extra spaces and normalize
	query = strings.Join(strings.Fields(query), " ")

	// Log the query transformation for debugging
	if query != strings.ToLower(originalQuery) {
		logrus.WithFields(logrus.Fields{
			"original":  originalQuery,
			"processed": query,
		}).Info("query preprocessing applied")
	}

	return query
}

// expandSearchQuery adds contextual terms to improve search quality
func (pc *PineconeController) expandSearchQuery(query, originalQuery string) string {
	// Focused contextual expansions for entertainment content
	contextualExpansions := map[string][]string{
		"batman":            {"dark knight", "gotham", "bruce wayne"},
		"superman":          {"man of steel", "clark kent", "krypton"},
		"spider-man":        {"peter parker", "marvel", "web slinger"},
		"avengers":          {"marvel", "iron man", "captain america"},
		"star wars":         {"jedi", "sith", "force"},
		"lord of the rings": {"tolkien", "gandalf", "frodo"},
		"harry potter":      {"hogwarts", "wizard", "magic"},
		"game of thrones":   {"westeros", "jon snow", "daenerys"},
		"breaking bad":      {"walter white", "jesse pinkman", "heisenberg"},
		"the office":        {"michael scott", "jim halpert", "dwight schrute"},
		"friends":           {"rachel", "monica", "phoebe"},
		"marvel":            {"superhero", "mcu", "comic"},
		"dc":                {"superhero", "comic", "justice league"},
		"disney":            {"animation", "family", "pixar"},
		"anime":             {"japanese", "manga", "animation"},
		"pokemon":           {"pikachu", "nintendo", "anime"},
		"final fantasy":     {"square enix", "rpg", "jrpg"},
		"call of duty":      {"fps", "shooter", "war"},
		"grand theft auto":  {"gta", "rockstar", "open world"},
		"the witcher":       {"geralt", "fantasy", "monster hunter"},
		"horror":            {"scary", "thriller", "suspense"},
		"comedy":            {"funny", "humor", "laugh"},
		"romance":           {"love", "relationship", "romantic"},
		"action":            {"adventure", "fight", "explosive"},
		"sci-fi":            {"science fiction", "futuristic", "space"},
		"fantasy":           {"magic", "medieval", "mythical"},
		"superhero":         {"powers", "cape", "villain"},
		"pirate":            {"ship", "treasure", "sea"},
		"zombie":            {"undead", "apocalypse", "survival"},
		"vampire":           {"blood", "immortal", "gothic"},
		"alien":             {"extraterrestrial", "space", "ufo"},
		"robot":             {"artificial intelligence", "ai", "android"},
	}

	queryLower := strings.ToLower(query)

	// Find matching expansions
	for term, expansions := range contextualExpansions {
		if strings.Contains(queryLower, term) {
			// Add relevant expansions (limit to 2 to avoid query bloat)
			for i, expansion := range expansions {
				if i >= 2 { // Limit to 2 expansions per term
					break
				}
				if !strings.Contains(queryLower, expansion) {
					query += " " + expansion
				}
			}
		}
	}

	return query
}

// getQueryEmbedding gets embedding with Redis caching
func (pc *PineconeController) getQueryEmbedding(query string) ([]float32, error) {
	// Try to get from cache first
	if embedding, found := pc.getEmbeddingFromCache(query); found {
		logrus.WithField("query", query).Debug("embedding cache hit")
		return embedding, nil
	}

	logrus.WithField("query", query).Debug("embedding cache miss, calling OpenAI")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create request body
	requestBody := fmt.Sprintf(`{
		"input": "%s",
		"model": "text-embedding-3-small"
	}`, strings.ReplaceAll(query, `"`, `\"`))

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	embedding := response.Data[0].Embedding

	// Cache the embedding
	go pc.setEmbeddingCache(query, embedding)

	return embedding, nil
}

// searchPineconeByVector searches Pinecone using the provided embedding vector
func (pc *PineconeController) searchPineconeByVector(embedding []float32, contentType string, limit int) ([]map[string]interface{}, error) {
	ctx := context.Background()

	// Build metadata filter for content type
	filterMap := map[string]interface{}{"type": contentType}
	mf, err := structpb.NewStruct(filterMap)
	if err != nil {
		return nil, fmt.Errorf("failed to build metadata filter: %w", err)
	}

	req := &pinecone.QueryByVectorValuesRequest{
		Vector:          embedding,
		TopK:            uint32(limit),
		MetadataFilter:  mf,
		IncludeValues:   false,
		IncludeMetadata: true,
	}

	res, err := pc.PineconeIndex.QueryByVectorValues(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("pinecone query error: %w", err)
	}

	// Format the results
	results := make([]map[string]interface{}, 0, len(res.Matches))
	for _, m := range res.Matches {
		var id string
		if m.Vector != nil {
			id = m.Vector.Id
		}
		result := map[string]interface{}{
			"id":           id,
			"vector_score": m.Score,
			"type":         contentType,
		}

		// Override type if metadata contains it
		if m.Vector != nil && m.Vector.Metadata != nil {
			if val, ok := m.Vector.Metadata.Fields["type"]; ok {
				result["type"] = val.GetStringValue()
			}
		}

		results = append(results, result)
	}

	logrus.WithFields(logrus.Fields{
		"contentType": contentType,
		"returned":    len(results),
		"requested":   limit,
	}).Info("Pinecone search completed")

	return results, nil
}

// enrichWithMongoDBData fetches full content data from MongoDB and applies hybrid scoring
func (pc *PineconeController) enrichWithMongoDBData(pineconeResults []map[string]interface{}, originalQuery string, contentType string) ([]map[string]interface{}, error) {
	if len(pineconeResults) == 0 {
		return []map[string]interface{}{}, nil
	}

	collection := pc.GetCollectionForContentType(contentType)
	if collection == nil {
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	// Extract all content IDs for batch processing
	contentIDs := make([]primitive.ObjectID, 0, len(pineconeResults))
	idToResultMap := make(map[string]map[string]interface{})

	for _, result := range pineconeResults {
		contentID, ok := result["id"].(string)
		if !ok || contentID == "" {
			continue
		}

		// Try to convert to ObjectID
		if objectID, err := primitive.ObjectIDFromHex(contentID); err == nil {
			contentIDs = append(contentIDs, objectID)
			idToResultMap[contentID] = result
		}
	}

	if len(contentIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Batch fetch all content data from MongoDB with projection for only needed fields
	ctx := context.Background()
	projection := bson.M{
		"_id":              1,
		"title_en":         1,
		"title_original":   1,
		"title":            1,
		"title_jp":         1,
		"image_url":        1,
		"description":      1,
		"tmdb_id":          1,
		"tmdb_vote":        1,
		"tmdb_vote_count":  1,
		"mal_id":           1,
		"mal_score":        1,
		"mal_scored_by":    1,
		"rawg_id":          1,
		"rawg_rating":      1,
		"metacritic_score": 1,
		"release_date":     1,
		"first_air_date":   1,
		"total_seasons":    1,
		"episodes":         1,
	}

	cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": contentIDs}}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("failed to batch fetch content: %w", err)
	}
	defer cursor.Close(ctx)

	var enrichedResults []map[string]interface{}

	// Process results
	for cursor.Next(ctx) {
		var contentData bson.M
		if err := cursor.Decode(&contentData); err != nil {
			continue
		}

		// Get content ID
		var contentID string
		if id, ok := contentData["_id"].(primitive.ObjectID); ok {
			contentID = id.Hex()
		} else {
			continue
		}

		// Get corresponding Pinecone result
		pineconeResult, exists := idToResultMap[contentID]
		if !exists {
			continue
		}

		// Calculate hybrid score
		vectorScore := pineconeResult["vector_score"].(float32)
		popularityScore := pc.calculatePopularityScore(contentData, contentType)
		titleRelevance := pc.calculateTitleRelevance(contentData, originalQuery)

		// Hybrid scoring: 60% vector + 25% popularity + 15% title relevance
		hybridScore := (float64(vectorScore) * 0.6) + (popularityScore * 0.25) + (titleRelevance * 0.15)

		// Convert bson.M to map[string]interface{} for JSON serialization
		dataMap := make(map[string]interface{})
		for k, v := range contentData {
			dataMap[k] = v
		}

		// Create enriched result
		enrichedResult := map[string]interface{}{
			"id":               contentID,
			"vector_score":     vectorScore,
			"popularity_score": popularityScore,
			"title_relevance":  titleRelevance,
			"hybrid_score":     hybridScore,
			"type":             contentType,
			"data":             dataMap,
		}

		enrichedResults = append(enrichedResults, enrichedResult)
	}

	if err := cursor.Err(); err != nil {
		logrus.WithError(err).Warn("cursor error during batch processing")
	}

	logrus.WithFields(logrus.Fields{
		"contentType":     contentType,
		"pineconeResults": len(pineconeResults),
		"mongoResults":    len(enrichedResults),
	}).Info("batch enrichment completed")

	return enrichedResults, nil
}

// getContentByIDFromMongoDB fetches content data from MongoDB
func (pc *PineconeController) getContentByIDFromMongoDB(contentID string, collection *mongo.Collection) (bson.M, error) {
	var query bson.M

	// Try ObjectID first, then string ID
	if objectID, err := primitive.ObjectIDFromHex(contentID); err == nil {
		query = bson.M{"_id": objectID}
	} else {
		query = bson.M{"_id": contentID}
	}

	var result bson.M
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find content with ID %s: %w", contentID, err)
	}

	return result, nil
}

// calculatePopularityScore calculates popularity score based on content type
func (pc *PineconeController) calculatePopularityScore(contentData bson.M, contentType string) float64 {
	switch contentType {
	case "movie":
		if vote, ok := contentData["tmdb_vote"].(float64); ok {
			if voteCount, ok := contentData["tmdb_vote_count"].(int32); ok {
				// Normalize score (0-1 range)
				normalizedVote := vote / 10.0
				popularityBoost := math.Log(float64(voteCount)+1) / 20.0 // Log scale for vote count
				return math.Min(normalizedVote+popularityBoost, 1.0)
			}
		}
	case "tvseries":
		if vote, ok := contentData["tmdb_vote"].(float64); ok {
			if voteCount, ok := contentData["tmdb_vote_count"].(int32); ok {
				normalizedVote := vote / 10.0
				popularityBoost := math.Log(float64(voteCount)+1) / 20.0
				return math.Min(normalizedVote+popularityBoost, 1.0)
			}
		}
	case "anime":
		if score, ok := contentData["mal_score"].(float64); ok {
			if scoredBy, ok := contentData["mal_scored_by"].(int32); ok {
				normalizedScore := score / 10.0
				popularityBoost := math.Log(float64(scoredBy)+1) / 25.0
				return math.Min(normalizedScore+popularityBoost, 1.0)
			}
		}
	case "game":
		if score, ok := contentData["metacritic_score"].(int32); ok {
			normalizedScore := float64(score) / 100.0
			return math.Min(normalizedScore, 1.0)
		}
	}

	return 0.5 // Default score
}

// calculateTitleRelevance calculates how relevant the title is to the search query
func (pc *PineconeController) calculateTitleRelevance(contentData bson.M, query string) float64 {
	query = strings.ToLower(query)

	// Get title fields based on content type
	var titles []string
	if titleEn, ok := contentData["title_en"].(string); ok && titleEn != "" {
		titles = append(titles, strings.ToLower(titleEn))
	}
	if titleOriginal, ok := contentData["title_original"].(string); ok && titleOriginal != "" {
		titles = append(titles, strings.ToLower(titleOriginal))
	}
	if title, ok := contentData["title"].(string); ok && title != "" {
		titles = append(titles, strings.ToLower(title))
	}
	if titleJp, ok := contentData["title_jp"].(string); ok && titleJp != "" {
		titles = append(titles, strings.ToLower(titleJp))
	}

	maxRelevance := 0.0
	queryWords := strings.Fields(query)

	for _, title := range titles {
		// Exact match gets highest score
		if title == query {
			return 1.0
		}

		// Contains query gets high score
		if strings.Contains(title, query) {
			maxRelevance = math.Max(maxRelevance, 0.8)
			continue
		}

		// Word overlap scoring
		titleWords := strings.Fields(title)
		matchingWords := 0

		for _, queryWord := range queryWords {
			for _, titleWord := range titleWords {
				if strings.Contains(titleWord, queryWord) || strings.Contains(queryWord, titleWord) {
					matchingWords++
					break
				}
			}
		}

		if len(queryWords) > 0 {
			wordOverlapScore := float64(matchingWords) / float64(len(queryWords)) * 0.6
			maxRelevance = math.Max(maxRelevance, wordOverlapScore)
		}
	}

	return maxRelevance
}

// getEmbeddingFromCache retrieves cached embedding from Redis
func (pc *PineconeController) getEmbeddingFromCache(query string) ([]float32, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := pc.generateCacheKey(EmbeddingCachePrefix, query)
	data, err := pc.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var embedding []float32
	if err := json.Unmarshal(data, &embedding); err != nil {
		logrus.WithError(err).Warn("failed to unmarshal cached embedding")
		return nil, false
	}

	return embedding, true
}

// setEmbeddingCache stores embedding in Redis with TTL
func (pc *PineconeController) setEmbeddingCache(query string, embedding []float32) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := pc.generateCacheKey(EmbeddingCachePrefix, query)
	data, err := json.Marshal(embedding)
	if err != nil {
		logrus.WithError(err).Warn("failed to marshal embedding for cache")
		return
	}

	if err := pc.RedisClient.Set(ctx, key, data, EmbeddingCacheTTL).Err(); err != nil {
		logrus.WithError(err).Warn("failed to cache embedding")
	}
}

// getSearchFromCache retrieves cached search results from Redis
func (pc *PineconeController) getSearchFromCache(query, contentType string, page, limit int) ([]map[string]interface{}, int, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := pc.generateCacheKey(SearchCachePrefix, query, contentType, fmt.Sprintf("%d", page), fmt.Sprintf("%d", limit))
	data, err := pc.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, 0, false
	}

	var cached struct {
		Results      []map[string]interface{} `json:"results"`
		TotalResults int                      `json:"total_results"`
	}
	if err := json.Unmarshal(data, &cached); err != nil {
		logrus.WithError(err).Warn("failed to unmarshal cached search results")
		return nil, 0, false
	}

	return cached.Results, cached.TotalResults, true
}

// setSearchCache stores search results in Redis with TTL
func (pc *PineconeController) setSearchCache(query, contentType string, page, limit int, results []map[string]interface{}, totalResults int) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := pc.generateCacheKey(SearchCachePrefix, query, contentType, fmt.Sprintf("%d", page), fmt.Sprintf("%d", limit))
	cached := struct {
		Results      []map[string]interface{} `json:"results"`
		TotalResults int                      `json:"total_results"`
	}{
		Results:      results,
		TotalResults: totalResults,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		logrus.WithError(err).Warn("failed to marshal search results for cache")
		return
	}

	if err := pc.RedisClient.Set(ctx, key, data, SearchCacheTTL).Err(); err != nil {
		logrus.WithError(err).Warn("failed to cache search results")
	}
}
