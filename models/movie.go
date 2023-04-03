package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MovieModel struct {
	Collection *mongo.Collection
}

func NewMovieModel(mongoDB *db.MongoDB) *MovieModel {
	return &MovieModel{
		Collection: mongoDB.Database.Collection("movies"),
	}
}

const (
	movieUpcomingPaginationLimit = 20
)

/* TODO Endpoints
* [x] Get upcoming movies by popularity etc.
* [ ] Get movies by release date, popularity, genre etc. (sort & filter)
* [ ] Get movie details
* [ ] Get top movies by every decade 1980's 1990's etc.
* [ ] Get top movies by every genre (?)
 */

func (movieModel *MovieModel) GetUpcomingMoviesBySort(data requests.SortUpcoming) ([]responses.Movie, p.PaginationData, error) {
	var (
		sortType            string
		sortOrder           int8
		hasReleaseDateOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "tmdb_popularity"
		sortOrder = -1
		hasReleaseDateOrder = -1
	case "soon":
		sortType = "release_date"
		sortOrder = 1
		hasReleaseDateOrder = -1
	case "later":
		sortType = "release_date"
		sortOrder = -1
		hasReleaseDateOrder = 1
	}

	match := bson.M{"$match": bson.M{
		"status": bson.M{
			"$ne": "Released",
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_release_date": bson.M{
			"$ne": bson.A{"$release_date", ""},
		},
	}}

	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(movieUpcomingPaginationLimit).
		Page(data.Page).Sort("has_release_date", hasReleaseDateOrder).Sort(sortType, sortOrder).Sort("_id", 1).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming movies: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming movies.")
	}

	var upcomingMovies []responses.Movie
	for _, raw := range paginatedData.Data {
		var movie *responses.Movie
		if marshalErr := bson.Unmarshal(raw, &movie); marshalErr == nil {
			upcomingMovies = append(upcomingMovies, *movie)
		}
	}

	return upcomingMovies, paginatedData.Pagination, nil
}

func (movieModel *MovieModel) GetMovieDetails(data requests.ID) (responses.Movie, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	result := movieModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectID,
	})

	var movie responses.Movie
	if err := result.Decode(&movie); err != nil {
		logrus.WithFields(logrus.Fields{
			"game_id": data.ID,
		}).Error("failed to find movie details by id: ", err)

		return responses.Movie{}, fmt.Errorf("Failed to find movie by id.")
	}

	return movie, nil
}
