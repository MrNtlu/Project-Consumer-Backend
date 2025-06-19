package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"app/utils"
	"context"
	"fmt"
	"strconv"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type MovieModel struct {
	Collection *mongo.Collection
}

func NewMovieModel(mongoDB *db.MongoDB) *MovieModel {
	return &MovieModel{
		Collection: mongoDB.Database.Collection("movies"),
	}
}

const (
	movieUpcomingPaginationLimit = 40
	movieSearchLimit             = 50
	moviePaginationLimit         = 40
	movieActorsLimit             = 50
	popularPlatformsLimit        = 15
	PreviewLimit                 = 30
)

func (movieModel *MovieModel) GetUpcomingPreviewMovies() ([]responses.PreviewMovie, error) {
	match := bson.M{
		"$or": bson.A{
			bson.M{
				"$and": bson.A{
					bson.M{
						"status": bson.M{
							"$ne": "Released",
						},
					},
					bson.M{"release_date": bson.M{"$gte": utils.GetCurrentDate()}},
				},
			},
			bson.M{
				"$and": bson.A{
					bson.M{
						"status": bson.M{
							"$ne": "Released",
						},
					},
					bson.M{"release_date": ""},
				},
			},
		},
	}

	// Only fetch required fields for preview
	opts := options.Find().
		SetSort(bson.M{"tmdb_popularity": -1}).
		SetLimit(PreviewLimit).
		SetProjection(bson.M{
			"_id":            1,
			"tmdb_id":        1,
			"title_en":       1,
			"title_original": 1,
			"image_url":      1,
		})

	cursor, err := movieModel.Collection.Find(context.TODO(), match, opts)
	if err != nil {
		logrus.Error("failed to find preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to find preview upcoming movies.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewMovie, 0, PreviewLimit)
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to decode preview upcoming movies.")
	}

	return results, nil
}

func (movieModel *MovieModel) GetPopularPreviewMovies() ([]responses.PreviewMovie, error) {
	filter := bson.D{}

	// Only fetch required fields for preview
	opts := options.Find().
		SetSort(bson.M{"tmdb_popularity": -1}).
		SetLimit(PreviewLimit).
		SetProjection(bson.M{
			"_id":            1,
			"tmdb_id":        1,
			"title_en":       1,
			"title_original": 1,
			"image_url":      1,
		})

	cursor, err := movieModel.Collection.Find(context.TODO(), filter, opts)
	if err != nil {
		logrus.Error("failed to find popular upcoming: ", err)

		return nil, fmt.Errorf("Failed to find popular movies.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewMovie, 0, PreviewLimit)
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode popular upcoming: ", err)

		return nil, fmt.Errorf("Failed to decode popular movies.")
	}

	return results, nil
}

func (movieModel *MovieModel) GetTopPreviewMovies() ([]responses.PreviewMovie, error) {
	addFields := bson.M{"$addFields": bson.M{
		"top_rated": bson.M{
			"$multiply": bson.A{
				"$tmdb_vote", "$tmdb_vote_count",
			},
		},
	}}

	// Only project required fields for preview
	project := bson.M{"$project": bson.M{
		"_id":            1,
		"tmdb_id":        1,
		"title_en":       1,
		"title_original": 1,
		"image_url":      1,
		"top_rated":      1,
	}}

	sort := bson.M{"$sort": bson.M{
		"top_rated": -1,
	}}

	limit := bson.M{"$limit": PreviewLimit}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		addFields, project, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate top preview movies: ", err)

		return nil, fmt.Errorf("Failed to aggregate top preview movies.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewMovie, 0, PreviewLimit)
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode top movies: ", err)

		return nil, fmt.Errorf("Failed to decode top movies.")
	}

	return results, nil
}

func (movieModel *MovieModel) GetInTheaterPreviewMovies() ([]responses.PreviewMovie, error) {
	match := bson.M{
		"status": "Released",
		"release_date": bson.M{
			"$gte": utils.GetCustomDate(0, -1, 0),
			"$lte": utils.GetCurrentDate(),
		},
	}

	// Only fetch required fields for preview
	opts := options.Find().
		SetSort(bson.M{"tmdb_popularity": -1}).
		SetLimit(PreviewLimit).
		SetProjection(bson.M{
			"_id":            1,
			"tmdb_id":        1,
			"title_en":       1,
			"title_original": 1,
			"image_url":      1,
		})

	cursor, err := movieModel.Collection.Find(context.TODO(), match, opts)
	if err != nil {
		logrus.Error("failed to find preview in theater movies: ", err)

		return nil, fmt.Errorf("Failed to find preview in theater movies.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewMovie, 0, PreviewLimit)
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview in theater movies: ", err)

		return nil, fmt.Errorf("Failed to decode preview in theater movies.")
	}

	return results, nil
}

func (movieModel *MovieModel) GetMoviesFromOpenAI(uid string, movieIDs []string, limitValue int) ([]responses.AISuggestion, error) {
	match := bson.M{"$match": bson.M{
		"$expr": bson.M{
			"$in": bson.A{bson.M{"$toString": "$_id"}, movieIDs},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"tmdb_popularity": -1,
	}}

	limit := bson.M{"$limit": limitValue}

	set := bson.M{"$set": bson.M{
		"movie_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
		"let": bson.M{
			"uid":      uid,
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
									bson.M{"$eq": bson.A{"$content_id", "$$movie_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$tmdb_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
						},
					},
				},
			},
		},
		"as": "watch_later",
	}}

	unwindWatchLater := bson.M{"$unwind": bson.M{
		"path":                       "$watch_later",
		"preserveNullAndEmptyArrays": true,
	}}

	lookupNotInterested := bson.M{"$lookup": bson.M{
		"from": "ai-suggestions-not-interested",
		"let": bson.M{
			"uid":      uid,
			"movie_id": "$movie_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$content_id", "$$movie_id"}},
							bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
						},
					},
				},
			},
		},
		"as": "not_interested",
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toString": "$_id",
		},
		"content_external_id": "$tmdb_id",
		"content_type":        "movie",
		"title_en":            1,
		"title_original":      1,
		"description":         1,
		"image_url":           1,
		"score":               "$tmdb_vote",
		"watch_later":         1,
		"not_interested": bson.M{
			"$cond": bson.M{
				"if":   bson.M{"$gt": bson.A{bson.M{"$size": "$not_interested"}, 0}},
				"then": true,
				"else": false,
			},
		},
	}}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		match, sort, limit, set, lookupWatchLater, unwindWatchLater,
		lookupNotInterested, project,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"movieIDs": movieIDs,
		}).Error("failed to aggregate movies: ", err)

		return nil, fmt.Errorf("Failed to get movie from recommendation.")
	}

	var movieList []responses.AISuggestion
	if err := cursor.All(context.TODO(), &movieList); err != nil {
		logrus.WithFields(logrus.Fields{
			"movieIDs": movieIDs,
		}).Error("failed to decode movies: ", err)

		return nil, fmt.Errorf("Failed to decode get movie from recommendation.")
	}

	return movieList, nil
}

func (movieModel *MovieModel) GetMoviesInTheater(data requests.Pagination) ([]responses.Movie, p.PaginationData, error) {
	match := bson.M{
		"status": "Released",
		"release_date": bson.M{
			"$gte": utils.GetCustomDate(0, -1, 0),
			"$lte": utils.GetCurrentDate(),
		},
	}

	var movies []responses.Movie
	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(moviePaginationLimit).
		Page(data.Page).Sort("tmdb_popularity", -1).Filter(match).Decode(&movies).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate movies in theater: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get movies in theater.")
	}

	return movies, paginatedData.Pagination, nil
}

func (movieModel *MovieModel) GetUpcomingMoviesBySort(data requests.Pagination) ([]responses.Movie, p.PaginationData, error) {
	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"$and": bson.A{
					bson.M{
						"status": bson.M{
							"$ne": "Released",
						},
					},
					bson.M{"release_date": bson.M{"$gte": utils.GetCurrentDate()}},
				},
			},
			bson.M{
				"$and": bson.A{
					bson.M{
						"status": bson.M{
							"$ne": "Released",
						},
					},
					bson.M{"release_date": ""},
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"has_release_date": bson.M{
			"$ne": bson.A{"$release_date", ""},
		},
		"tmdb_id":         1,
		"image_url":       1,
		"imdb_id":         1,
		"length":          1,
		"release_date":    1,
		"status":          1,
		"title_en":        1,
		"title_original":  1,
		"description":     1,
		"tmdb_vote":       1,
		"tmdb_vote_count": 1,
		"tmdb_popularity": 1,
	}}

	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(movieUpcomingPaginationLimit).
		Page(data.Page).Sort("has_release_date", -1).Sort("tmdb_popularity", -1).Sort("_id", 1).Aggregate(match, project)
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
	project := bson.M{"$project": bson.M{
		"top_rated": bson.M{
			"$multiply": bson.A{
				"$tmdb_vote", "$tmdb_vote_count",
			},
		},
		"tmdb_id":         1,
		"image_url":       1,
		"imdb_id":         1,
		"length":          1,
		"release_date":    1,
		"status":          1,
		"title_en":        1,
		"title_original":  1,
		"description":     1,
		"tmdb_vote":       1,
		"tmdb_vote_count": 1,
		"tmdb_popularity": 1,
	}}

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "top":
		sortType = "top_rated"
		sortOrder = -1
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

	matchFields := bson.M{}
	if data.Status != nil || data.Genres != nil || data.ProductionCompanies != nil ||
		data.ReleaseDateFrom != nil || data.ReleaseDateTo != nil ||
		data.ProductionCountry != nil || data.StreamingPlatforms != nil ||
		data.Rating != nil {

		if data.Status != nil {
			switch *data.Status {
			case "production":
				matchFields["$or"] = bson.A{
					bson.M{
						"status": "Post Production",
					},
					bson.M{
						"status": "In Production",
					},
				}
			case "released":
				matchFields["status"] = "Released"
			case "planned":
				matchFields["status"] = "Planned"
			}
		}

		if data.Genres != nil {
			matchFields["genres"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.ProductionCompanies != nil {
			matchFields["production_companies.name"] = bson.M{
				"$in": bson.A{data.ProductionCompanies},
			}
		}

		if data.ProductionCountry != nil {
			matchFields["production_companies.origin_country"] = bson.M{
				"$in": bson.A{data.ProductionCountry},
			}
		}

		if data.StreamingPlatforms != nil && (data.IsStreamingPlatformFiltered == nil || (data.IsStreamingPlatformFiltered != nil && !*data.IsStreamingPlatformFiltered)) {
			matchFields["streaming.streaming_platforms.name"] = bson.M{
				"$in": bson.A{data.StreamingPlatforms},
			}
		} else if data.StreamingPlatforms != nil && data.IsStreamingPlatformFiltered != nil && *data.IsStreamingPlatformFiltered && data.Region != nil {
			matchFields["$and"] = bson.A{
				bson.M{"streaming.streaming_platforms.name": bson.M{"$in": bson.A{data.StreamingPlatforms}}},
				bson.M{"streaming.country_code": data.Region},
			}
		}

		if data.Rating != nil {
			matchFields["$and"] = bson.A{
				bson.M{
					"tmdb_vote": bson.M{"$gte": data.Rating},
				},
				bson.M{
					"tmdb_vote_count": bson.M{"$gte": 100},
				},
			}
		}

		if data.ReleaseDateFrom != nil {
			if data.ReleaseDateTo != nil {
				matchFields["release_date"] = bson.M{
					"$gte": strconv.Itoa(*data.ReleaseDateFrom),
					"$lt":  strconv.Itoa(*data.ReleaseDateTo),
				}
			} else {
				matchFields["release_date"] = bson.M{
					"$gte": strconv.Itoa(*data.ReleaseDateFrom),
				}
			}
		}
	}

	match := bson.M{"$match": matchFields}

	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(moviePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(match, project)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
			"match":   match,
		}).Error("failed to aggregate movies by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get movies by selected filters.")
	}

	var movies []responses.Movie
	for _, raw := range paginatedData.Data {
		var movie *responses.Movie
		if marshalErr := bson.Unmarshal(raw, &movie); marshalErr == nil {
			movies = append(movies, *movie)
		}
	}

	return movies, paginatedData.Pagination, nil
}

func (movieModel *MovieModel) GetMovieDetails(data requests.ID) (responses.Movie, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	// Use FindOne with projection to only fetch required fields
	projection := bson.M{
		"_id":                  1,
		"title_en":             1,
		"title_original":       1,
		"description":          1,
		"image_url":            1,
		"status":               1,
		"length":               1,
		"imdb_id":              1,
		"tmdb_id":              1,
		"tmdb_popularity":      1,
		"tmdb_vote":            1,
		"tmdb_vote_count":      1,
		"release_date":         1,
		"backdrop":             1,
		"recommendations":      1,
		"production_companies": 1,
		"genres":               1,
		"images":               1,
		"videos":               1,
		"streaming":            1,
		"actors":               1,
	}

	options := options.FindOne().SetProjection(projection)

	result := movieModel.Collection.FindOne(context.TODO(), bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"tmdb_id": data.ID,
			},
		},
	}, options)

	var movie responses.Movie
	if err := result.Decode(&movie); err != nil {
		logrus.WithFields(logrus.Fields{
			"movie_id": data.ID,
		}).Error("failed to find movie details by id: ", err)

		return responses.Movie{}, fmt.Errorf("Failed to find movie by id.")
	}

	return movie, nil
}

func (movieModel *MovieModel) GetMovieDetailsWithWatchListAndWatchLater(data requests.ID, uuid string) (responses.MovieDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"tmdb_id": data.ID,
			},
		},
	}}

	// Add projection early to reduce data transfer
	project := bson.M{"$project": bson.M{
		"_id":                  1,
		"title_en":             1,
		"title_original":       1,
		"description":          1,
		"image_url":            1,
		"status":               1,
		"length":               1,
		"imdb_id":              1,
		"tmdb_id":              1,
		"tmdb_popularity":      1,
		"tmdb_vote":            1,
		"tmdb_vote_count":      1,
		"release_date":         1,
		"backdrop":             1,
		"recommendations":      1,
		"production_companies": 1,
		"genres":               1,
		"images":               1,
		"videos":               1,
		"streaming":            1,
		"actors":               1,
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
			// Project only needed fields from watch list
			bson.M{
				"$project": bson.M{
					"status":         1,
					"score":          1,
					"times_finished": 1,
					"created_at":     1,
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

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
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
									bson.M{"$eq": bson.A{"$content_id", "$$movie_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$tmdb_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
			// Project only needed fields from watch later
			bson.M{
				"$project": bson.M{
					"created_at": 1,
				},
			},
		},
		"as": "watch_later",
	}}

	unwindWatchLater := bson.M{"$unwind": bson.M{
		"path":                       "$watch_later",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		match, project, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate movie details: ", err)

		return responses.MovieDetails{}, fmt.Errorf("Failed to aggregate movie details with watch list.")
	}

	// Pre-allocate slice with capacity 1 since we expect only one result
	movieDetails := make([]responses.MovieDetails, 0, 1)
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

func (movieModel *MovieModel) GetPopularActors(data requests.Pagination) ([]responses.ActorDetails, error) {
	allowDiskPreventSet := bson.M{"$set": bson.M{
		"actors": bson.M{
			"$slice": bson.A{"$actors", 15},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$actors",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$actors.tmdb_id",
		"name": bson.M{
			"$first": "$actors.name",
		},
		"image_url": bson.M{
			"$first": "$actors.image",
		},
		"count": bson.M{
			"$sum": 1,
		},
		"popularity": bson.M{
			"$sum": "$tmdb_popularity",
		},
	}}

	set := bson.M{"$set": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{"$count", "$popularity"},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"popularity": -1,
	}}

	limit := bson.M{"$limit": movieActorsLimit}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		allowDiskPreventSet, unwind, group, set, sort, limit,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to aggregate actors: ", err)

		return nil, fmt.Errorf("Failed to get top actors.")
	}

	var actorList []responses.ActorDetails
	if err := cursor.All(context.TODO(), &actorList); err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to decode actors: ", err)

		return nil, fmt.Errorf("Failed to decode top actors.")
	}

	return actorList, nil
}

// GetCombinedPopularActors - OPTIMIZATION: Combined movie+TV actors in single query
// This replaces separate GetPopularActors calls for movies and TV to reduce database operations
func (movieModel *MovieModel) GetCombinedPopularActors(data requests.Pagination) ([]responses.ActorDetails, []responses.ActorDetails, error) {
	// Use facet to get both movie and TV actors in single aggregation
	facet := bson.M{"$facet": bson.M{
		"movieActors": bson.A{
			bson.M{"$set": bson.M{
				"actors": bson.M{
					"$slice": bson.A{"$actors", 15},
				},
			}},
			bson.M{"$unwind": bson.M{
				"path":                       "$actors",
				"includeArrayIndex":          "index",
				"preserveNullAndEmptyArrays": false,
			}},
			bson.M{"$group": bson.M{
				"_id": "$actors.tmdb_id",
				"name": bson.M{
					"$first": "$actors.name",
				},
				"image_url": bson.M{
					"$first": "$actors.image",
				},
				"count": bson.M{
					"$sum": 1,
				},
				"popularity": bson.M{
					"$sum": "$tmdb_popularity",
				},
			}},
			bson.M{"$set": bson.M{
				"popularity": bson.M{
					"$multiply": bson.A{"$count", "$popularity"},
				},
			}},
			bson.M{"$sort": bson.M{
				"popularity": -1,
			}},
			bson.M{"$limit": movieActorsLimit},
		},
		"tvActors": bson.A{
			bson.M{"$lookup": bson.M{
				"from": "tv-series",
				"pipeline": bson.A{
					bson.M{"$set": bson.M{
						"actors": bson.M{
							"$slice": bson.A{"$actors", 15},
						},
					}},
					bson.M{"$unwind": bson.M{
						"path":                       "$actors",
						"includeArrayIndex":          "index",
						"preserveNullAndEmptyArrays": false,
					}},
					bson.M{"$group": bson.M{
						"_id": "$actors.tmdb_id",
						"name": bson.M{
							"$first": "$actors.name",
						},
						"image_url": bson.M{
							"$first": "$actors.image",
						},
						"count": bson.M{
							"$sum": 1,
						},
						"popularity": bson.M{
							"$sum": "$tmdb_popularity",
						},
					}},
					bson.M{"$set": bson.M{
						"popularity": bson.M{
							"$multiply": bson.A{"$count", "$popularity"},
						},
					}},
					bson.M{"$sort": bson.M{
						"popularity": -1,
					}},
					bson.M{"$limit": movieActorsLimit},
				},
				"as": "tvActors",
			}},
			bson.M{"$unwind": "$tvActors"},
			bson.M{"$replaceRoot": bson.M{
				"newRoot": "$tvActors",
			}},
		},
	}}

	// Limit to single document for facet operation
	match := bson.M{"$match": bson.M{
		"_id": bson.M{"$exists": true},
	}}

	limitOne := bson.M{"$limit": 1}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		match, limitOne, facet,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to aggregate combined actors: ", err)

		return nil, nil, fmt.Errorf("Failed to get combined actors.")
	}

	var results []struct {
		MovieActors []responses.ActorDetails `bson:"movieActors"`
		TVActors    []responses.ActorDetails `bson:"tvActors"`
	}

	if err := cursor.All(context.TODO(), &results); err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to decode combined actors: ", err)

		return nil, nil, fmt.Errorf("Failed to decode combined actors.")
	}

	if len(results) > 0 {
		return results[0].MovieActors, results[0].TVActors, nil
	}

	return []responses.ActorDetails{}, []responses.ActorDetails{}, nil
}

func (movieModel *MovieModel) GetMoviesByActor(data requests.IDPagination) ([]responses.Movie, p.PaginationData, error) {
	match := bson.M{
		"actors.tmdb_id": bson.M{
			"$in": bson.A{data.ID},
		},
	}

	var movies []responses.Movie
	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(moviePaginationLimit).
		Page(data.Page).Sort("tmdb_popularity", -1).Filter(match).Decode(&movies).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate movies by actor: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get movies by actor.")
	}

	return movies, paginatedData.Pagination, nil
}

func (movieModel *MovieModel) GetPopularStreamingPlatforms(region string) ([]responses.StreamingPlatform, error) {
	match := bson.M{"$match": bson.M{
		"streaming.country_code": region,
	}}

	project := bson.M{"$project": bson.M{
		"_id": 1,
		"streaming": bson.M{
			"$slice": bson.A{
				bson.M{
					"$filter": bson.M{
						"input": "$streaming",
						"cond": bson.M{
							"$eq": bson.A{
								"$$this.country_code", region,
							},
						},
					},
				},
				1,
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$streaming",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	set := bson.M{"$set": bson.M{
		"streaming": "$streaming.streaming_platforms",
	}}

	unwindAgain := bson.M{"$unwind": bson.M{
		"path":                       "$streaming",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$streaming.name",
		"name": bson.M{
			"$first": "$streaming.name",
		},
		"logo": bson.M{
			"$first": "$streaming.logo",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"count": -1,
	}}

	limit := bson.M{"$limit": popularPlatformsLimit}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		match, project, unwind, set, unwindAgain, group, sort, limit,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"region": region,
		}).Error("failed to aggregate platforms: ", err)

		return nil, fmt.Errorf("Failed to get top platforms.")
	}

	var streamingPlatforms []responses.StreamingPlatform
	if err := cursor.All(context.TODO(), &streamingPlatforms); err != nil {
		logrus.WithFields(logrus.Fields{
			"region": region,
		}).Error("failed to decode platforms: ", err)

		return nil, fmt.Errorf("Failed to decode top platforms.")
	}

	return streamingPlatforms, nil
}

func (movieModel *MovieModel) GetMoviesByStreamingPlatform(data requests.FilterByStreamingPlatformAndRegion) ([]responses.Movie, p.PaginationData, error) {
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

	match := bson.M{
		"streaming.country_code":             data.Region,
		"streaming.streaming_platforms.name": data.StreamingPlatform,
	}

	var movies []responses.Movie
	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(moviePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&movies).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate movies by streaming platforms: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get movies by streaming platforms.")
	}

	return movies, paginatedData.Pagination, nil
}

func (movieModel *MovieModel) GetPopularProductionCompanies() ([]responses.StreamingPlatform, error) {
	match := bson.M{"$match": bson.M{
		"production_companies.logo": bson.M{
			"$ne": "",
		},
	}}

	project := bson.M{"$project": bson.M{
		"_id":                  1,
		"production_companies": 1,
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$production_companies",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	set := bson.M{"$project": bson.M{
		"logo": "$production_companies.logo",
		"name": "$production_companies.name",
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$name",
		"name": bson.M{
			"$first": "$name",
		},
		"logo": bson.M{
			"$first": "$logo",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"count": -1,
	}}

	limit := bson.M{"$limit": popularPlatformsLimit}

	cursor, err := movieModel.Collection.Aggregate(context.TODO(), bson.A{
		match, project, unwind, set, group, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate production companies: ", err)

		return nil, fmt.Errorf("Failed to get top production companies.")
	}

	var popularProductionCompanies []responses.StreamingPlatform
	if err := cursor.All(context.TODO(), &popularProductionCompanies); err != nil {
		logrus.Error("failed to decode production companies: ", err)

		return nil, fmt.Errorf("Failed to decode top production companies.")
	}

	return popularProductionCompanies, nil
}

func (movieModel *MovieModel) SearchMovieByTitle(data requests.Search) ([]responses.Movie, p.PaginationData, error) {
	search := bson.M{"$search": bson.M{
		"index": "movies_search",
		"text": bson.M{
			"query": data.Search,
			"path": bson.A{
				"title_en",
				"title_original",
			},
			"fuzzy": bson.M{
				"maxEdits": 1,
			},
		},
	}}

	paginatedData, err := p.New(movieModel.Collection).Context(context.TODO()).Limit(movieSearchLimit).
		Page(data.Page).Aggregate(search)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to search movies by title: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to search movies by title.")
	}

	var movies []responses.Movie
	for _, raw := range paginatedData.Data {
		var movie *responses.Movie
		if marshallErr := bson.Unmarshal(raw, &movie); marshallErr == nil {
			movies = append(movies, *movie)
		}
	}

	return movies, paginatedData.Pagination, nil
}
