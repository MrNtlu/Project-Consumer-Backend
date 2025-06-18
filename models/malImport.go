package models

import (
	"app/db"
	"app/responses"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type MALImportModel struct {
	AnimeCollection     *mongo.Collection
	AnimeListCollection *mongo.Collection
}

func NewMALImportModel(mongoDB *db.MongoDB) *MALImportModel {
	return &MALImportModel{
		AnimeCollection:     mongoDB.Database.Collection("animes"),
		AnimeListCollection: mongoDB.Database.Collection("anime-lists"),
	}
}

type MALAPIResponse struct {
	Data []struct {
		Node struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"node"`
		ListStatus struct {
			Status             string `json:"status"`
			Score              int    `json:"score"`
			NumEpisodesWatched int    `json:"num_episodes_watched"`
			IsRewatching       bool   `json:"is_rewatching"`
			NumTimesRewatched  int    `json:"num_times_rewatched"`
			StartDate          string `json:"start_date"`
			FinishDate         string `json:"finish_date"`
		} `json:"list_status"`
	} `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func (m *MALImportModel) ImportUserAnimeList(userID, malUsername string) (responses.MALImportResponse, error) {
	// Fetch anime list from MAL
	animeEntries, err := m.fetchMALAnimeList(malUsername)
	if err != nil {
		return responses.MALImportResponse{}, err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"mal_username":  malUsername,
		"total_entries": len(animeEntries),
	}).Info("Starting MAL import")

	// Get existing anime entries in bulk to avoid repeated database calls
	existingEntries, err := m.getExistingAnimeEntries(userID)
	if err != nil {
		return responses.MALImportResponse{}, fmt.Errorf("failed to get existing entries: %v", err)
	}

	// Get all anime MAL IDs from our database in bulk
	animeMALIDs, err := m.getAllAnimeMALIDs()
	if err != nil {
		return responses.MALImportResponse{}, fmt.Errorf("failed to get anime MAL IDs: %v", err)
	}

	var importedCount, skippedCount, errorCount int
	var entriesToInsert []interface{}
	var importedTitles, skippedTitles []string

	// Process entries in batches
	for _, entry := range animeEntries {
		// Check if anime exists in our database
		animeObjectID, exists := animeMALIDs[int64(entry.ID)]
		if !exists {
			logrus.WithFields(logrus.Fields{
				"mal_id": entry.ID,
				"title":  entry.Title,
			}).Debug("anime not found in database, skipping")
			errorCount++
			continue
		}

		// Check if already exists in user's list
		if _, exists := existingEntries[int64(entry.ID)]; exists {
			skippedCount++
			skippedTitles = append(skippedTitles, entry.Title)
			continue
		}

		// Convert score to float32 pointer
		var score *float32
		if entry.Score != nil {
			scoreFloat := float32(*entry.Score)
			score = &scoreFloat
		}

		// Create anime list entry for batch insert
		animeListEntry := AnimeList{
			UserID:          userID,
			AnimeID:         animeObjectID,
			AnimeMALID:      int64(entry.ID),
			Status:          entry.Status,
			WatchedEpisodes: int64(entry.WatchedEpisodes),
			Score:           score,
			TimesFinished:   entry.TimesFinished,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		entriesToInsert = append(entriesToInsert, animeListEntry)
		importedCount++
		importedTitles = append(importedTitles, entry.Title)
	}

	// Bulk insert all entries at once
	if len(entriesToInsert) > 0 {
		_, err = m.AnimeListCollection.InsertMany(context.TODO(), entriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(entriesToInsert),
			}).Error("failed to bulk insert anime entries: ", err)
			return responses.MALImportResponse{}, fmt.Errorf("failed to import anime entries: %v", err)
		}
	}

	message := fmt.Sprintf("Import completed: %d imported, %d skipped, %d errors",
		importedCount, skippedCount, errorCount)

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
	}).Info("MAL import completed")

	return responses.MALImportResponse{
		ImportedCount:  importedCount,
		SkippedCount:   skippedCount,
		ErrorCount:     errorCount,
		Message:        message,
		ImportedTitles: importedTitles,
		SkippedTitles:  skippedTitles,
	}, nil
}

func (m *MALImportModel) fetchMALAnimeList(username string) ([]responses.MALAnimeEntry, error) {
	// Using the unofficial MAL API endpoint that doesn't require authentication
	// This endpoint is publicly accessible for user lists
	url := fmt.Sprintf("https://myanimelist.net/animelist/%s/load.json", username)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers to mimic a browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch anime list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user not found or anime list is private")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse the JSON response
	var malData []map[string]interface{}
	if err := json.Unmarshal(body, &malData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	var animeEntries []responses.MALAnimeEntry
	for _, item := range malData {
		// Safely extract values with type checking
		animeID, ok := item["anime_id"].(float64)
		if !ok {
			continue // Skip if anime_id is not a number
		}

		// Handle anime_title - it might be string or other type
		var title string
		if titleVal, exists := item["anime_title"]; exists {
			if titleStr, ok := titleVal.(string); ok {
				title = titleStr
			} else {
				// Convert to string if it's not already a string
				title = fmt.Sprintf("%v", titleVal)
			}
		}

		watchedEpisodes, _ := item["num_watched_episodes"].(float64)

		entry := responses.MALAnimeEntry{
			ID:              int(animeID),
			Title:           title,
			WatchedEpisodes: int(watchedEpisodes),
			TimesFinished:   1, // Default to 1 if completed
		}

		// Map MAL status to our status
		entry.Status = "active" // Default status
		if statusVal, ok := item["status"].(float64); ok {
			malStatus := int(statusVal)
			switch malStatus {
			case 1:
				entry.Status = "active" // Currently Watching
			case 2:
				entry.Status = "finished" // Completed
			case 3:
				entry.Status = "active" // On-Hold (map to active)
			case 4:
				entry.Status = "dropped" // Dropped
			case 6:
				entry.Status = "active" // Plan to Watch (map to active)
			default:
				entry.Status = "active"
			}
		}

		// Handle score (0 means no score)
		if scoreVal, exists := item["score"]; exists {
			if score, ok := scoreVal.(float64); ok && score > 0 {
				scoreInt := int(score)
				entry.Score = &scoreInt
			}
		}

		// Handle dates if present
		if startDateVal, exists := item["start_date_string"]; exists {
			if startDate, ok := startDateVal.(string); ok && startDate != "" {
				entry.StartDate = &startDate
			}
		}
		if finishDateVal, exists := item["finish_date_string"]; exists {
			if finishDate, ok := finishDateVal.(string); ok && finishDate != "" {
				entry.FinishDate = &finishDate
			}
		}

		animeEntries = append(animeEntries, entry)
	}

	return animeEntries, nil
}

// getExistingAnimeEntries gets all existing anime entries for a user in bulk
func (m *MALImportModel) getExistingAnimeEntries(userID string) (map[int64]bool, error) {
	cursor, err := m.AnimeListCollection.Find(context.TODO(), bson.M{
		"user_id": userID,
	}, options.Find().SetProjection(bson.M{"anime_mal_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	existingEntries := make(map[int64]bool)
	for cursor.Next(context.TODO()) {
		var entry struct {
			AnimeMALID int64 `bson:"anime_mal_id"`
		}
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		existingEntries[entry.AnimeMALID] = true
	}

	return existingEntries, nil
}

// getAllAnimeMALIDs gets all anime MAL IDs from our database in bulk
func (m *MALImportModel) getAllAnimeMALIDs() (map[int64]string, error) {
	cursor, err := m.AnimeCollection.Find(context.TODO(), bson.M{
		"mal_id": bson.M{"$exists": true, "$ne": nil},
	}, options.Find().SetProjection(bson.M{"_id": 1, "mal_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	animeMALIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var anime struct {
			ID    primitive.ObjectID `bson:"_id"`
			MALID int64              `bson:"mal_id"`
		}
		if err := cursor.Decode(&anime); err != nil {
			continue
		}
		animeMALIDs[anime.MALID] = anime.ID.Hex()
	}

	return animeMALIDs, nil
}
