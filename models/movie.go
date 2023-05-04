package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"
	"strconv"

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
	moviePaginationLimit         = 20
)

/* TODO Endpoints
* [x] Get upcoming movies by popularity etc.
* [x] Get movies by release date, popularity, genre etc. (sort & filter)
* [ ] Get movie details
* [x] Get top movies by every decade 1980's 1990's etc.
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

func (movieModel *MovieModel) GetMoviesBySortAndFilter(data requests.SortFilterMovie) ([]responses.Movie, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "tmdb_popularity"
		sortOrder = -1
	case "new":
		sortType = "release_date"
		sortOrder = -1
	case "old":
		sortType = "release_date"
		sortOrder = 1
	}

	match := bson.M{}
	if data.Status != nil || data.Genres != nil || data.ProductionCompanies != nil ||
		data.ReleaseDateFrom != nil || data.ReleaseDateTo != nil {

		if data.Status != nil {
			switch *data.Status {
			case "production":
				match["$or"] = bson.A{
					bson.M{
						"status": "Post Production",
					},
					bson.M{
						"status": "In Production",
					},
				}
			case "released":
				match["status"] = "Released"
			case "planned":
				match["status"] = "Planned"
			}
		}

		if data.Genres != nil {
			match["genres.name"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.ProductionCompanies != nil {
			match["production_companies.name"] = bson.M{
				"$in": bson.A{data.ProductionCompanies},
			}
		}

		if data.ReleaseDateFrom != nil {
			if data.ReleaseDateTo != nil {
				match["release_date"] = bson.M{
					"$gte": strconv.Itoa(*data.ReleaseDateFrom),
					"$lt":  strconv.Itoa(*data.ReleaseDateTo),
				}
			} else {
				match["release_date"] = bson.M{
					"$gte": strconv.Itoa(*data.ReleaseDateFrom),
				}
			}
		}
	}

	var movies []responses.Movie
	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(moviePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&movies).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate movies by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get movies by selected filters.")
	}

	return movies, paginatedData.Pagination, nil
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

func (movieModel *MovieModel) GetMovieDetailsWithWatchList(data requests.ID, uuid string) (responses.MovieDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	match := bson.M{"$match": bson.M{
		"_id": objectID,
	}}

	set := bson.M{"$set": bson.M{
		"movie_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from": "movie-watch-lists",
		"let": bson.M{
			"uuid":     uuid,
			"movie_id": "$movie_id",
			"tmdb_id":  "$tmdb_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$movie_id", "$$movie_id"}},
									bson.M{"$eq": bson.A{"$movie_tmdb_id", "$$tmdb_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
		},
		"as": "watch_list",
	}}

	unwindWatchList := bson.M{"$unwind": bson.M{
		"path":                       "$watch_list",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate movie details: ", err)

		return responses.MovieDetails{}, fmt.Errorf("Failed to aggregate movie details with watch list.")
	}

	var movieDetails []responses.MovieDetails
	if err = cursor.All(context.TODO(), &movieDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to decode movie details: ", err)

		return responses.MovieDetails{}, fmt.Errorf("Failed to decode movie details.")
	}

	if len(movieDetails) > 0 {
		return movieDetails[0], nil
	}

	return responses.MovieDetails{}, nil
}
