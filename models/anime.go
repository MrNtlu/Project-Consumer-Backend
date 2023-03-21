package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"
	"time"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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
	animeUpcomingPaginationLimit = 20
)

/* TODO Endpoints
* [] Get upcoming by popularity etc.
* [] Get by season/year
* [] Get currently airing animes by day
* [] Get anime by popularity, genre etc.
* [] Get anime details
 */

func (animeModel *AnimeModel) GetUpcomingAnimesBySort(data requests.SortUpcomingAnime) ([]responses.Anime, p.PaginationData, error) {
	currentSeason := getSeasonFromMonth()

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "popularity"
		sortOrder = data.SortOrder
	case "date":
		sortType = "aired.from"
		sortOrder = data.SortOrder
	}

	match := bson.M{"$match": bson.M{
		"is_airing": false,
		"$or": bson.A{
			bson.M{"status": "Not yet aired"},
			bson.M{"aired.from": bson.M{"$gte": time.Now().UTC()}},
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
		Page(data.Page).Sort("has_year", -1).Sort(sortType, sortOrder).Sort("_id", 1).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming animes: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming animes.")
	}

	var lists []responses.Anime
	for _, raw := range paginatedData.Data {
		var product *responses.Anime
		if marshallErr := bson.Unmarshal(raw, &product); marshallErr == nil {
			lists = append(lists, *product)
		}

	}

	return lists, paginatedData.Pagination, nil
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
