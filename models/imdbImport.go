package models

import (
	"app/db"
	"app/responses"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type IMDBImportModel struct {
	MovieCollection     *mongo.Collection
	TVCollection        *mongo.Collection
	MovieListCollection *mongo.Collection
	TVListCollection    *mongo.Collection
}

func NewIMDBImportModel(mongoDB *db.MongoDB) *IMDBImportModel {
	return &IMDBImportModel{
		MovieCollection:     mongoDB.Database.Collection("movies"),
		TVCollection:        mongoDB.Database.Collection("tv-series"),
		MovieListCollection: mongoDB.Database.Collection("movie-watch-lists"),
		TVListCollection:    mongoDB.Database.Collection("tvseries-watch-lists"),
	}
}

func (i *IMDBImportModel) ImportUserWatchlist(userID, imdbUserID, imdbListID string) (responses.IMDBImportResponse, error) {
	// Determine URL based on input
	var targetURL string
	if imdbUserID != "" {
		targetURL = fmt.Sprintf("https://www.imdb.com/user/%s/watchlist", imdbUserID)
	} else if imdbListID != "" {
		targetURL = fmt.Sprintf("https://www.imdb.com/list/%s/", imdbListID)
	} else {
		return responses.IMDBImportResponse{}, fmt.Errorf("either IMDB User ID or List ID must be provided")
	}

	// Validate and fetch watchlist
	entries, err := i.fetchIMDBWatchlist(targetURL)
	if err != nil {
		return responses.IMDBImportResponse{}, err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"imdb_user_id":  imdbUserID,
		"imdb_list_id":  imdbListID,
		"total_entries": len(entries),
	}).Info("Starting IMDB import")

	// Get existing entries in bulk
	existingMovieEntries, err := i.getExistingMovieEntries(userID)
	if err != nil {
		return responses.IMDBImportResponse{}, fmt.Errorf("failed to get existing movie entries: %v", err)
	}

	existingTVEntries, err := i.getExistingTVEntries(userID)
	if err != nil {
		return responses.IMDBImportResponse{}, fmt.Errorf("failed to get existing TV entries: %v", err)
	}

	// Get all IMDB IDs from database
	movieIMDBIDs, err := i.getAllMovieIMDBIDs()
	if err != nil {
		return responses.IMDBImportResponse{}, fmt.Errorf("failed to get movie IMDB IDs: %v", err)
	}

	tvIMDBIDs, err := i.getAllTVIMDBIDs()
	if err != nil {
		return responses.IMDBImportResponse{}, fmt.Errorf("failed to get TV IMDB IDs: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":          userID,
		"movie_imdb_count": len(movieIMDBIDs),
		"tv_imdb_count":    len(tvIMDBIDs),
	}).Info("loaded IMDB mappings from database")

	var importedCount, skippedCount, errorCount int
	var movieEntriesToInsert []interface{}
	var tvEntriesToInsert []interface{}

	// Process entries
	for _, entry := range entries {
		logrus.WithFields(logrus.Fields{
			"imdb_id": entry.IMDBID,
			"title":   entry.Title,
			"type":    entry.Type,
		}).Debug("processing IMDB entry")

		// Check if it's a movie
		if movieObjectID, exists := movieIMDBIDs[entry.IMDBID]; exists {
			// Check if already exists in user's movie list
			if _, exists := existingMovieEntries[movieObjectID]; exists {
				// Update existing entry
				err := i.updateExistingMovieEntry(userID, movieObjectID, entry.UserRating, "planning")
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"imdb_id":  entry.IMDBID,
						"title":    entry.Title,
						"movie_id": movieObjectID,
						"error":    err.Error(),
					}).Error("failed to update existing movie entry")
					errorCount++
				} else {
					skippedCount++
				}
				continue
			}

			// Create new movie list entry
			var score *float32
			if entry.UserRating > 0 {
				score = &entry.UserRating
			}
			movieListEntry := MovieWatchList{
				UserID:    userID,
				MovieID:   movieObjectID,
				Status:    "planning",
				Score:     score,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			movieEntriesToInsert = append(movieEntriesToInsert, movieListEntry)
			importedCount++
		} else if tvObjectID, exists := tvIMDBIDs[entry.IMDBID]; exists {
			// Check if already exists in user's TV list
			if _, exists := existingTVEntries[tvObjectID]; exists {
				// Update existing entry
				err := i.updateExistingTVEntry(userID, tvObjectID, entry.UserRating, "planning")
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"imdb_id": entry.IMDBID,
						"title":   entry.Title,
						"tv_id":   tvObjectID,
						"error":   err.Error(),
					}).Error("failed to update existing TV entry")
					errorCount++
				} else {
					skippedCount++
				}
				continue
			}

			// Create new TV list entry
			var tvScore *float32
			if entry.UserRating > 0 {
				tvScore = &entry.UserRating
			}
			tvListEntry := TVSeriesWatchList{
				UserID:    userID,
				TvID:      tvObjectID,
				Status:    "planning",
				Score:     tvScore,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}

			tvEntriesToInsert = append(tvEntriesToInsert, tvListEntry)
			importedCount++
		} else {
			logrus.WithFields(logrus.Fields{
				"imdb_id": entry.IMDBID,
				"title":   entry.Title,
			}).Debug("content not found in database, skipping")
			errorCount++
		}
	}

	// Bulk insert movie entries
	if len(movieEntriesToInsert) > 0 {
		_, err = i.MovieListCollection.InsertMany(context.TODO(), movieEntriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(movieEntriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert movie entries")
			return responses.IMDBImportResponse{}, fmt.Errorf("failed to import movie entries: %v", err)
		}
	}

	// Bulk insert TV entries
	if len(tvEntriesToInsert) > 0 {
		_, err = i.TVListCollection.InsertMany(context.TODO(), tvEntriesToInsert)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userID,
				"count":   len(tvEntriesToInsert),
				"error":   err.Error(),
			}).Error("failed to bulk insert TV entries")
			return responses.IMDBImportResponse{}, fmt.Errorf("failed to import TV entries: %v", err)
		}
	}

	message := fmt.Sprintf("Import completed: %d imported, %d skipped, %d errors",
		importedCount, skippedCount, errorCount)

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
	}).Info("IMDB import completed")

	return responses.IMDBImportResponse{
		ImportedCount: importedCount,
		SkippedCount:  skippedCount,
		ErrorCount:    errorCount,
		Message:       message,
	}, nil
}

func (i *IMDBImportModel) fetchIMDBWatchlist(url string) ([]responses.IMDBEntry, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IMDB page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("IMDB list not found or is private")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("IMDB returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var entries []responses.IMDBEntry

	// Parse watchlist items
	doc.Find(".lister-item, .titleColumn, .cli-item").Each(func(index int, s *goquery.Selection) {
		entry := i.parseWatchlistItem(s)
		if entry.IMDBID != "" {
			entries = append(entries, entry)
		}
	})

	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries found in IMDB list or list is private")
	}

	return entries, nil
}

func (i *IMDBImportModel) parseWatchlistItem(s *goquery.Selection) responses.IMDBEntry {
	var entry responses.IMDBEntry

	// Extract IMDB ID from various possible selectors
	titleLink := s.Find("h3.titleColumn a, .titleColumn a, .cli-title a, a[href*='/title/']").First()
	if href, exists := titleLink.Attr("href"); exists {
		entry.IMDBID = i.extractIMDBIDFromURL(href)
	}

	// Extract title
	entry.Title = strings.TrimSpace(titleLink.Text())

	// Extract year
	yearText := s.Find(".secondaryInfo, .cli-title-metadata, .titleColumn .secondaryInfo").First().Text()
	if yearMatch := regexp.MustCompile(`\((\d{4})\)`).FindStringSubmatch(yearText); len(yearMatch) > 1 {
		if year, err := strconv.Atoi(yearMatch[1]); err == nil {
			entry.Year = year
		}
	}

	// Extract user rating if available
	ratingText := s.Find(".ipl-rating-star__rating, .ratingColumn strong, .cli-rating").First().Text()
	if ratingText != "" {
		if rating, err := strconv.ParseFloat(strings.TrimSpace(ratingText), 32); err == nil {
			entry.UserRating = float32(rating)
		}
	}

	// Determine type (movie vs TV series)
	typeText := s.Find(".genre, .cli-title-metadata").Text()
	if strings.Contains(strings.ToLower(typeText), "tv") || strings.Contains(strings.ToLower(typeText), "series") {
		entry.Type = "tv"
	} else {
		entry.Type = "movie"
	}

	return entry
}

func (i *IMDBImportModel) extractIMDBIDFromURL(url string) string {
	// Extract IMDB ID from URL like /title/tt0111161/ or /title/tt0111161/...
	re := regexp.MustCompile(`/title/(tt\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func (i *IMDBImportModel) getExistingMovieEntries(userID string) (map[string]bool, error) {
	filter := bson.M{"user_id": userID}
	projection := bson.M{"movie_id": 1}

	cursor, err := i.MovieListCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
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

func (i *IMDBImportModel) getExistingTVEntries(userID string) (map[string]bool, error) {
	filter := bson.M{"user_id": userID}
	projection := bson.M{"tv_id": 1}

	cursor, err := i.TVListCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
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

func (i *IMDBImportModel) getAllMovieIMDBIDs() (map[string]string, error) {
	filter := bson.M{"imdb_id": bson.M{"$exists": true, "$ne": ""}}
	projection := bson.M{"_id": 1, "imdb_id": 1}

	cursor, err := i.MovieCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
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

func (i *IMDBImportModel) getAllTVIMDBIDs() (map[string]string, error) {
	filter := bson.M{"imdb_id": bson.M{"$exists": true, "$ne": ""}}
	projection := bson.M{"_id": 1, "imdb_id": 1}

	cursor, err := i.TVCollection.Find(context.TODO(), filter, options.Find().SetProjection(projection))
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

func (i *IMDBImportModel) updateExistingMovieEntry(userID, movieID string, rating float32, status string) error {
	filter := bson.M{
		"user_id":  userID,
		"movie_id": movieID,
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
	}

	if rating > 0 {
		update["$set"].(bson.M)["score"] = &rating
	}

	_, err := i.MovieListCollection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (i *IMDBImportModel) updateExistingTVEntry(userID, tvID string, rating float32, status string) error {
	filter := bson.M{
		"user_id": userID,
		"tv_id":   tvID,
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
	}

	if rating > 0 {
		update["$set"].(bson.M)["score"] = &rating
	}

	_, err := i.TVListCollection.UpdateOne(context.TODO(), filter, update)
	return err
}
