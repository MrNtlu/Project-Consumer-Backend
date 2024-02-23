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
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

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
	tvSeriesActorsLimit             = 50
)

func (tvModel *TVModel) GetUpcomingPreviewTVSeries() ([]responses.PreviewTVSeries, error) {
	match := bson.M{
		"status": "In Production",
	}

	opts := options.Find().SetSort(bson.M{"tmdb_popularity": -1}).SetLimit(tvSeriesUpcomingPaginationLimit)

	cursor, err := tvModel.Collection.Find(context.TODO(), match, opts)
	if err != nil {
		logrus.Error("failed to find preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to find preview tv series.")
	}

	var results []responses.PreviewTVSeries
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to decode preview tv series.")
	}

	return results, nil
}

func (tvModel *TVModel) GetPopularPreviewTVSeries() ([]responses.PreviewTVSeries, error) {
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

	sort := bson.M{"$sort": bson.M{
		"tmdb_popularity": -1,
	}}

	limit := bson.D{{"$limit", tvSeriesPaginationLimit}}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		set, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate popular preview tv: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview tv series.")
	}

	var results []responses.PreviewTVSeries
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to decode preview tv series.")
	}

	return results, nil
}

func (tvModel *TVModel) GetTopPreviewTVSeries() ([]responses.PreviewTVSeries, error) {
	addFields := bson.M{"$addFields": bson.M{
		"top_rated": bson.M{
			"$multiply": bson.A{
				"$tmdb_vote", "$tmdb_vote_count",
			},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"top_rated": -1,
	}}

	limit := bson.D{{"$limit", tvSeriesPaginationLimit}}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		addFields, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate popular preview tv: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview tv series.")
	}

	var results []responses.PreviewTVSeries
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to decode preview tv series.")
	}

	return results, nil
}

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

	sortByPopularity := bson.M{"$sort": bson.M{
		"tmdb_popularity": 1,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$dayOfWeek",
		"data": bson.M{
			"$push": "$$ROOT",
		},
	}}

	setDataSlice := bson.M{"$set": bson.M{
		"data": bson.M{
			"$slice": bson.A{"$data", 25},
		},
		"day_of_week": "$_id",
	}}

	sort := bson.M{"$sort": bson.M{
		"_id": 1,
	}}

	// sortArray := bson.M{"$set": bson.M{
	// 	"data": bson.M{
	// 		"$sortArray": bson.M{
	// 			"input": "$data",
	// 			"sortBy": bson.M{
	// 				"tmdb_popularity": -1,
	// 			},
	// 		},
	// 	},
	// }}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, airDateNotNull, addFields, sortByPopularity, group, setDataSlice, sort,
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

func (tvModel *TVModel) GetPopularActors(data requests.Pagination) ([]responses.ActorDetails, error) {
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

	limit := bson.M{"$limit": tvSeriesActorsLimit}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		unwind, group, set, sort, limit,
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

func (tvModel *TVModel) GetTVSeriesByActor(data requests.IDPagination) ([]responses.TVSeries, p.PaginationData, error) {
	match := bson.M{
		"actors.tmdb_id": bson.M{
			"$in": bson.A{data.ID},
		},
	}

	var tvList []responses.TVSeries
	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesPaginationLimit).
		Page(data.Page).Sort("tmdb_popularity", -1).Filter(match).Decode(&tvList).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate tv series by actor: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get tv series by actor.")
	}

	return tvList, paginatedData.Pagination, nil
}

func (tvModel *TVModel) GetPopularStreamingPlatforms(region string) ([]responses.StreamingPlatform, error) {
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

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (tvModel *TVModel) GetTVSeriesByStreamingPlatform(data requests.FilterByStreamingPlatformAndRegion) ([]responses.TVSeries, p.PaginationData, error) {
	match := bson.M{
		"streaming.country_code":             data.Region,
		"streaming.streaming_platforms.name": data.StreamingPlatform,
	}

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

	var tvList []responses.TVSeries
	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesPaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&tvList).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate tv series by streaming platform: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get tv series by streaming platform.")
	}

	return tvList, paginatedData.Pagination, nil
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
