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
	currentDate := utils.GetCustomDate(0, 0, -1)

	match := bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"status": "In Production"},
					bson.M{
						"streaming.streaming_platforms.name": "Netflix",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Netflix basic with Ads",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Amazon Prime Video",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Disney Plus",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Rakuten Viki",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Apple TV Plus",
					},
					bson.M{
						"streaming.streaming_platforms.name": "HBO Max",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Sky Go",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Paramount Plus",
					},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{
						"first_air_date": bson.M{
							"$gte": currentDate,
						},
					},
					bson.M{
						"seasons.air_date": bson.M{
							"$gte": currentDate,
						},
					},
				},
			},
		},
	}

	// Only fetch required fields for preview
	opts := options.Find().
		SetSort(bson.M{"tmdb_popularity": -1}).
		SetLimit(tvSeriesUpcomingPaginationLimit).
		SetProjection(bson.M{
			"_id":            1,
			"tmdb_id":        1,
			"title_en":       1,
			"title_original": 1,
			"image_url":      1,
		})

	cursor, err := tvModel.Collection.Find(context.TODO(), match, opts)
	if err != nil {
		logrus.Error("failed to find preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to find preview tv series.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewTVSeries, 0, tvSeriesUpcomingPaginationLimit)
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

	// Only project required fields for preview
	project := bson.M{"$project": bson.M{
		"_id":             1,
		"tmdb_id":         1,
		"title_en":        1,
		"title_original":  1,
		"image_url":       1,
		"tmdb_popularity": 1,
	}}

	sort := bson.M{"$sort": bson.M{
		"tmdb_popularity": -1,
	}}

	limit := bson.M{"$limit": tvSeriesPaginationLimit}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		set, project, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate popular preview tv: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview tv series.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewTVSeries, 0, tvSeriesPaginationLimit)
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

	limit := bson.M{"$limit": tvSeriesPaginationLimit}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		addFields, project, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate popular preview tv: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview tv series.")
	}

	// Pre-allocate with known capacity
	results := make([]responses.PreviewTVSeries, 0, tvSeriesPaginationLimit)
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview upcoming: ", err)

		return nil, fmt.Errorf("Failed to decode preview tv series.")
	}

	return results, nil
}

func (tvModel *TVModel) GetTVSeriesFromOpenAI(uid string, tvSeriesIDs []string, limitValue int) ([]responses.AISuggestion, error) {
	match := bson.M{"$match": bson.M{
		"$expr": bson.M{
			"$in": bson.A{bson.M{"$toString": "$_id"}, tvSeriesIDs},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"tmdb_popularity": -1,
	}}

	limit := bson.M{"$limit": limitValue}

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
		"preserveNullAndEmptyArrays": true,
	}}

	lookupNotInterested := bson.M{"$lookup": bson.M{
		"from": "ai-suggestions-not-interested",
		"let": bson.M{
			"uid":   uid,
			"tv_id": "$tv_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": bson.A{"$content_id", "$$tv_id"}},
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
		"content_type":        "tv",
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

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		match, sort, limit, set, lookupWatchLater, unwindWatchLater,
		lookupNotInterested, project,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tvIDs": tvSeriesIDs,
		}).Error("failed to aggregate tv series: ", err)

		return nil, fmt.Errorf("Failed to get tv series from recommendation.")
	}

	var tvList []responses.AISuggestion
	if err := cursor.All(context.TODO(), &tvList); err != nil {
		logrus.WithFields(logrus.Fields{
			"tvIDs": tvSeriesIDs,
		}).Error("failed to decode tv series: ", err)

		return nil, fmt.Errorf("Failed to decode get tv series from recommendation.")
	}

	return tvList, nil
}

func (tvModel *TVModel) GetUpcomingTVSeries(data requests.Pagination) ([]responses.TVSeries, p.PaginationData, error) {
	currentDate := utils.GetCustomDate(0, 0, -1)

	match := bson.M{"$match": bson.M{
		"$and": bson.A{
			bson.M{
				"$or": bson.A{
					bson.M{"status": "In Production"},
					bson.M{
						"streaming.streaming_platforms.name": "Netflix",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Netflix basic with Ads",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Amazon Prime Video",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Disney Plus",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Rakuten Viki",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Apple TV Plus",
					},
					bson.M{
						"streaming.streaming_platforms.name": "HBO Max",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Sky Go",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Paramount Plus",
					},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{
						"first_air_date": bson.M{
							"$gte": currentDate,
						},
					},
					bson.M{
						"seasons.air_date": bson.M{
							"$gte": currentDate,
						},
					},
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
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
		"tmdb_id":         1,
		"image_url":       1,
		"title_en":        1,
		"title_original":  1,
		"description":     1,
		"first_air_date":  1,
		"status":          1,
		"tmdb_popularity": 1,
		"tmdb_vote":       1,
		"tmdb_vote_count": 1,
		"total_episodes":  1,
		"total_seasons":   1,
	}}

	paginatedData, err := p.New(tvModel.Collection).Context(context.TODO()).Limit(tvSeriesUpcomingPaginationLimit).
		Page(data.Page).Sort("has_air_date", -1).Sort("tmdb_popularity", -1).Sort("_id", 1).Aggregate(match, project)
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

func (tvModel *TVModel) GetCurrentlyAiringTVSeriesByDayOfWeek(dayOfWeek int16) ([]responses.PreviewTVSeries, error) {
	currentDate := utils.GetCustomDate(0, -3, 0)

	match := bson.M{"$match": bson.M{
		"$and": bson.A{
			bson.M{"status": "Returning Series"},
			bson.M{"streaming.streaming_platforms": bson.M{
				"$ne": nil,
			}},
			bson.M{
				"$or": bson.A{
					bson.M{
						"streaming.streaming_platforms.name": "Netflix",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Netflix basic with Ads",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Amazon Prime Video",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Disney Plus",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Rakuten Viki",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Apple TV Plus",
					},
					bson.M{
						"streaming.streaming_platforms.name": "HBO Max",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Sky Go",
					},
					bson.M{
						"streaming.streaming_platforms.name": "Paramount Plus",
					},
				},
			},
			bson.M{
				"$or": bson.A{
					bson.M{
						"first_air_date": bson.M{
							"$gte": currentDate,
						},
					},
					bson.M{
						"seasons.air_date": bson.M{
							"$gte": currentDate,
						},
					},
				},
			},
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
		"tmdb_popularity": bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$or": bson.A{
						bson.M{"$in": bson.A{"Reality", "$genres"}},
						bson.M{"$in": bson.A{"Talk", "$genres"}},
					},
				},
				"then": bson.M{
					"$multiply": bson.A{"$tmdb_popularity", 0.5},
				},
				"else": "$tmdb_popularity",
			},
		},
	}}

	matchDayOfWeek := bson.M{"$match": bson.M{
		"dayOfWeek": dayOfWeek,
	}}

	sortByPopularity := bson.M{"$sort": bson.M{
		"tmdb_popularity": -1,
	}}

	limit := bson.M{"$limit": 25}

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
		match, set, airDateNotNull, addFields, matchDayOfWeek, sortByPopularity, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate currently airing tv series: ", err)

		return nil, fmt.Errorf("Failed to get currently airing tv series.")
	}

	var tvSeriesList []responses.PreviewTVSeries
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

	set := bson.M{"$project": bson.M{
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
		"tmdb_id":          1,
		"image_url":        1,
		"has_release_date": 1,
		"title_en":         1,
		"title_original":   1,
		"description":      1,
		"first_air_date":   1,
		"status":           1,
		"tmdb_vote":        1,
		"tmdb_vote_count":  1,
		"total_episodes":   1,
		"total_seasons":    1,
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
		data.FirstAirDateFrom != nil || data.NumSeason != nil ||
		data.ProductionCountry != nil || data.StreamingPlatforms != nil ||
		data.Rating != nil {

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

	// Use FindOne with projection to only fetch required fields
	projection := bson.M{
		"_id":             1,
		"title_en":        1,
		"title_original":  1,
		"description":     1,
		"image_url":       1,
		"status":          1,
		"tmdb_id":         1,
		"tmdb_popularity": 1,
		"tmdb_vote":       1,
		"tmdb_vote_count": 1,
		"total_seasons":   1,
		"total_episodes":  1,
		"first_air_date":  1,
		"backdrop":        1,
		"recommendations": 1,
		"genres":          1,
		"images":          1,
		"videos":          1,
		"streaming":       1,
		"seasons":         1,
		"networks":        1,
		"actors":          1,
	}

	options := options.FindOne().SetProjection(projection)

	result := tvModel.Collection.FindOne(context.TODO(), bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"tmdb_id": data.ID,
			},
		},
	}, options)

	var tvSeries responses.TVSeries
	if err := result.Decode(&tvSeries); err != nil {
		logrus.WithFields(logrus.Fields{
			"tv_id": data.ID,
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

	// Add projection early to reduce data transfer
	project := bson.M{"$project": bson.M{
		"_id":             1,
		"title_en":        1,
		"title_original":  1,
		"description":     1,
		"image_url":       1,
		"status":          1,
		"tmdb_id":         1,
		"tmdb_popularity": 1,
		"tmdb_vote":       1,
		"tmdb_vote_count": 1,
		"total_seasons":   1,
		"total_episodes":  1,
		"first_air_date":  1,
		"backdrop":        1,
		"recommendations": 1,
		"genres":          1,
		"images":          1,
		"videos":          1,
		"streaming":       1,
		"seasons":         1,
		"networks":        1,
		"actors":          1,
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
			// Project only needed fields from TV watch list
			bson.M{
				"$project": bson.M{
					"status":          1,
					"score":           1,
					"times_finished":  1,
					"current_season":  1,
					"current_episode": 1,
					"created_at":      1,
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

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
		match, project, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate tv details: ", err)

		return responses.TVSeriesDetails{}, fmt.Errorf("Failed to aggregate tv details with watch list.")
	}

	// Pre-allocate slice with capacity 1 since we expect only one result
	tvDetails := make([]responses.TVSeriesDetails, 0, 1)
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

	limit := bson.M{"$limit": tvSeriesActorsLimit}

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (tvModel *TVModel) GetPopularProductionCompanies() ([]responses.StreamingPlatform, error) {
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

	cursor, err := tvModel.Collection.Aggregate(context.TODO(), bson.A{
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

func (tvModel *TVModel) SearchTVSeriesByTitle(data requests.Search) ([]responses.TVSeries, p.PaginationData, error) {
	search := bson.M{"$search": bson.M{
		"index": "tv_series_search",
		"compound": bson.M{
			"should": bson.A{
				bson.M{
					"text": bson.M{
						"query": data.Search,
						"path":  "title_en",
						"fuzzy": bson.M{
							"maxEdits": 1,
						},
						"score": bson.M{"boost": bson.M{"value": 5}},
					},
				},
				bson.M{
					"text": bson.M{
						"query": data.Search,
						"path":  "title_original",
						"fuzzy": bson.M{
							"maxEdits": 1,
						},
					},
				},
			},
			"minimumShouldMatch": 1,
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
