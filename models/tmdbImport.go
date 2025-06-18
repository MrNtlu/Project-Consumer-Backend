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
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type TMDBImportModel struct {
	MovieCollection     *mongo.Collection
	TVCollection        *mongo.Collection
	MovieListCollection *mongo.Collection
	TVListCollection    *mongo.Collection
}

func NewTMDBImportModel(mongoDB *db.MongoDB) *TMDBImportModel {
	return &TMDBImportModel{
		MovieCollection:     mongoDB.Database.Collection("movies"),
		TVCollection:        mongoDB.Database.Collection("tv-series"),
		MovieListCollection: mongoDB.Database.Collection("movie-watch-lists"),
		TVListCollection:    mongoDB.Database.Collection("tvseries-watch-lists"),
	}
}

func (t *TMDBImportModel) ImportUserData(userID, tmdbUsername string) (responses.TMDBImportResponse, error) {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return responses.TMDBImportResponse{}, fmt.Errorf("TMDB API key not configured")
	}

	// Get account ID from username
	accountID, err := t.getAccountIDFromUsername(tmdbUsername, apiKey)
	if err != nil {
		return responses.TMDBImportResponse{}, fmt.Errorf("failed to get account ID: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"tmdb_username": tmdbUsername,
		"account_id":    accountID,
	}).Info("Starting TMDB import")

	// Get existing entries in bulk
	existingMovieEntries, err := t.getExistingMovieEntries(userID)
	if err != nil {
		return responses.TMDBImportResponse{}, fmt.Errorf("failed to get existing movie entries: %v", err)
	}

	existingTVEntries, err := t.getExistingTVEntries(userID)
	if err != nil {
		return responses.TMDBImportResponse{}, fmt.Errorf("failed to get existing TV entries: %v", err)
	}

	// Get all TMDB IDs from database
	movieTMDBIDs, err := t.getAllMovieTMDBIDs()
	if err != nil {
		return responses.TMDBImportResponse{}, fmt.Errorf("failed to get movie TMDB IDs: %v", err)
	}

	tvTMDBIDs, err := t.getAllTVTMDBIDs()
	if err != nil {
		return responses.TMDBImportResponse{}, fmt.Errorf("failed to get TV TMDB IDs: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":          userID,
		"movie_tmdb_count": len(movieTMDBIDs),
		"tv_tmdb_count":    len(tvTMDBIDs),
	}).Info("loaded TMDB mappings from database")

	var importedCount, skippedCount, errorCount int
	var movieEntriesToInsert []interface{}
	var tvEntriesToInsert []interface{}
	var importedTitles, skippedTitles []string

	// Import watchlist
	watchlistEntries, err := t.fetchWatchlist(accountID, apiKey)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch watchlist, continuing with other data")
	} else {
		t.processEntries(watchlistEntries, userID, "active", movieTMDBIDs, tvTMDBIDs,
			existingMovieEntries, existingTVEntries, &movieEntriesToInsert, &tvEntriesToInsert,
			&importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Import rated items
	ratedEntries, err := t.fetchRatedItems(accountID, apiKey)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch rated items, continuing with other data")
	} else {
		t.processEntries(ratedEntries, userID, "finished", movieTMDBIDs, tvTMDBIDs,
			existingMovieEntries, existingTVEntries, &movieEntriesToInsert, &tvEntriesToInsert,
			&importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Import favorites
	favoriteEntries, err := t.fetchFavorites(accountID, apiKey)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch favorites, continuing with other data")
	} else {
		t.processEntries(favoriteEntries, userID, "finished", movieTMDBIDs, tvTMDBIDs,
			existingMovieEntries, existingTVEntries, &movieEntriesToInsert, &tvEntriesToInsert,
			&importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Bulk insert movie entries
	if len(movieEntriesToInsert) > 0 {
		_, err = t.MovieListCollection.InsertMany(context.TODO(), movieEntriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(movieEntriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert movie entries")
			return responses.TMDBImportResponse{}, fmt.Errorf("failed to import movie entries: %v", err)
		}
	}

	// Bulk insert TV entries
	if len(tvEntriesToInsert) > 0 {
		_, err = t.TVListCollection.InsertMany(context.TODO(), tvEntriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(tvEntriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert TV entries")
			return responses.TMDBImportResponse{}, fmt.Errorf("failed to import TV entries: %v", err)
		}
	}

	message := fmt.Sprintf("Import completed: %d imported, %d skipped, %d errors",
		importedCount, skippedCount, errorCount)

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
	}).Info("TMDB import completed")

	return responses.TMDBImportResponse{
		ImportedCount:  importedCount,
		SkippedCount:   skippedCount,
		ErrorCount:     errorCount,
		Message:        message,
		ImportedTitles: importedTitles,
		SkippedTitles:  skippedTitles,
	}, nil
}

func (t *TMDBImportModel) getAccountIDFromUsername(username, apiKey string) (int, error) {
	// TMDB doesn't have a direct username to account ID endpoint
	// We need to search for the user and get their account details
	// This is a limitation of TMDB API - it's primarily designed for authenticated users
	return 0, fmt.Errorf("TMDB API doesn't support username-based imports without user authentication. Please use session-based authentication instead")
}

func (t *TMDBImportModel) fetchWatchlist(accountID int, apiKey string) ([]responses.TMDBEntry, error) {
	var allEntries []responses.TMDBEntry

	// Fetch movie watchlist
	movieURL := fmt.Sprintf("https://api.themoviedb.org/3/account/%d/watchlist/movies?api_key=%s", accountID, apiKey)
	movieEntries, err := t.fetchTMDBData(movieURL, "movie")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie watchlist: %v", err)
	}
	allEntries = append(allEntries, movieEntries...)

	// Fetch TV watchlist
	tvURL := fmt.Sprintf("https://api.themoviedb.org/3/account/%d/watchlist/tv?api_key=%s", accountID, apiKey)
	tvEntries, err := t.fetchTMDBData(tvURL, "tv")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch TV watchlist: %v", err)
	}
	allEntries = append(allEntries, tvEntries...)

	return allEntries, nil
}

func (t *TMDBImportModel) fetchRatedItems(accountID int, apiKey string) ([]responses.TMDBEntry, error) {
	var allEntries []responses.TMDBEntry

	// Fetch rated movies
	movieURL := fmt.Sprintf("https://api.themoviedb.org/3/account/%d/rated/movies?api_key=%s", accountID, apiKey)
	movieEntries, err := t.fetchTMDBData(movieURL, "movie")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rated movies: %v", err)
	}
	allEntries = append(allEntries, movieEntries...)

	// Fetch rated TV shows
	tvURL := fmt.Sprintf("https://api.themoviedb.org/3/account/%d/rated/tv?api_key=%s", accountID, apiKey)
	tvEntries, err := t.fetchTMDBData(tvURL, "tv")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rated TV shows: %v", err)
	}
	allEntries = append(allEntries, tvEntries...)

	return allEntries, nil
}

func (t *TMDBImportModel) fetchFavorites(accountID int, apiKey string) ([]responses.TMDBEntry, error) {
	var allEntries []responses.TMDBEntry

	// Fetch favorite movies
	movieURL := fmt.Sprintf("https://api.themoviedb.org/3/account/%d/favorite/movies?api_key=%s", accountID, apiKey)
	movieEntries, err := t.fetchTMDBData(movieURL, "movie")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch favorite movies: %v", err)
	}
	allEntries = append(allEntries, movieEntries...)

	// Fetch favorite TV shows
	tvURL := fmt.Sprintf("https://api.themoviedb.org/3/account/%d/favorite/tv?api_key=%s", accountID, apiKey)
	tvEntries, err := t.fetchTMDBData(tvURL, "tv")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch favorite TV shows: %v", err)
	}
	allEntries = append(allEntries, tvEntries...)

	return allEntries, nil
}

func (t *TMDBImportModel) fetchTMDBData(url, mediaType string) ([]responses.TMDBEntry, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized - invalid API key or account access")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("TMDB API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var tmdbResponse responses.TMDBWatchlistResponse
	if err := json.Unmarshal(body, &tmdbResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Set media type for all entries
	for i := range tmdbResponse.Results {
		tmdbResponse.Results[i].MediaType = mediaType
		// Set title based on media type
		if mediaType == "tv" {
			// For TV shows, TMDB uses "name" field instead of "title"
			var tvData map[string]interface{}
			if err := json.Unmarshal(body, &tvData); err == nil {
				if results, ok := tvData["results"].([]interface{}); ok && len(results) > i {
					if tvShow, ok := results[i].(map[string]interface{}); ok {
						if name, ok := tvShow["name"].(string); ok {
							tmdbResponse.Results[i].Title = name
						}
					}
				}
			}
		}
	}

	return tmdbResponse.Results, nil
}

func (t *TMDBImportModel) processEntries(entries []responses.TMDBEntry, userID, status string,
	movieTMDBIDs, tvTMDBIDs map[int64]string, existingMovieEntries, existingTVEntries map[string]bool,
	movieEntriesToInsert, tvEntriesToInsert *[]interface{},
	importedCount, skippedCount, errorCount *int, importedTitles, skippedTitles *[]string) {

	for _, entry := range entries {
		logrus.WithFields(logrus.Fields{
			"tmdb_id":    entry.ID,
			"title":      entry.Title,
			"media_type": entry.MediaType,
		}).Debug("processing TMDB entry")

		if entry.MediaType == "movie" {
			if movieObjectID, exists := movieTMDBIDs[int64(entry.ID)]; exists {
				if _, exists := existingMovieEntries[movieObjectID]; exists {
					*skippedCount++
					*skippedTitles = append(*skippedTitles, entry.Title)
					continue
				}

				var score *float32
				if entry.Rating > 0 {
					score = &entry.Rating
				} else if status == "finished" {
					// For favorites without rating, assign high score
					defaultScore := float32(9.0)
					score = &defaultScore
				}

				movieListEntry := MovieWatchList{
					UserID:    userID,
					MovieID:   movieObjectID,
					Status:    status,
					Score:     score,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}

				*movieEntriesToInsert = append(*movieEntriesToInsert, movieListEntry)
				*importedCount++
				*importedTitles = append(*importedTitles, entry.Title)
			} else {
				*errorCount++
			}
		} else if entry.MediaType == "tv" {
			if tvObjectID, exists := tvTMDBIDs[int64(entry.ID)]; exists {
				if _, exists := existingTVEntries[tvObjectID]; exists {
					*skippedCount++
					*skippedTitles = append(*skippedTitles, entry.Title)
					continue
				}

				var score *float32
				if entry.Rating > 0 {
					score = &entry.Rating
				} else if status == "finished" {
					// For favorites without rating, assign high score
					defaultScore := float32(9.0)
					score = &defaultScore
				}

				tvListEntry := TVSeriesWatchList{
					UserID:    userID,
					TvID:      tvObjectID,
					Status:    status,
					Score:     score,
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				}

				*tvEntriesToInsert = append(*tvEntriesToInsert, tvListEntry)
				*importedCount++
				*importedTitles = append(*importedTitles, entry.Title)
			} else {
				*errorCount++
			}
		}
	}
}

func (t *TMDBImportModel) getExistingMovieEntries(userID string) (map[string]bool, error) {
	filter := bson.M{"user_id": userID}
	projection := bson.M{"movie_id": 1}

	cursor, err := t.MovieListCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	existingEntries := make(map[string]bool)
	for cursor.Next(context.TODO()) {
		var entry struct {
			MovieID string `bson:"movie_id"`
		}
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		existingEntries[entry.MovieID] = true
	}

	return existingEntries, nil
}

func (t *TMDBImportModel) getExistingTVEntries(userID string) (map[string]bool, error) {
	filter := bson.M{"user_id": userID}
	projection := bson.M{"tv_id": 1}

	cursor, err := t.TVListCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	existingEntries := make(map[string]bool)
	for cursor.Next(context.TODO()) {
		var entry struct {
			TvID string `bson:"tv_id"`
		}
		if err := cursor.Decode(&entry); err != nil {
			continue
		}
		existingEntries[entry.TvID] = true
	}

	return existingEntries, nil
}

func (t *TMDBImportModel) getAllMovieTMDBIDs() (map[int64]string, error) {
	filter := bson.M{"tmdb_id": bson.M{"$exists": true, "$ne": nil}}
	projection := bson.M{"_id": 1, "tmdb_id": 1}

	cursor, err := t.MovieCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	movieTMDBIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var movie struct {
			ID     primitive.ObjectID `bson:"_id"`
			TMDBID int64              `bson:"tmdb_id"`
		}
		if err := cursor.Decode(&movie); err != nil {
			continue
		}
		movieTMDBIDs[movie.TMDBID] = movie.ID.Hex()
	}

	return movieTMDBIDs, nil
}

func (t *TMDBImportModel) getAllTVTMDBIDs() (map[int64]string, error) {
	filter := bson.M{"tmdb_id": bson.M{"$exists": true, "$ne": nil}}
	projection := bson.M{"_id": 1, "tmdb_id": 1}

	cursor, err := t.TVCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	tvTMDBIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var tv struct {
			ID     primitive.ObjectID `bson:"_id"`
			TMDBID int64              `bson:"tmdb_id"`
		}
		if err := cursor.Decode(&tv); err != nil {
			continue
		}
		tvTMDBIDs[tv.TMDBID] = tv.ID.Hex()
	}

	return tvTMDBIDs, nil
}
