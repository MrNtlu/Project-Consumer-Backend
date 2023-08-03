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
)

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

/* TODO Endpoints
* [x] Get upcoming by popularity etc.
* [x] Get by season/year
* [x] Get currently airing animes by day
* [x] Get anime by popularity, genre etc.
* [x] Get top airing
* [x] Get top upcoming
* [x] Get anime details
 */

func (animeModel *AnimeModel) GetPopularAnimesBySort(data requests.Pagination) ([]responses.Anime, p.PaginationData, error) {
	addFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$mal_score", "$mal_scored_by",
			},
		},
	}}

	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animePaginationLimit).
		Page(data.Page).Sort("popularity", -1).Aggregate(addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate popular animes: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get popular animes.")
	}

	var popularAnimes []responses.Anime
	for _, raw := range paginatedData.Data {
		var anime *responses.Anime
		if marshalErr := bson.Unmarshal(raw, &anime); marshalErr == nil {
			popularAnimes = append(popularAnimes, *anime)
		}
	}

	return popularAnimes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) GetUpcomingAnimesBySort(data requests.SortUpcoming) ([]responses.Anime, p.PaginationData, error) {
	currentSeason := getSeasonFromMonth()

	var (
		sortType            string
		sortOrder           int8
		hasReleaseDateOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "popularity"
		sortOrder = -1
		hasReleaseDateOrder = -1
	case "soon":
		sortType = "aired.from"
		sortOrder = 1
		hasReleaseDateOrder = -1
	case "later":
		sortType = "aired.from"
		sortOrder = -1
		hasReleaseDateOrder = 1
	}

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

	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animeUpcomingPaginationLimit).
		Page(data.Page).Sort("has_year", hasReleaseDateOrder).Sort(sortType, sortOrder).Sort("_id", 1).Aggregate(match, addFields)
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

func (animeModel *AnimeModel) GetAnimesByYearAndSeason(data requests.SortByYearSeasonAnime) ([]responses.Anime, p.PaginationData, error) {
	year := time.Now().Year()

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		if year == int(data.Year) && getSeasonIndex(getSeasonFromMonth()) < getSeasonIndex(data.Season) {
			sortType = "popularity"
		} else {
			sortType = "mal_score"
		}
		sortOrder = -1
	case "new":
		sortType = "aired.from"
		sortOrder = -1
	case "old":
		sortType = "aired.from"
		sortOrder = 1
	}

	match := bson.M{
		"year":   data.Year,
		"season": data.Season,
	}

	var animes []responses.Anime
	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animeUpcomingPaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&animes).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate animes by season and year: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get animes by season and year.")
	}

	return animes, paginatedData.Pagination, nil
}

func (animeModel *AnimeModel) GetCurrentlyAiringAnimesByDayOfWeek() ([]responses.CurrentlyAiringAnimeResponse, error) {
	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{"status": "Currently Airing"},
			bson.M{"is_airing": true},
		},
		"aired.from": bson.M{
			"$ne": nil,
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

	sortByScore := bson.M{"$sort": bson.M{
		"mal_score": -1,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$dayOfWeek",
		"animes": bson.M{
			"$push": "$$ROOT",
		},
	}}

	sortByWeekDay := bson.M{"$sort": bson.M{
		"_id": 1,
	}}

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, addFields, sortByScore, group, sortByWeekDay,
	})
	if err != nil {
		logrus.Error("failed to aggregate currently airing animes: ", err)

		return nil, err
	}

	var currentlyAiringAnimeResponse []responses.CurrentlyAiringAnimeResponse
	if err = cursor.All(context.TODO(), &currentlyAiringAnimeResponse); err != nil {
		logrus.Error("failed to decode currently airing animes: ", err)

		return nil, err
	}

	return currentlyAiringAnimeResponse, nil
}

func (animeModel *AnimeModel) GetAnimesBySortAndFilter(data requests.SortFilterAnime) ([]responses.Anime, p.PaginationData, error) {
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

	match := bson.M{}
	if data.Status != nil || data.Genres != nil || data.Demographics != nil ||
		data.Studios != nil || data.Themes != nil {

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

			match["status"] = status
		}

		if data.Genres != nil {
			match["genres.name"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.Demographics != nil {
			match["demographics.name"] = bson.M{
				"$in": bson.A{data.Demographics},
			}
		}

		if data.Studios != nil {
			match["studios.name"] = bson.M{
				"$in": bson.A{data.Studios},
			}
		}

		if data.Themes != nil {
			match["themes.name"] = bson.M{
				"$in": bson.A{data.Themes},
			}
		}
	}

	var animes []responses.Anime
	paginatedData, err := p.New(animeModel.Collection).Context(context.TODO()).Limit(animePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&animes).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate animes by sort and filter: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get animes by selected filters.")
	}

	return animes, paginatedData.Pagination, nil
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

	cursor, err := animeModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
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

func (animeModel *AnimeModel) SearchAnimeByTitle(data requests.Search) ([]responses.Anime, p.PaginationData, error) {
	search := bson.M{"$search": bson.M{
		"index": "anime_search",
		"text": bson.M{
			"query": data.Search,
			"path":  bson.A{"title_en", "title_jp", "title_original"},
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

	result := animeModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectID,
	})

	var anime responses.Anime
	if err := result.Decode(&anime); err != nil {
		logrus.WithFields(logrus.Fields{
			"anime_id": data.ID,
		}).Error("failed to find anime details by id: ", err)

		return responses.Anime{}, fmt.Errorf("Failed to find anime by id.")
	}

	return anime, nil
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
