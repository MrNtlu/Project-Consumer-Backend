package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"app/utils"
	"context"
	"fmt"
	"strconv"
	"time"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type AnimeModel struct {
	Collection *mongo.Collection
}

func NewAnimeModel(mongoDB *db.MongoDB) *AnimeModel {
	return &AnimeModel{
		Collection: mongoDB.Database.Collection("animes"),
	}
}

const (
	animeUpcomingPaginationLimit = 40
	animeSearchLimit             = 50
	animePaginationLimit         = 40
)

func (animeModel *AnimeModel) GetPreviewUpcomingAnimes() ([]responses.PreviewAnime, error) {
	match := bson.M{"$match": bson.M{
		"is_airing": false,
		"$or": bson.A{
			bson.M{"status": "Not yet aired"},
			bson.M{"aired.from": bson.M{"$gte": utils.GetCurrentDate()}},
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_year": bson.M{
			"$ne": bson.A{"$year", nil},
		},
	}}

	addPopularityFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"has_year":   -1,
		"popularity": -1,
	}}

	limit := bson.M{"$limit": animeUpcomingPaginationLimit}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, addPopularityFields, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate preview upcoming anime: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview upcoming animes.")
	}

	var results []responses.PreviewAnime
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview upcoming animes: ", err)

		return nil, fmt.Errorf("Failed to decode preview upcoming animes.")
	}

	return results, nil
}

func (animeModel *AnimeModel) GetPreviewPopularAnimes() ([]responses.PreviewAnime, error) {
	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{"status": "Currently Airing"},
			bson.M{"is_airing": true},
		},
		"aired.from": bson.M{
			"$gte": utils.GetCustomDate(0, -4, 0),
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"popularity": -1,
	}}

	limit := bson.M{"$limit": animePaginationLimit}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate popular preview anime: ", err)

		return nil, fmt.Errorf("Failed to aggregate preview popular animes.")
	}

	var results []responses.PreviewAnime
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview popular animes: ", err)

		return nil, fmt.Errorf("Failed to decode preview popular animes.")
	}

	return results, nil
}

func (animeModel *AnimeModel) GetPreviewTopAnimes() ([]responses.PreviewAnime, error) {
	filter := bson.D{}
	opts := options.Find().SetSort(bson.M{"mal_score": -1}).SetLimit(animePaginationLimit)

	cursor, err := animeModel.Collection.Find(context.TODO(), filter, opts)
	if err != nil {
		logrus.Error("failed to find preview top animes: ", err)

		return nil, fmt.Errorf("Failed to find preview top animes.")
	}

	var results []responses.PreviewAnime
	if err = cursor.All(context.TODO(), &results); err != nil {
		logrus.Error("failed to decode preview top animes: ", err)

		return nil, fmt.Errorf("Failed to decode preview top animes.")
	}

	return results, nil
}

func (animeModel *AnimeModel) GetAnimeFromOpenAI(uid string, anime []string) ([]responses.AISuggestion, error) {
	match := bson.M{"$match": bson.M{
		"title_original": bson.M{
			"$in": anime,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"mal_score": -1,
	}}

	limit := bson.M{"$limit": 3}

	set := bson.M{"$set": bson.M{
		"anime_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
		"let": bson.M{
			"uid":      uid,
			"anime_id": "$anime_id",
			"mal_id":   "$mal_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$content_id", "$$anime_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$mal_id"}},
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
		"content_external_int_id": "$mal_id",
		"content_type":            "anime",
		"title_en":                1,
		"title_original":          1,
		"description":             1,
		"image_url":               1,
		"score":                   "$mal_score",
		"watch_later":             1,
	}}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, sort, limit, set, lookupWatchLater, unwindWatchLater, project,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"anime": anime,
		}).Error("failed to aggregate anime: ", err)

		return nil, fmt.Errorf("Failed to get anime from recommendation.")
	}

	var animeList []responses.AISuggestion
	if err := cursor.All(context.TODO(), &animeList); err != nil {
		logrus.WithFields(logrus.Fields{
			"anime": anime,
		}).Error("failed to decode animes: ", err)

		return nil, fmt.Errorf("Failed to decode get anime from recommendation.")
	}

	return animeList, nil
}

func (animeModel *AnimeModel) GetUpcomingAnimesBySort(data requests.Pagination) ([]responses.Anime, p.PaginationData, error) {
	currentSeason := getSeasonFromMonth()

	match := bson.M{"$match": bson.M{
		"is_airing": false,
		"$or": bson.A{
			bson.M{"status": "Not yet aired"},
			bson.M{"aired.from": bson.M{"$gte": utils.GetCurrentDate()}},
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_year": bson.M{
			"$ne": bson.A{"$year", nil},
		},
		"season_priority": bson.M{
			"$switch": bson.M{
				"branches": bson.A{
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$season", "winter"},
						},
						"then": getSeasonPriority(currentSeason, "winter"),
					},
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$season", "spring"},
						},
						"then": getSeasonPriority(currentSeason, "spring"),
					},
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$season", "summer"},
						},
						"then": getSeasonPriority(currentSeason, "summer"),
					},
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$season", "fall"},
						},
						"then": getSeasonPriority(currentSeason, "fall"),
					},
				},
				"default": 4,
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
		"has_year":        1,
		"season_priority": 1,
		"mal_id":          1,
		"age_rating":      1,
		"aired":           1,
		"description":     1,
		"episodes":        1,
		"image_url":       1,
		"is_airing":       1,
		"mal_score":       1,
		"mal_scored_by":   1,
		"season":          1,
		"source":          1,
		"status":          1,
		"title_en":        1,
		"title_jp":        1,
		"title_original":  1,
		"type":            1,
		"year":            1,
		"mal_favorites":   1,
		"mal_members":     1,
	}}

	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animeUpcomingPaginationLimit).
		Page(data.Page).Sort("has_year", -1).Sort("popularity", -1).Sort("_id", 1).Aggregate(match, addFields, project)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming animes: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming animes.")
	}

	var upcomingAnimes []responses.Anime
	for _, raw := range paginatedData.Data {
		var anime *responses.Anime
		if marshallErr := bson.Unmarshal(raw, &anime); marshallErr == nil {
			upcomingAnimes = append(upcomingAnimes, *anime)
		}
	}

	return upcomingAnimes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) GetCurrentlyAiringAnimesByDayOfWeek(dayOfWeek int16) ([]responses.PreviewAnime, error) {
	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{"status": "Currently Airing"},
			bson.M{"is_airing": true},
		},
		"aired.from": bson.M{
			"$ne": nil,
		},
		"type": bson.M{
			"$ne": "Movie",
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"dayOfWeek": bson.M{
			"$dayOfWeek": bson.M{
				"$dateFromString": bson.M{
					"dateString": "$aired.from",
				},
			},
		},
	}}

	matchDayOfWeek := bson.M{"$match": bson.M{
		"dayOfWeek": dayOfWeek,
	}}

	sortByScore := bson.M{"$sort": bson.M{
		"mal_score": -1,
	}}

	limit := bson.M{"$limit": 25}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, matchDayOfWeek, sortByScore, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate currently airing animes: ", err)

		return nil, err
	}

	var animeList []responses.PreviewAnime
	if err = cursor.All(context.TODO(), &animeList); err != nil {
		logrus.Error("failed to decode currently airing animes: ", err)

		return nil, err
	}

	return animeList, nil
}

func (animeModel *AnimeModel) GetPopularAnimes(data requests.Pagination) ([]responses.Anime, p.PaginationData, error) {
	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{"status": "Currently Airing"},
			bson.M{"is_airing": true},
		},
		"aired.from": bson.M{
			"$gte": utils.GetCustomDate(0, -4, 0),
		},
	}}

	project := bson.M{"$project": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
		"mal_id":         1,
		"age_rating":     1,
		"aired":          1,
		"description":    1,
		"episodes":       1,
		"image_url":      1,
		"is_airing":      1,
		"mal_score":      1,
		"mal_scored_by":  1,
		"season":         1,
		"source":         1,
		"status":         1,
		"title_en":       1,
		"title_jp":       1,
		"title_original": 1,
		"type":           1,
		"year":           1,
		"mal_favorites":  1,
		"mal_members":    1,
	}}

	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animePaginationLimit).
		Page(data.Page).Sort("popularity", -1).Aggregate(match, project)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
			"match":   match,
		}).Error("failed to aggregate popular animes: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get popular anime.")
	}

	var animes []responses.Anime
	for _, raw := range paginatedData.Data {
		var anime *responses.Anime
		if marshalErr := bson.Unmarshal(raw, &anime); marshalErr == nil {
			animes = append(animes, *anime)
		}
	}

	return animes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) GetAnimesBySortAndFilter(data requests.SortFilterAnime) ([]responses.Anime, p.PaginationData, error) {
	project := bson.M{"$project": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
		"mal_id":         1,
		"age_rating":     1,
		"aired":          1,
		"description":    1,
		"episodes":       1,
		"image_url":      1,
		"is_airing":      1,
		"mal_score":      1,
		"mal_scored_by":  1,
		"season":         1,
		"source":         1,
		"status":         1,
		"title_en":       1,
		"title_jp":       1,
		"title_original": 1,
		"type":           1,
		"year":           1,
		"mal_favorites":  1,
		"mal_members":    1,
	}}

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "top":
		sortType = "mal_score"
		sortOrder = -1
	case "popularity":
		if data.Year != nil && data.Season != nil {
			sortType = "mal_score"
		} else {
			sortType = "popularity"
		}
		sortOrder = -1
	case "new":
		sortType = "aired.from"
		sortOrder = -1
	case "old":
		sortType = "aired.from"
		sortOrder = 1
	}

	matchFields := bson.M{}
	if data.Status != nil || data.Genres != nil || data.Demographics != nil ||
		data.Studios != nil || data.Themes != nil || data.Year != nil ||
		data.Season != nil || data.StreamingPlatforms != nil {

		if data.Status != nil {

			var status string
			switch *data.Status {
			case "airing":
				status = "Currently Airing"
			case "upcoming":
				status = "Not yet aired"
			case "finished":
				status = "Finished Airing"
			}

			matchFields["status"] = status
		}

		if data.Genres != nil {
			matchFields["genres.name"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.Demographics != nil {
			matchFields["demographics.name"] = bson.M{
				"$in": bson.A{data.Demographics},
			}
		}

		if data.Studios != nil {
			matchFields["studios.name"] = bson.M{
				"$in": bson.A{data.Studios},
			}
		}

		if data.Themes != nil {
			matchFields["themes.name"] = bson.M{
				"$in": bson.A{data.Themes},
			}
		}

		if data.StreamingPlatforms != nil {
			matchFields["streaming.name"] = bson.M{
				"$in": bson.A{data.StreamingPlatforms},
			}
		}

		if data.Year != nil {
			matchFields["year"] = data.Year
		}

		if data.Season != nil {
			matchFields["season"] = data.Season
		}
	}

	match := bson.M{"$match": matchFields}

	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(match, project)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
			"match":   match,
		}).Error("failed to aggregate animes by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get animes by selected filters.")
	}

	var animes []responses.Anime
	for _, raw := range paginatedData.Data {
		var anime *responses.Anime
		if marshalErr := bson.Unmarshal(raw, &anime); marshalErr == nil {
			animes = append(animes, *anime)
		}
	}

	return animes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) SearchAnimeByTitle(data requests.Search) ([]responses.Anime, p.PaginationData, error) {
	search := bson.M{"$search": bson.M{
		"index": "anime_search",
		"text": bson.M{
			"query": data.Search,
			"path":  bson.A{"title_en", "title_jp", "title_original"},
			"fuzzy": bson.M{
				"maxEdits": 1,
			},
		},
	}}

	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animeSearchLimit).
		Page(data.Page).Aggregate(search)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
		}).Error("failed to search anime by title: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to search anime by title.")
	}

	var animes []responses.Anime
	for _, raw := range paginatedData.Data {
		var anime *responses.Anime
		if marshallErr := bson.Unmarshal(raw, &anime); marshallErr == nil {
			animes = append(animes, *anime)
		}
	}

	return animes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) GetAnimeDetails(data requests.ID) (responses.Anime, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)
	malID, _ := strconv.Atoi(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"mal_id": malID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"anime_id": bson.M{
			"$toString": "$_id",
		},
	}}

	unwindRelations := bson.M{"$unwind": bson.M{
		"path":                       "$relations",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindSource := bson.M{"$unwind": bson.M{
		"path":                       "$relations.source",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	setRelation := bson.M{"$set": bson.M{
		"relations.mal_id": "$relations.source.mal_id",
		"relations.type":   "$relations.source.type",
	}}

	relationLookup := bson.M{"$lookup": bson.M{
		"from": "animes",
		"let": bson.M{
			"mal_id":   "$relations.mal_id",
			"relation": "$relations.relation",
			"type":     "$relations.type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$mal_id", "$$mal_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 1,
					"anime_id": bson.M{
						"$toString": "$_id",
					},
					"mal_id":         1,
					"title_en":       1,
					"title_original": 1,
					"image_url":      1,
					"relation":       "$$relation",
					"type":           "$$type",
				},
			},
		},
		"as": "relation",
	}}

	unwindRelation := bson.M{"$unwind": bson.M{
		"path":                       "$relation",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"title_original": bson.M{
			"$first": "$title_original",
		},
		"title_en": bson.M{
			"$first": "$title_en",
		},
		"title_jp": bson.M{
			"$first": "$title_jp",
		},
		"description": bson.M{
			"$first": "$description",
		},
		"image_url": bson.M{
			"$first": "$image_url",
		},
		"mal_id": bson.M{
			"$first": "$mal_id",
		},
		"mal_score": bson.M{
			"$first": "$mal_score",
		},
		"mal_scored_by": bson.M{
			"$first": "$mal_scored_by",
		},
		"trailer": bson.M{
			"$first": "$trailer",
		},
		"type": bson.M{
			"$first": "$type",
		},
		"source": bson.M{
			"$first": "$source",
		},
		"episodes": bson.M{
			"$first": "$episodes",
		},
		"season": bson.M{
			"$first": "$season",
		},
		"year": bson.M{
			"$first": "$year",
		},
		"status": bson.M{
			"$first": "$status",
		},
		"is_airing": bson.M{
			"$first": "$is_airing",
		},
		"age_rating": bson.M{
			"$first": "$age_rating",
		},
		"aired": bson.M{
			"$first": "$aired",
		},
		"recommendations": bson.M{
			"$first": "$recommendations",
		},
		"streaming": bson.M{
			"$first": "$streaming",
		},
		"producers": bson.M{
			"$first": "$producers",
		},
		"studios": bson.M{
			"$first": "$studios",
		},
		"genres": bson.M{
			"$first": "$genres",
		},
		"themes": bson.M{
			"$first": "$themes",
		},
		"demographics": bson.M{
			"$first": "$demographics",
		},
		"relations": bson.M{
			"$addToSet": "$relation",
		},
		"characters": bson.M{
			"$first": "$characters",
		},
		"anime_list": bson.M{
			"$first": "$anime_list",
		},
		"watch_later": bson.M{
			"$first": "$watch_later",
		},
	}}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, unwindRelations, unwindSource,
		setRelation, relationLookup, unwindRelation, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to aggregate anime details: ", err)

		return responses.Anime{}, fmt.Errorf("Failed to aggregate anime details.")
	}

	var animeDetails []responses.Anime
	if err = cursor.All(context.TODO(), &animeDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to decode anime details: ", err)

		return responses.Anime{}, fmt.Errorf("Failed to decode anime details.")
	}

	if len(animeDetails) > 0 {
		return animeDetails[0], nil
	}

	return responses.Anime{}, nil
}

func (animeModel *AnimeModel) GetAnimeDetailsWithWatchList(data requests.ID, uuid string) (responses.AnimeDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)
	malID, _ := strconv.Atoi(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"mal_id": malID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"anime_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from": "anime-lists",
		"let": bson.M{
			"uuid":     uuid,
			"anime_id": "$anime_id",
			"mal_id":   "$mal_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$anime_id", "$$anime_id"}},
									bson.M{"$eq": bson.A{"$anime_mal_id", "$$anime_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
		},
		"as": "anime_list",
	}}

	unwindWatchList := bson.M{"$unwind": bson.M{
		"path":                       "$anime_list",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
		"let": bson.M{
			"uuid":     uuid,
			"anime_id": "$anime_id",
			"mal_id":   "$mal_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$content_id", "$$anime_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$mal_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$content_type", "anime"}},
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

	unwindRelations := bson.M{"$unwind": bson.M{
		"path":                       "$relations",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	unwindSource := bson.M{"$unwind": bson.M{
		"path":                       "$relations.source",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	setRelation := bson.M{"$set": bson.M{
		"relations.mal_id": "$relations.source.mal_id",
		"relations.type":   "$relations.source.type",
	}}

	relationLookup := bson.M{"$lookup": bson.M{
		"from": "animes",
		"let": bson.M{
			"mal_id":   "$relations.mal_id",
			"relation": "$relations.relation",
			"type":     "$relations.type",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$mal_id", "$$mal_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 1,
					"anime_id": bson.M{
						"$toString": "$_id",
					},
					"mal_id":         1,
					"title_en":       1,
					"title_original": 1,
					"image_url":      1,
					"relation":       "$$relation",
					"type":           "$$type",
				},
			},
		},
		"as": "relation",
	}}

	unwindRelation := bson.M{"$unwind": bson.M{
		"path":                       "$relation",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"title_original": bson.M{
			"$first": "$title_original",
		},
		"title_en": bson.M{
			"$first": "$title_en",
		},
		"title_jp": bson.M{
			"$first": "$title_jp",
		},
		"description": bson.M{
			"$first": "$description",
		},
		"image_url": bson.M{
			"$first": "$image_url",
		},
		"mal_id": bson.M{
			"$first": "$mal_id",
		},
		"mal_score": bson.M{
			"$first": "$mal_score",
		},
		"mal_scored_by": bson.M{
			"$first": "$mal_scored_by",
		},
		"trailer": bson.M{
			"$first": "$trailer",
		},
		"type": bson.M{
			"$first": "$type",
		},
		"source": bson.M{
			"$first": "$source",
		},
		"episodes": bson.M{
			"$first": "$episodes",
		},
		"season": bson.M{
			"$first": "$season",
		},
		"year": bson.M{
			"$first": "$year",
		},
		"status": bson.M{
			"$first": "$status",
		},
		"is_airing": bson.M{
			"$first": "$is_airing",
		},
		"age_rating": bson.M{
			"$first": "$age_rating",
		},
		"aired": bson.M{
			"$first": "$aired",
		},
		"recommendations": bson.M{
			"$first": "$recommendations",
		},
		"streaming": bson.M{
			"$first": "$streaming",
		},
		"producers": bson.M{
			"$first": "$producers",
		},
		"studios": bson.M{
			"$first": "$studios",
		},
		"genres": bson.M{
			"$first": "$genres",
		},
		"themes": bson.M{
			"$first": "$themes",
		},
		"demographics": bson.M{
			"$first": "$demographics",
		},
		"relations": bson.M{
			"$addToSet": "$relation",
		},
		"characters": bson.M{
			"$first": "$characters",
		},
		"anime_list": bson.M{
			"$first": "$anime_list",
		},
		"watch_later": bson.M{
			"$first": "$watch_later",
		},
	}}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
		unwindRelations, unwindSource, setRelation, relationLookup, unwindRelation, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate anime details: ", err)

		return responses.AnimeDetails{}, fmt.Errorf("Failed to aggregate anime details with watch list.")
	}

	var animeDetails []responses.AnimeDetails
	if err = cursor.All(context.TODO(), &animeDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to decode anime details: ", err)

		return responses.AnimeDetails{}, fmt.Errorf("Failed to decode anime details.")
	}

	if len(animeDetails) > 0 {
		return animeDetails[0], nil
	}

	return responses.AnimeDetails{}, nil
}

func (animeModel *AnimeModel) GetPopularStreamingPlatforms() ([]responses.AnimeNameURL, error) {
	match := bson.M{"$match": bson.M{
		"streaming": bson.M{
			"$not": bson.M{
				"$size": 0,
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$streaming",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$streaming.name",
		"name": bson.M{
			"$first": "$streaming.name",
		},
		"url": bson.M{
			"$first": "$streaming.url",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"count": -1,
	}}

	limit := bson.M{"$limit": popularPlatformsLimit}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, unwind, group, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate platforms: ", err)

		return nil, fmt.Errorf("Failed to get popular platforms.")
	}

	var streamingPlatforms []responses.AnimeNameURL
	if err := cursor.All(context.TODO(), &streamingPlatforms); err != nil {
		logrus.Error("failed to decode platforms: ", err)

		return nil, fmt.Errorf("Failed to decode popular platforms.")
	}

	return streamingPlatforms, nil
}

func (animeModel *AnimeModel) GetAnimesByStreamingPlatform(data requests.FilterByStreamingPlatform) ([]responses.Anime, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "mal_score"
		sortOrder = -1
	case "new":
		sortType = "aired.from"
		sortOrder = -1
	case "old":
		sortType = "aired.from"
		sortOrder = 1
	}

	match := bson.M{
		"streaming.name": data.StreamingPlatform,
	}

	var animes []responses.Anime
	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&animes).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate animes by streaming platforms: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get animes by streaming platforms.")
	}

	return animes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) GetPopularStudios() ([]responses.AnimeNameURL, error) {
	match := bson.M{"$match": bson.M{
		"studios": bson.M{
			"$not": bson.M{
				"$size": 0,
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$studios",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$studios.name",
		"name": bson.M{
			"$first": "$studios.name",
		},
		"url": bson.M{
			"$first": "$studios.url",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"count": -1,
	}}

	limit := bson.M{"$limit": popularPlatformsLimit}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, unwind, group, sort, limit,
	})
	if err != nil {
		logrus.Error("failed to aggregate studios: ", err)

		return nil, fmt.Errorf("Failed to get popular studios.")
	}

	var studios []responses.AnimeNameURL
	if err := cursor.All(context.TODO(), &studios); err != nil {
		logrus.Error("failed to decode studios: ", err)

		return nil, fmt.Errorf("Failed to decode popular studios.")
	}

	return studios, nil
}

func (animeModel *AnimeModel) GetAnimesByStudios(data requests.FilterByStudio) ([]responses.Anime, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "mal_score"
		sortOrder = -1
	case "new":
		sortType = "aired.from"
		sortOrder = -1
	case "old":
		sortType = "aired.from"
		sortOrder = 1
	}

	match := bson.M{
		"studios.name": data.Studio,
	}

	var animes []responses.Anime
	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&animes).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate animes by streaming platforms: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get animes by streaming platforms.")
	}

	return animes, paginatedData.Pagination, nil
}

func getSeasonFromMonth() string {
	switch int(time.Now().Month()) {
	case 12, 1, 2:
		return "winter"
	case 3, 4, 5:
		return "spring"
	case 6, 7, 8:
		return "summer"
	default:
		return "fall"
	}
}

func getSeasonPriority(currentSeason, season string) int {
	if currentSeason == season {
		return 0
	}

	currentSeasonIndex := getSeasonIndex(currentSeason)
	seasonIndex := getSeasonIndex(season)

	if currentSeasonIndex > seasonIndex {
		return 4 - (currentSeasonIndex + seasonIndex)
	} else {
		return seasonIndex - currentSeasonIndex
	}
}

func getSeasonIndex(season string) int {
	seasons := [4]string{"winter", "spring", "summer", "fall"}

	for index, s := range seasons {
		if s == season {
			return index
		}
	}
	return -1
}
