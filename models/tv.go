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

type TVModel struct {
	Collection *mongo.Collection
}

func NewTVModel(mongoDB *db.MongoDB) *TVModel {
	return &TVModel{
		Collection: mongoDB.Database.Collection("tv-series"),
	}
}

const (
	tvSeriesUpcomingPaginationLimit = 40
	tvSeriesSearchLimit             = 50
	tvSeriesPaginationLimit         = 40
)

func (tvModel *TVModel) GetTVSeriesFromOpenAI(uid string, tvSeries []string) ([]responses.AISuggestion, error) {
	match := bson.M{"$match": bson.M{
		"title_original": bson.M{
			"$in": tvSeries,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"tmdb_popularity": -1,
	}}

	limit := bson.M{"$limit": 3}

	set := bson.M{"$set": bson.M{
		"tv_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
		"let": bson.M{
			"uid":     uid,
			"tv_id":   "$tv_id",
			"tmdb_id": "$tmdb_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$content_id", "$$tv_id"}},
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
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toString": "$_id",
		},
		"content_external_id": "$tmdb_id",
		"content_type":        "tv",
		"title_en":            1,
		"title_original":      1,
		"description":         1,
		"image_url":           1,
		"score":               "$tmdb_vote",
		"watch_later":         1,
	}}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		match, sort, limit, set, lookupWatchLater, unwindWatchLater, project,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tv": tvSeries,
		}).Error("failed to aggregate tv series: ", err)

		return nil, fmt.Errorf("Failed to get tv series from recommendation.")
	}

	var tvList []responses.AISuggestion
	if err := cursor.All(context.TODO(), &tvList); err != nil {
		logrus.WithFields(logrus.Fields{
			"tv": tvSeries,
		}).Error("failed to decode tv series: ", err)

		return nil, fmt.Errorf("Failed to decode get tv series from recommendation.")
	}

	return tvList, nil
}

func (tvModel *TVModel) GetUpcomingTVSeries(data requests.Pagination) ([]responses.TVSeries, p.PaginationData, error) {
	match := bson.M{"$match": bson.M{
		"status": "In Production",
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_air_date": bson.M{
			"$or": bson.A{
				bson.M{
					"$ne": bson.A{"$first_air_date", ""},
				},
				bson.M{
					"$ne": bson.A{"$first_air_date", nil},
				},
			},
		},
	}}

	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesUpcomingPaginationLimit).
		Page(data.Page).Sort("has_air_date", -1).Sort("tmdb_popularity", -1).Sort("_id", 1).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming tv series: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming tv series.")
	}

	var upcomingTVSeries []responses.TVSeries
	for _, raw := range paginatedData.Data {
		var tvSeries *responses.TVSeries
		if marshalErr := bson.Unmarshal(raw, &tvSeries); marshalErr == nil {
			upcomingTVSeries = append(upcomingTVSeries, *tvSeries)
		}
	}

	return upcomingTVSeries, paginatedData.Pagination, nil
}

func (tvModel *TVModel) GetCurrentlyAiringTVSeriesByDayOfWeek() ([]responses.DayOfWeekTVSeries, error) {
	match := bson.M{"$match": bson.M{
		"streaming.country_code": "US",
		"status":                 "Returning Series",
		"streaming.streaming_platforms": bson.M{
			"$ne": nil,
		},
	}}

	set := bson.M{"$set": bson.M{
		"latest_season": bson.M{
			"$arrayElemAt": bson.A{"$seasons", -1},
		},
	}}

	airDateNotNull := bson.M{"$match": bson.M{
		"latest_season.air_date": bson.M{
			"$ne": nil,
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"dayOfWeek": bson.M{
			"$dayOfWeek": bson.M{
				"$dateFromString": bson.M{
					"dateString": "$latest_season.air_date",
				},
			},
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$dayOfWeek",
		"data": bson.M{
			"$push": "$$ROOT",
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"_id": 1,
	}}

	setSlice := bson.M{"$set": bson.M{
		"day_of_week": "$_id",
		"data": bson.M{
			"$slice": bson.A{"$data", 25},
		},
	}}

	sortArray := bson.M{"$set": bson.M{
		"data": bson.M{
			"$sortArray": bson.M{
				"input": "$data",
				"sortBy": bson.M{
					"tmdb_popularity": -1,
				},
			},
		},
	}}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, airDateNotNull, addFields, group, sort, setSlice, sortArray,
	})
	if err != nil {
		logrus.Error("failed to aggregate currently airing tv series: ", err)

		return nil, fmt.Errorf("Failed to get currently airing tv series.")
	}

	var tvSeriesList []responses.DayOfWeekTVSeries
	if err := cursor.All(context.TODO(), &tvSeriesList); err != nil {
		logrus.Error("failed to decode tv series by user id: ", err)

		return nil, fmt.Errorf("Failed to decode tv series by user id.")
	}

	return tvSeriesList, nil
}

func (tvModel *TVModel) GetTVSeriesBySortAndFilter(data requests.SortFilterTVSeries) ([]responses.TVSeries, p.PaginationData, error) {
	addFields := bson.M{"$addFields": bson.M{
		"top_rated": bson.M{
			"$multiply": bson.A{
				"$tmdb_vote", "$tmdb_vote_count",
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"tmdb_popularity": bson.M{
			"$multiply": bson.A{
				bson.M{
					"$sqrt": bson.M{
						"$multiply": bson.A{
							"$tmdb_vote", "$tmdb_vote_count",
						},
					},
				},
				"$tmdb_popularity",
			},
		},
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
		sortType = "first_air_date"
		sortOrder = -1
	case "old":
		sortType = "first_air_date"
		sortOrder = 1
	}

	matchFields := bson.M{}
	if data.Status != nil || data.Genres != nil || data.ProductionCompanies != nil ||
		data.FirstAirDateFrom != nil || data.NumSeason != nil {

		if data.Status != nil {
			var status string
			switch *data.Status {
			case "production":
				status = "In Production"
			case "airing":
				status = "Returning Series"
			case "ended":
				status = "Ended"
			}

			matchFields["status"] = status
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

		if data.FirstAirDateFrom != nil {
			if data.FirstAirDateTo != nil {
				matchFields["first_air_date"] = bson.M{
					"$gte": strconv.Itoa(*data.FirstAirDateFrom),
					"$lt":  strconv.Itoa(*data.FirstAirDateTo),
				}
			} else {
				matchFields["first_air_date"] = bson.M{
					"$gte": strconv.Itoa(*data.FirstAirDateFrom),
				}
			}
		}

		if data.NumSeason != nil {
			matchFields["total_seasons"] = bson.M{
				"$gte": *data.NumSeason,
			}
		}
	}

	match := bson.M{"$match": matchFields}

	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesPaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(match, addFields, set)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
			"match":   match,
		}).Error("failed to aggregate tv series by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get tv series by selected filters.")
	}

	var tvSeriesList []responses.TVSeries
	for _, raw := range paginatedData.Data {
		var tvSeries *responses.TVSeries
		if marshalErr := bson.Unmarshal(raw, &tvSeries); marshalErr == nil {
			tvSeriesList = append(tvSeriesList, *tvSeries)
		}
	}

	return tvSeriesList, paginatedData.Pagination, nil
}

func (tvModel *TVModel) GetTVSeriesDetails(data requests.ID) (responses.TVSeries, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	result := tvModel.Collection.FindOne(context.TODO(), bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"tmdb_id": data.ID,
			},
		},
	})

	var tvSeries responses.TVSeries
	if err := result.Decode(&tvSeries); err != nil {
		logrus.WithFields(logrus.Fields{
			"game_id": data.ID,
		}).Error("failed to find tv series details by id: ", err)

		return responses.TVSeries{}, fmt.Errorf("Failed to find tv series by id.")
	}

	return tvSeries, nil
}

func (tvModel *TVModel) GetTVSeriesDetailsWithWatchListAndWatchLater(data requests.ID, uuid string) (responses.TVSeriesDetails, error) {
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

	set := bson.M{"$set": bson.M{
		"tv_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from": "tvseries-watch-lists",
		"let": bson.M{
			"uuid":    uuid,
			"tv_id":   "$tv_id",
			"tmdb_id": "$tmdb_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$tv_id", "$$tv_id"}},
									bson.M{"$eq": bson.A{"$tv_tmdb_id", "$$tmdb_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
		},
		"as": "tv_list",
	}}

	unwindWatchList := bson.M{"$unwind": bson.M{
		"path":                       "$tv_list",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
		"let": bson.M{
			"uuid":    uuid,
			"tv_id":   "$tv_id",
			"tmdb_id": "$tmdb_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$content_id", "$$tv_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$tmdb_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
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

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate tv details: ", err)

		return responses.TVSeriesDetails{}, fmt.Errorf("Failed to aggregate tv details with watch list.")
	}

	var tvDetails []responses.TVSeriesDetails
	if err = cursor.All(context.TODO(), &tvDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to decode tv details: ", err)

		return responses.TVSeriesDetails{}, fmt.Errorf("Failed to decode tv details.")
	}

	if len(tvDetails) > 0 {
		return tvDetails[0], nil
	}

	return responses.TVSeriesDetails{}, nil
}

func (tvModel *TVModel) SearchTVSeriesByTitle(data requests.Search) ([]responses.TVSeries, p.PaginationData, error) {
	search := bson.M{"$search": bson.M{
		"index": "tv_series_search",
		"text": bson.M{
			"query": data.Search,
			"path":  bson.A{"title_en", "title_original", "translations.title"},
			"fuzzy": bson.M{
				"maxEdits": 1,
			},
		},
	}}

	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesSearchLimit).
		Page(data.Page).Aggregate(search)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to search tv series by title: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to search tv series by title.")
	}

	var tvSeries []responses.TVSeries
	for _, raw := range paginatedData.Data {
		var tv *responses.TVSeries
		if marshallErr := bson.Unmarshal(raw, &tv); marshallErr == nil {
			tvSeries = append(tvSeries, *tv)
		}
	}

	return tvSeries, paginatedData.Pagination, nil
}
