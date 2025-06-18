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

type TraktImportModel struct {
	MovieCollection     *mongo.Collection
	TVCollection        *mongo.Collection
	MovieListCollection *mongo.Collection
	TVListCollection    *mongo.Collection
}

func NewTraktImportModel(mongoDB *db.MongoDB) *TraktImportModel {
	return &TraktImportModel{
		MovieCollection:     mongoDB.Database.Collection("movies"),
		TVCollection:        mongoDB.Database.Collection("tv-series"),
		MovieListCollection: mongoDB.Database.Collection("movie-watch-lists"),
		TVListCollection:    mongoDB.Database.Collection("tvseries-watch-lists"),
	}
}

func (t *TraktImportModel) ImportUserData(userID, traktUsername string) (responses.TraktImportResponse, error) {
	clientID := os.Getenv("TRAKT_CLIENT_ID")
	if clientID == "" {
		return responses.TraktImportResponse{}, fmt.Errorf("Trakt client ID not configured")
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"trakt_username": traktUsername,
	}).Info("Starting Trakt import")

	// Get existing entries in bulk
	existingMovieEntries, err := t.getExistingMovieEntries(userID)
	if err != nil {
		return responses.TraktImportResponse{}, fmt.Errorf("failed to get existing movie entries: %v", err)
	}

	existingTVEntries, err := t.getExistingTVEntries(userID)
	if err != nil {
		return responses.TraktImportResponse{}, fmt.Errorf("failed to get existing TV entries: %v", err)
	}

	// Get all Trakt/IMDB IDs from database
	movieTraktIDs, err := t.getAllMovieTraktIDs()
	if err != nil {
		return responses.TraktImportResponse{}, fmt.Errorf("failed to get movie Trakt IDs: %v", err)
	}

	movieIMDBIDs, err := t.getAllMovieIMDBIDs()
	if err != nil {
		return responses.TraktImportResponse{}, fmt.Errorf("failed to get movie IMDB IDs: %v", err)
	}

	tvTraktIDs, err := t.getAllTVTraktIDs()
	if err != nil {
		return responses.TraktImportResponse{}, fmt.Errorf("failed to get TV Trakt IDs: %v", err)
	}

	tvIMDBIDs, err := t.getAllTVIMDBIDs()
	if err != nil {
		return responses.TraktImportResponse{}, fmt.Errorf("failed to get TV IMDB IDs: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":           userID,
		"movie_trakt_count": len(movieTraktIDs),
		"movie_imdb_count":  len(movieIMDBIDs),
		"tv_trakt_count":    len(tvTraktIDs),
		"tv_imdb_count":     len(tvIMDBIDs),
	}).Info("loaded Trakt mappings from database")

	var importedCount, skippedCount, errorCount int
	var movieEntriesToInsert []interface{}
	var tvEntriesToInsert []interface{}
	var importedTitles, skippedTitles []string

	// Import watched movies
	watchedMovies, err := t.fetchWatchedMovies(traktUsername, clientID)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch watched movies, continuing with other data")
	} else {
		t.processWatchedMovies(watchedMovies, userID, movieTraktIDs, movieIMDBIDs, existingMovieEntries,
			&movieEntriesToInsert, &importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Import watched TV shows
	watchedShows, err := t.fetchWatchedShows(traktUsername, clientID)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch watched shows, continuing with other data")
	} else {
		t.processWatchedShows(watchedShows, userID, tvTraktIDs, tvIMDBIDs, existingTVEntries,
			&tvEntriesToInsert, &importedCount, &skippedCount, &errorCount, &importedTitles, &skippedTitles)
	}

	// Import watchlist
	watchlistItems, err := t.fetchWatchlist(traktUsername, clientID)
	if err != nil {
		logrus.WithError(err).Warn("failed to fetch watchlist, continuing with other data")
	} else {
		t.processWatchlistItems(watchlistItems, userID, movieTraktIDs, movieIMDBIDs, tvTraktIDs, tvIMDBIDs,
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
			return responses.TraktImportResponse{}, fmt.Errorf("failed to import movie entries: %v", err)
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
			return responses.TraktImportResponse{}, fmt.Errorf("failed to import TV entries: %v", err)
		}
	}

	message := fmt.Sprintf("Import completed: %d imported, %d skipped, %d errors",
		importedCount, skippedCount, errorCount)

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
	}).Info("Trakt import completed")

	return responses.TraktImportResponse{
		ImportedCount:  importedCount,
		SkippedCount:   skippedCount,
		ErrorCount:     errorCount,
		Message:        message,
		ImportedTitles: importedTitles,
		SkippedTitles:  skippedTitles,
	}, nil
}

func (t *TraktImportModel) fetchWatchedMovies(username, clientID string) ([]responses.TraktWatchedMovie, error) {
	url := fmt.Sprintf("https://api.trakt.tv/users/%s/watched/movies", username)

	body, err := t.makeTraktRequest(url, clientID)
	if err != nil {
		return nil, err
	}

	var movies []responses.TraktWatchedMovie
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, fmt.Errorf("failed to parse watched movies response: %v", err)
	}

	return movies, nil
}

func (t *TraktImportModel) fetchWatchedShows(username, clientID string) ([]responses.TraktWatchedShow, error) {
	url := fmt.Sprintf("https://api.trakt.tv/users/%s/watched/shows", username)

	body, err := t.makeTraktRequest(url, clientID)
	if err != nil {
		return nil, err
	}

	var shows []responses.TraktWatchedShow
	if err := json.Unmarshal(body, &shows); err != nil {
		return nil, fmt.Errorf("failed to parse watched shows response: %v", err)
	}

	return shows, nil
}

func (t *TraktImportModel) fetchWatchlist(username, clientID string) ([]responses.TraktWatchlistItem, error) {
	url := fmt.Sprintf("https://api.trakt.tv/users/%s/watchlist", username)

	body, err := t.makeTraktRequest(url, clientID)
	if err != nil {
		return nil, err
	}

	var items []responses.TraktWatchlistItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("failed to parse watchlist response: %v", err)
	}

	return items, nil
}

func (t *TraktImportModel) makeTraktRequest(url, clientID string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", clientID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user not found or profile is private")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Trakt API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

func (t *TraktImportModel) processWatchedMovies(movies []responses.TraktWatchedMovie, userID string,
	movieTraktIDs map[int64]string, movieIMDBIDs map[string]string, existingMovieEntries map[string]bool,
	movieEntriesToInsert *[]interface{}, importedCount, skippedCount, errorCount *int,
	importedTitles, skippedTitles *[]string) {

	for _, movie := range movies {
		logrus.WithFields(logrus.Fields{
			"trakt_id": movie.Movie.IDs.Trakt,
			"imdb_id":  movie.Movie.IDs.IMDB,
			"title":    movie.Movie.Title,
		}).Debug("processing Trakt watched movie")

		// Try to find movie by Trakt ID first, then IMDB ID
		var movieObjectID string
		var found bool

		if movieObjectID, found = movieTraktIDs[int64(movie.Movie.IDs.Trakt)]; !found {
			if movie.Movie.IDs.IMDB != "" {
				movieObjectID, found = movieIMDBIDs[movie.Movie.IDs.IMDB]
			}
		}

		if !found {
			logrus.WithFields(logrus.Fields{
				"trakt_id": movie.Movie.IDs.Trakt,
				"imdb_id":  movie.Movie.IDs.IMDB,
				"title":    movie.Movie.Title,
			}).Debug("movie not found in database, skipping")
			*errorCount++
			continue
		}

		// Check if already exists in user's list
		if _, exists := existingMovieEntries[movieObjectID]; exists {
			*skippedCount++
			*skippedTitles = append(*skippedTitles, movie.Movie.Title)
			continue
		}

		// Create movie list entry
		movieListEntry := MovieWatchList{
			UserID:        userID,
			MovieID:       movieObjectID,
			Status:        "finished",
			TimesFinished: movie.Plays,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		*movieEntriesToInsert = append(*movieEntriesToInsert, movieListEntry)
		*importedCount++
		*importedTitles = append(*importedTitles, movie.Movie.Title)
	}
}

func (t *TraktImportModel) processWatchedShows(shows []responses.TraktWatchedShow, userID string,
	tvTraktIDs map[int64]string, tvIMDBIDs map[string]string, existingTVEntries map[string]bool,
	tvEntriesToInsert *[]interface{}, importedCount, skippedCount, errorCount *int,
	importedTitles, skippedTitles *[]string) {

	for _, show := range shows {
		logrus.WithFields(logrus.Fields{
			"trakt_id": show.Show.IDs.Trakt,
			"imdb_id":  show.Show.IDs.IMDB,
			"title":    show.Show.Title,
		}).Debug("processing Trakt watched show")

		// Try to find show by Trakt ID first, then IMDB ID
		var tvObjectID string
		var found bool

		if tvObjectID, found = tvTraktIDs[int64(show.Show.IDs.Trakt)]; !found {
			if show.Show.IDs.IMDB != "" {
				tvObjectID, found = tvIMDBIDs[show.Show.IDs.IMDB]
			}
		}

		if !found {
			logrus.WithFields(logrus.Fields{
				"trakt_id": show.Show.IDs.Trakt,
				"imdb_id":  show.Show.IDs.IMDB,
				"title":    show.Show.Title,
			}).Debug("show not found in database, skipping")
			*errorCount++
			continue
		}

		// Check if already exists in user's list
		if _, exists := existingTVEntries[tvObjectID]; exists {
			*skippedCount++
			*skippedTitles = append(*skippedTitles, show.Show.Title)
			continue
		}

		// Calculate total episodes watched
		totalEpisodes := 0
		for _, season := range show.Seasons {
			totalEpisodes += len(season.Episodes)
		}

		// Create TV list entry
		tvListEntry := TVSeriesWatchList{
			UserID:          userID,
			TvID:            tvObjectID,
			Status:          "finished",
			WatchedEpisodes: totalEpisodes,
			TimesFinished:   show.Plays,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		*tvEntriesToInsert = append(*tvEntriesToInsert, tvListEntry)
		*importedCount++
		*importedTitles = append(*importedTitles, show.Show.Title)
	}
}

func (t *TraktImportModel) processWatchlistItems(items []responses.TraktWatchlistItem, userID string,
	movieTraktIDs map[int64]string, movieIMDBIDs map[string]string, tvTraktIDs map[int64]string, tvIMDBIDs map[string]string,
	existingMovieEntries, existingTVEntries map[string]bool,
	movieEntriesToInsert, tvEntriesToInsert *[]interface{},
	importedCount, skippedCount, errorCount *int, importedTitles, skippedTitles *[]string) {

	for _, item := range items {
		if item.Type == "movie" && item.Movie != nil {
			// Process movie watchlist item
			var movieObjectID string
			var found bool

			if movieObjectID, found = movieTraktIDs[int64(item.Movie.IDs.Trakt)]; !found {
				if item.Movie.IDs.IMDB != "" {
					movieObjectID, found = movieIMDBIDs[item.Movie.IDs.IMDB]
				}
			}

			if !found {
				*errorCount++
				continue
			}

			if _, exists := existingMovieEntries[movieObjectID]; exists {
				*skippedCount++
				*skippedTitles = append(*skippedTitles, item.Movie.Title)
				continue
			}

			movieListEntry := MovieWatchList{
				UserID:        userID,
				MovieID:       movieObjectID,
				Status:        "active",
				TimesFinished: 0,
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			}

			*movieEntriesToInsert = append(*movieEntriesToInsert, movieListEntry)
			*importedCount++
			*importedTitles = append(*importedTitles, item.Movie.Title)

		} else if item.Type == "show" && item.Show != nil {
			// Process show watchlist item
			var tvObjectID string
			var found bool

			if tvObjectID, found = tvTraktIDs[int64(item.Show.IDs.Trakt)]; !found {
				if item.Show.IDs.IMDB != "" {
					tvObjectID, found = tvIMDBIDs[item.Show.IDs.IMDB]
				}
			}

			if !found {
				*errorCount++
				continue
			}

			if _, exists := existingTVEntries[tvObjectID]; exists {
				*skippedCount++
				*skippedTitles = append(*skippedTitles, item.Show.Title)
				continue
			}

			tvListEntry := TVSeriesWatchList{
				UserID:          userID,
				TvID:            tvObjectID,
				Status:          "active",
				WatchedEpisodes: 0,
				TimesFinished:   0,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			}

			*tvEntriesToInsert = append(*tvEntriesToInsert, tvListEntry)
			*importedCount++
			*importedTitles = append(*importedTitles, item.Show.Title)
		}
	}
}

func (t *TraktImportModel) getExistingMovieEntries(userID string) (map[string]bool, error) {
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

func (t *TraktImportModel) getExistingTVEntries(userID string) (map[string]bool, error) {
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

func (t *TraktImportModel) getAllMovieTraktIDs() (map[int64]string, error) {
	filter := bson.M{"trakt_id": bson.M{"$exists": true, "$ne": nil}}
	projection := bson.M{"_id": 1, "trakt_id": 1}

	cursor, err := t.MovieCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	movieTraktIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var movie struct {
			ID      primitive.ObjectID `bson:"_id"`
			TraktID int64              `bson:"trakt_id"`
		}
		if err := cursor.Decode(&movie); err != nil {
			continue
		}
		movieTraktIDs[movie.TraktID] = movie.ID.Hex()
	}

	return movieTraktIDs, nil
}

func (t *TraktImportModel) getAllMovieIMDBIDs() (map[string]string, error) {
	filter := bson.M{"imdb_id": bson.M{"$exists": true, "$ne": ""}}
	projection := bson.M{"_id": 1, "imdb_id": 1}

	cursor, err := t.MovieCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	movieIMDBIDs := make(map[string]string)
	for cursor.Next(context.TODO()) {
		var movie struct {
			ID     primitive.ObjectID `bson:"_id"`
			IMDBID string             `bson:"imdb_id"`
		}
		if err := cursor.Decode(&movie); err != nil {
			continue
		}
		movieIMDBIDs[movie.IMDBID] = movie.ID.Hex()
	}

	return movieIMDBIDs, nil
}

func (t *TraktImportModel) getAllTVTraktIDs() (map[int64]string, error) {
	filter := bson.M{"trakt_id": bson.M{"$exists": true, "$ne": nil}}
	projection := bson.M{"_id": 1, "trakt_id": 1}

	cursor, err := t.TVCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	tvTraktIDs := make(map[int64]string)
	for cursor.Next(context.TODO()) {
		var tv struct {
			ID      primitive.ObjectID `bson:"_id"`
			TraktID int64              `bson:"trakt_id"`
		}
		if err := cursor.Decode(&tv); err != nil {
			continue
		}
		tvTraktIDs[tv.TraktID] = tv.ID.Hex()
	}

	return tvTraktIDs, nil
}

func (t *TraktImportModel) getAllTVIMDBIDs() (map[string]string, error) {
	filter := bson.M{"imdb_id": bson.M{"$exists": true, "$ne": ""}}
	projection := bson.M{"_id": 1, "imdb_id": 1}

	cursor, err := t.TVCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	tvIMDBIDs := make(map[string]string)
	for cursor.Next(context.TODO()) {
		var tv struct {
			ID     primitive.ObjectID `bson:"_id"`
			IMDBID string             `bson:"imdb_id"`
		}
		if err := cursor.Decode(&tv); err != nil {
			continue
		}
		tvIMDBIDs[tv.IMDBID] = tv.ID.Hex()
	}

	return tvIMDBIDs, nil
}
