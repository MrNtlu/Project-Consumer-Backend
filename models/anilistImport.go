package models

import (
	"app/db"
	"app/responses"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type AniListImportModel struct {
	AnimeCollection     *mongo.Collection
	MangaCollection     *mongo.Collection
	AnimeListCollection *mongo.Collection
	MangaListCollection *mongo.Collection
}

func NewAniListImportModel(mongoDB *db.MongoDB) *AniListImportModel {
	return &AniListImportModel{
		AnimeCollection:     mongoDB.Database.Collection("animes"),
		MangaCollection:     mongoDB.Database.Collection("mangas"),
		AnimeListCollection: mongoDB.Database.Collection("anime-lists"),
		MangaListCollection: mongoDB.Database.Collection("manga-lists"),
	}
}

func (a *AniListImportModel) ImportUserLists(userID, anilistUsername string) (responses.AniListImportResponse, error) {
	logrus.WithFields(logrus.Fields{
		"user_id":          userID,
		"anilist_username": anilistUsername,
	}).Info("Starting AniList import")

	// Get existing entries in bulk
	existingAnimeEntries, err := a.getExistingAnimeEntries(userID)
	if err != nil {
		return responses.AniListImportResponse{}, fmt.Errorf("failed to get existing anime entries: %v", err)
	}

	existingMangaEntries, err := a.getExistingMangaEntries(userID)
	if err != nil {
		return responses.AniListImportResponse{}, fmt.Errorf("failed to get existing manga entries: %v", err)
	}

	// Get all AniList IDs from database
	animeAniListIDs, err := a.getAllAnimeAniListIDs()
	if err != nil {
		return responses.AniListImportResponse{}, fmt.Errorf("failed to get anime AniList IDs: %v", err)
	}

	mangaAniListIDs, err := a.getAllMangaAniListIDs()
	if err != nil {
		return responses.AniListImportResponse{}, fmt.Errorf("failed to get manga AniList IDs: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":             userID,
		"anime_anilist_count": len(animeAniListIDs),
		"manga_anilist_count": len(mangaAniListIDs),
	}).Info("loaded AniList mappings from database")

	var importedCount, skippedCount, errorCount int
	var animeEntriesToInsert []interface{}
	var mangaEntriesToInsert []interface{}
	var importedTitles, skippedTitles []string

	// Fetch anime list
	animeEntries, err := a.fetchUserMediaList(anilistUsername, "ANIME")
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch anime list, continuing with manga")
	} else {
		a.processAnimeEntries(animeEntries, userID, animeAniListIDs, existingAnimeEntries,
			&animeEntriesToInsert, &importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Fetch manga list
	mangaEntries, err := a.fetchUserMediaList(anilistUsername, "MANGA")
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch manga list")
	} else {
		a.processMangaEntries(mangaEntries, userID, mangaAniListIDs, existingMangaEntries,
			&mangaEntriesToInsert, &importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Bulk insert anime entries
	if len(animeEntriesToInsert) > 0 {
		_, err = a.AnimeListCollection.InsertMany(context.TODO(), animeEntriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(animeEntriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert anime entries")
			return responses.AniListImportResponse{}, fmt.Errorf("failed to import anime entries: %v", err)
		}
	}

	// Bulk insert manga entries
	if len(mangaEntriesToInsert) > 0 {
		_, err = a.MangaListCollection.InsertMany(context.TODO(), mangaEntriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(mangaEntriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert manga entries")
			return responses.AniListImportResponse{}, fmt.Errorf("failed to import manga entries: %v", err)
		}
	}

	message := fmt.Sprintf("Import completed: %d imported, %d skipped, %d errors",
		importedCount, skippedCount, errorCount)

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
	}).Info("AniList import completed")

	return responses.AniListImportResponse{
		ImportedCount:  importedCount,
		SkippedCount:   skippedCount,
		ErrorCount:     errorCount,
		Message:        message,
		ImportedTitles: importedTitles,
		SkippedTitles:  skippedTitles,
	}, nil
}

func (a *AniListImportModel) fetchUserMediaList(username, mediaType string) ([]responses.AniListMediaListEntry, error) {
	query := `
		query ($username: String, $type: MediaType) {
			MediaListCollection(userName: $username, type: $type) {
				lists {
					name
					isCustomList
					isSplitCompletedList
					status
					entries {
						id
						status
						score
						progress
						progressVolumes
						repeat
						priority
						private
						notes
						hiddenFromStatusLists
						customLists
						startedAt {
							year
							month
							day
						}
						completedAt {
							year
							month
							day
						}
						updatedAt
						createdAt
						media {
							id
							title {
								romaji
								english
								native
							}
							episodes
							chapters
							volumes
							status
							type
						}
					}
				}
				user {
					id
					name
					mediaListOptions {
						scoreFormat
						animeList {
							sectionOrder
							splitCompletedSectionByFormat
							customLists
						}
						mangaList {
							sectionOrder
							splitCompletedSectionByFormat
							customLists
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"username": username,
		"type":     mediaType,
	}

	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make GraphQL request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user not found or media list is private")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("AniList API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var graphqlResponse responses.AniListGraphQLResponse
	if err := json.Unmarshal(body, &graphqlResponse); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %v", err)
	}

	if len(graphqlResponse.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %s", graphqlResponse.Errors[0].Message)
	}

	// Flatten all entries from all lists
	var allEntries []responses.AniListMediaListEntry
	for _, list := range graphqlResponse.Data.MediaListCollection.Lists {
		allEntries = append(allEntries, list.Entries...)
	}

	return allEntries, nil
}

func (a *AniListImportModel) processAnimeEntries(entries []responses.AniListMediaListEntry, userID string,
	animeAniListIDs map[int64]string, existingAnimeEntries map[int64]bool,
	animeEntriesToInsert *[]interface{}, importedCount, skippedCount, errorCount *int,
	importedTitles, skippedTitles *[]string) {

	for _, entry := range entries {
		title := a.getPreferredTitle(entry.Media.Title)

		logrus.WithFields(logrus.Fields{
			"anilist_id": entry.Media.ID,
			"title":      title,
			"status":     entry.Status,
		}).Debug("processing AniList anime entry")

		// Check if anime exists in our database
		animeObjectID, exists := animeAniListIDs[int64(entry.Media.ID)]
		if !exists {
			logrus.WithFields(logrus.Fields{
				"anilist_id": entry.Media.ID,
				"title":      title,
			}).Debug("anime not found in database, skipping")
			*errorCount++
			continue
		}

		// Check if already exists in user's list
		if _, exists := existingAnimeEntries[int64(entry.Media.ID)]; exists {
			*skippedCount++
			*skippedTitles = append(*skippedTitles, title)
			continue
		}

		// Convert AniList status to application status
		status := a.convertAniListStatus(entry.Status)

		// Convert score
		var score *float32
		if entry.Score > 0 {
			convertedScore := a.convertAniListScore(entry.Score, "POINT_10") // Assume 10-point for now
			score = &convertedScore
		}

		// Create anime list entry
		animeListEntry := AnimeList{
			UserID:          userID,
			AnimeID:         animeObjectID,
			AnimeMALID:      int64(entry.Media.ID), // Using AniList ID in MAL field for now
			Status:          status,
			WatchedEpisodes: int64(entry.Progress),
			Score:           score,
			TimesFinished:   entry.Repeat,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		*animeEntriesToInsert = append(*animeEntriesToInsert, animeListEntry)
		*importedCount++
		*importedTitles = append(*importedTitles, title)
	}
}

func (a *AniListImportModel) processMangaEntries(entries []responses.AniListMediaListEntry, userID string,
	mangaAniListIDs map[int64]string, existingMangaEntries map[int64]bool,
	mangaEntriesToInsert *[]interface{}, importedCount, skippedCount, errorCount *int,
	importedTitles, skippedTitles *[]string) {

	for _, entry := range entries {
		title := a.getPreferredTitle(entry.Media.Title)

		logrus.WithFields(logrus.Fields{
			"anilist_id": entry.Media.ID,
			"title":      title,
			"status":     entry.Status,
		}).Debug("processing AniList manga entry")

		// Check if manga exists in our database
		mangaObjectID, exists := mangaAniListIDs[int64(entry.Media.ID)]
		if !exists {
			logrus.WithFields(logrus.Fields{
				"anilist_id": entry.Media.ID,
				"title":      title,
			}).Debug("manga not found in database, skipping")
			*errorCount++
			continue
		}

		// Check if already exists in user's list
		if _, exists := existingMangaEntries[int64(entry.Media.ID)]; exists {
			*skippedCount++
			*skippedTitles = append(*skippedTitles, title)
			continue
		}

		// Convert AniList status to application status
		status := a.convertAniListStatus(entry.Status)

		// Convert score
		var score *float32
		if entry.Score > 0 {
			convertedScore := a.convertAniListScore(entry.Score, "POINT_10") // Assume 10-point for now
			score = &convertedScore
		}

		// Create manga list entry
		mangaListEntry := MangaList{
			UserID:        userID,
			MangaID:       mangaObjectID,
			MangaMALID:    int64(entry.Media.ID), // Using AniList ID in MAL field for now
			Status:        status,
			ReadChapters:  int64(entry.Progress),
			ReadVolumes:   0, // AniList progress volumes not always available
			Score:         score,
			TimesFinished: entry.Repeat,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		*mangaEntriesToInsert = append(*mangaEntriesToInsert, mangaListEntry)
		*importedCount++
		*importedTitles = append(*importedTitles, title)
	}
}

func (a *AniListImportModel) getPreferredTitle(title struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
	Native  string `json:"native"`
}) string {
	if title.English != "" {
		return title.English
	}
	if title.Romaji != "" {
		return title.Romaji
	}
	return title.Native
}

func (a *AniListImportModel) convertAniListStatus(status string) string {
	switch strings.ToUpper(status) {
	case "CURRENT":
		return "active"
	case "COMPLETED":
		return "finished"
	case "PLANNING":
		return "active"
	case "PAUSED":
		return "active"
	case "DROPPED":
		return "dropped"
	case "REPEATING":
		return "finished"
	default:
		return "active"
	}
}

func (a *AniListImportModel) convertAniListScore(score int, scoreFormat string) float32 {
	switch scoreFormat {
	case "POINT_100":
		return float32(score) / 10.0
	case "POINT_10_DECIMAL":
		return float32(score)
	case "POINT_10":
		return float32(score)
	case "POINT_5":
		return float32(score) * 2.0
	case "POINT_3":
		// Convert 3-point to 10-point scale
		switch score {
		case 1:
			return 3.0 // :(
		case 2:
			return 6.0 // :|
		case 3:
			return 9.0 // :)
		default:
			return float32(score)
		}
	default:
		return float32(score)
	}
}

func (a *AniListImportModel) getExistingAnimeEntries(userID string) (map[int64]bool, error) {
	cursor, err := a.AnimeListCollection.Find(context.TODO(), bson.M{
		"user_id": userID,
	}, options.Find().SetProjection(bson.M{"anime_anilist_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	existingEntries := make(map[int64]bool)
	for cursor.Next(context.TODO()) {
		var entry struct {
			AnimeAniListID int64 `bson:"anime_anilist_id"`
		}
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		existingEntries[entry.AnimeAniListID] = true
	}

	return existingEntries, nil
}

func (a *AniListImportModel) getExistingMangaEntries(userID string) (map[int64]bool, error) {
	cursor, err := a.MangaListCollection.Find(context.TODO(), bson.M{
		"user_id": userID,
	}, options.Find().SetProjection(bson.M{"manga_anilist_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	existingEntries := make(map[int64]bool)
	for cursor.Next(context.TODO()) {
		var entry struct {
			MangaAniListID int64 `bson:"manga_anilist_id"`
		}
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		existingEntries[entry.MangaAniListID] = true
	}

	return existingEntries, nil
}

func (a *AniListImportModel) getAllAnimeAniListIDs() (map[int64]string, error) {
	cursor, err := a.AnimeCollection.Find(context.TODO(), bson.M{
		"anilist_id": bson.M{"$exists": true, "$ne": nil},
	}, options.Find().SetProjection(bson.M{"_id": 1, "anilist_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	animeAniListIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var anime struct {
			ID        primitive.ObjectID `bson:"_id"`
			AniListID int64              `bson:"anilist_id"`
		}
		if err := cursor.Decode(&anime); err != nil {
			continue
		}
		animeAniListIDs[anime.AniListID] = anime.ID.Hex()
	}

	return animeAniListIDs, nil
}

func (a *AniListImportModel) getAllMangaAniListIDs() (map[int64]string, error) {
	cursor, err := a.MangaCollection.Find(context.TODO(), bson.M{
		"anilist_id": bson.M{"$exists": true, "$ne": nil},
	}, options.Find().SetProjection(bson.M{"_id": 1, "anilist_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	mangaAniListIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var manga struct {
			ID        primitive.ObjectID `bson:"_id"`
			AniListID int64              `bson:"anilist_id"`
		}
		if err := cursor.Decode(&manga); err != nil {
			continue
		}
		mangaAniListIDs[manga.AniListID] = manga.ID.Hex()
	}

	return mangaAniListIDs, nil
}
