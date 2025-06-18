package models

import (
	"app/db"
	"app/responses"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type SteamImportModel struct {
	GameCollection     *mongo.Collection
	GameListCollection *mongo.Collection
}

func NewSteamImportModel(mongoDB *db.MongoDB) *SteamImportModel {
	return &SteamImportModel{
		GameCollection:     mongoDB.Database.Collection("games"),
		GameListCollection: mongoDB.Database.Collection("game-lists"),
	}
}

func (s *SteamImportModel) ImportUserGameLibrary(userID, steamID string) (responses.SteamImportResponse, error) {
	// Validate Steam ID and check profile visibility
	if err := s.validateSteamProfile(steamID); err != nil {
		return responses.SteamImportResponse{}, err
	}

	// Fetch game library from Steam
	gameEntries, err := s.fetchSteamGameLibrary(steamID)
	if err != nil {
		return responses.SteamImportResponse{}, err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"steam_id":      steamID,
		"total_entries": len(gameEntries),
	}).Info("Starting Steam import")

	// Get existing game entries in bulk to avoid repeated database calls
	existingEntries, err := s.getExistingGameEntries(userID)
	if err != nil {
		return responses.SteamImportResponse{}, fmt.Errorf("failed to get existing entries: %v", err)
	}

	// Get all game Steam App IDs from our database in bulk
	gameSteamIDs, err := s.getAllGameSteamIDs()
	if err != nil {
		return responses.SteamImportResponse{}, fmt.Errorf("failed to get game Steam IDs: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":           userID,
		"steam_games_count": len(gameSteamIDs),
	}).Info("loaded Steam game mappings from database")

	var importedCount, skippedCount, errorCount int
	var entriesToInsert []interface{}
	var importedTitles, skippedTitles []string

	// Process entries in batches
	for _, entry := range gameEntries {
		logrus.WithFields(logrus.Fields{
			"steam_appid":   entry.AppID,
			"title":         entry.Name,
			"playtime_mins": entry.PlaytimeForever,
		}).Debug("processing Steam game entry")

		// Check if game exists in our database
		gameObjectID, exists := gameSteamIDs[int64(entry.AppID)]
		if !exists {
			logrus.WithFields(logrus.Fields{
				"steam_appid": entry.AppID,
				"title":       entry.Name,
			}).Debug("game not found in database, skipping")
			errorCount++
			continue
		}

		logrus.WithFields(logrus.Fields{
			"steam_appid": entry.AppID,
			"title":       entry.Name,
			"game_id":     gameObjectID,
		}).Debug("found matching game in database")

		// Determine status based on playtime
		status := s.determineGameStatus(entry.PlaytimeForever)

		// Convert playtime from minutes to hours for storage
		playtimeHours := float32(entry.PlaytimeForever) / 60.0
		playtimeHoursInt := int(playtimeHours)

		// Note: Steam provides last played timestamp (entry.RtimeLastPlayed) but
		// the GameList struct doesn't have a last_played_at field to store it

		// Check if already exists in user's list
		if _, exists := existingEntries[gameObjectID]; exists {
			// Update existing entry with new playtime data
			err := s.updateExistingGameEntry(userID, gameObjectID, playtimeHoursInt, status, s.calculateTimesFinished(entry.PlaytimeForever))
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"steam_appid": entry.AppID,
					"title":       entry.Name,
					"game_id":     gameObjectID,
					"error":       err.Error(),
				}).Error("failed to update existing game entry")
				errorCount++
			} else {
				logrus.WithFields(logrus.Fields{
					"steam_appid":  entry.AppID,
					"title":        entry.Name,
					"game_id":      gameObjectID,
					"hours_played": playtimeHoursInt,
				}).Debug("updated existing game entry with new playtime")
				skippedCount++ // Count as skipped since it wasn't a new import
				skippedTitles = append(skippedTitles, entry.Name)
			}
			continue
		}

		// Create game list entry for batch insert
		gameListEntry := GameList{
			UserID:        userID,
			GameID:        gameObjectID,
			GameRAWGID:    int64(entry.AppID), // Using Steam AppID as identifier
			Status:        status,
			HoursPlayed:   &playtimeHoursInt, // Always include hours, even if 0
			TimesFinished: s.calculateTimesFinished(entry.PlaytimeForever),
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		logrus.WithFields(logrus.Fields{
			"steam_appid":    entry.AppID,
			"title":          entry.Name,
			"game_id":        gameObjectID,
			"status":         status,
			"hours_played":   playtimeHoursInt,
			"times_finished": s.calculateTimesFinished(entry.PlaytimeForever),
		}).Debug("prepared game entry for import")

		entriesToInsert = append(entriesToInsert, gameListEntry)
		importedCount++
		importedTitles = append(importedTitles, entry.Name)
	}

	// Bulk insert all entries at once
	if len(entriesToInsert) > 0 {
		logrus.WithFields(logrus.Fields{
			"user_id":      userID,
			"insert_count": len(entriesToInsert),
		}).Info("attempting bulk insert of game entries")

		_, err = s.GameListCollection.InsertMany(context.TODO(), entriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(entriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert game entries")
			return responses.SteamImportResponse{}, fmt.Errorf("failed to import game entries: %v", err)
		}

		logrus.WithFields(logrus.Fields{
			"user_id":      userID,
			"insert_count": len(entriesToInsert),
		}).Info("successfully bulk inserted game entries")
	} else {
		logrus.WithFields(logrus.Fields{
			"user_id": userID,
		}).Info("no new entries to insert")
	}

	message := fmt.Sprintf("Import completed: %d imported, %d skipped, %d errors",
		importedCount, skippedCount, errorCount)

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
	}).Info("Steam import completed")

	return responses.SteamImportResponse{
		ImportedCount:  importedCount,
		SkippedCount:   skippedCount,
		ErrorCount:     errorCount,
		Message:        message,
		ImportedTitles: importedTitles,
		SkippedTitles:  skippedTitles,
	}, nil
}

func (s *SteamImportModel) ResolveSteamUsername(username string) (string, error) {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("Steam API key not configured")
	}

	// Use ResolveVanityURL API to convert username to Steam ID
	url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v1/?key=%s&vanityurl=%s", apiKey, username)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Steam username: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Steam API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Steam API response: %v", err)
	}

	var vanityResponse struct {
		Response struct {
			SteamID string `json:"steamid"`
			Success int    `json:"success"`
			Message string `json:"message"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &vanityResponse); err != nil {
		return "", fmt.Errorf("failed to parse Steam API response: %v", err)
	}

	if vanityResponse.Response.Success != 1 {
		return "", fmt.Errorf("Steam username '%s' not found or invalid", username)
	}

	logrus.WithFields(logrus.Fields{
		"username": username,
		"steam_id": vanityResponse.Response.SteamID,
	}).Info("resolved Steam username to Steam ID")

	return vanityResponse.Response.SteamID, nil
}

func (s *SteamImportModel) validateSteamProfile(steamID string) error {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("Steam API key not configured")
	}

	url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s", apiKey, steamID)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to validate Steam profile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Steam API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Steam API response: %v", err)
	}

	var playerSummary responses.SteamPlayerSummaryResponse
	if err := json.Unmarshal(body, &playerSummary); err != nil {
		return fmt.Errorf("failed to parse Steam API response: %v", err)
	}

	if len(playerSummary.Response.Players) == 0 {
		return fmt.Errorf("Steam profile not found or invalid Steam ID")
	}

	player := playerSummary.Response.Players[0]
	if player.CommunityVisibilityState != 3 {
		return fmt.Errorf("Steam profile is private. Please set your profile and game details to public")
	}

	return nil
}

func (s *SteamImportModel) fetchSteamGameLibrary(steamID string) ([]responses.SteamGameEntry, error) {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("Steam API key not configured")
	}

	url := fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?key=%s&steamid=%s&include_appinfo=true&include_played_free_games=true", apiKey, steamID)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Steam game library: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Steam API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Steam API response: %v", err)
	}

	var ownedGames responses.SteamOwnedGamesResponse
	if err := json.Unmarshal(body, &ownedGames); err != nil {
		return nil, fmt.Errorf("failed to parse Steam API response: %v", err)
	}

	if len(ownedGames.Response.Games) == 0 {
		return nil, fmt.Errorf("no games found in Steam library or library is private")
	}

	return ownedGames.Response.Games, nil
}

func (s *SteamImportModel) determineGameStatus(playtimeMinutes int) string {
	if playtimeMinutes == 0 {
		return "planning" // Never played
	} else if playtimeMinutes < 60 { // Less than 1 hour
		return "active" // Started but not much progress
	} else if playtimeMinutes >= 60 {
		return "active" // Actively playing or played
	}
	return "active"
}

func (s *SteamImportModel) calculateTimesFinished(playtimeMinutes int) int {
	// This is a rough estimation - could be improved with achievement data
	// For now, assume games over 20 hours might be finished once
	if playtimeMinutes >= 1200 { // 20+ hours
		return 1
	}
	return 0
}

func (s *SteamImportModel) getExistingGameEntries(userID string) (map[string]bool, error) {
	filter := bson.M{"user_id": userID}
	projection := bson.M{"game_id": 1}

	cursor, err := s.GameListCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("failed to query existing game entries")
		return nil, err
	}
	defer cursor.Close(context.TODO())

	existingEntries := make(map[string]bool)
	for cursor.Next(context.TODO()) {
		var entry struct {
			GameID string `bson:"game_id"`
		}
		if err := cursor.Decode(&entry); err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Debug("failed to decode existing game entry")
			continue
		}
		existingEntries[entry.GameID] = true
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"existing_count": len(existingEntries),
	}).Info("loaded existing game entries for user")

	return existingEntries, nil
}

func (s *SteamImportModel) getAllGameSteamIDs() (map[int64]string, error) {
	filter := bson.M{"stores": bson.M{"$exists": true, "$ne": nil}}
	projection := bson.M{"_id": 1, "stores": 1, "title": 1}

	cursor, err := s.GameCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to query games collection for Steam mappings")
		return nil, err
	}
	defer cursor.Close(context.TODO())

	gameSteamIDs := make(map[int64]string)
	processedCount := 0
	steamGamesFound := 0

	for cursor.Next(context.TODO()) {
		var game struct {
			ID     primitive.ObjectID `bson:"_id"`
			Title  string             `bson:"title"`
			Stores []struct {
				StoreID int    `bson:"store_id"`
				URL     string `bson:"url"`
			} `bson:"stores"`
		}
		if err := cursor.Decode(&game); err != nil {
			logrus.WithFields(logrus.Fields{
				"game_id": game.ID.Hex(),
				"error":   err.Error(),
			}).Debug("failed to decode game document")
			continue
		}

		processedCount++

		// Look for Steam store (store_id = 1) and extract App ID from URL
		for _, store := range game.Stores {
			if store.StoreID == 1 { // Steam store ID is 1
				steamAppID := s.extractSteamAppIDFromURL(store.URL)
				if steamAppID != 0 {
					gameSteamIDs[steamAppID] = game.ID.Hex()
					steamGamesFound++
					logrus.WithFields(logrus.Fields{
						"game_id":     game.ID.Hex(),
						"title":       game.Title,
						"steam_appid": steamAppID,
						"steam_url":   store.URL,
					}).Debug("mapped Steam game to database")
					break
				} else {
					logrus.WithFields(logrus.Fields{
						"game_id":   game.ID.Hex(),
						"title":     game.Title,
						"steam_url": store.URL,
					}).Debug("failed to extract Steam App ID from URL")
				}
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"processed_games":   processedCount,
		"steam_games_found": steamGamesFound,
		"steam_mappings":    len(gameSteamIDs),
	}).Info("completed Steam game mapping process")

	return gameSteamIDs, nil
}

func (s *SteamImportModel) updateExistingGameEntry(userID, gameID string, hoursPlayed int, status string, timesFinished int) error {
	filter := bson.M{
		"user_id": userID,
		"game_id": gameID,
	}

	update := bson.M{
		"$set": bson.M{
			"hours_played":   &hoursPlayed,
			"status":         status,
			"times_finished": timesFinished,
			"updated_at":     time.Now().UTC(),
		},
	}

	result, err := s.GameListCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update game entry: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no game entry found to update")
	}

	return nil
}

func (s *SteamImportModel) extractSteamAppIDFromURL(steamURL string) int64 {
	// Steam store URLs are in format: https://store.steampowered.com/app/292030/The_Witcher_3_Wild_Hunt/
	// We need to extract the App ID (292030 in this example)
	re := regexp.MustCompile(`/app/(\d+)/`)
	matches := re.FindStringSubmatch(steamURL)
	if len(matches) >= 2 {
		if appID, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			return appID
		}
	}
	return 0
}
