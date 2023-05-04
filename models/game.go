package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"app/utils"
	"context"
	"fmt"
	"time"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameModel struct {
	Collection *mongo.Collection
}

func NewGameModel(mongoDB *db.MongoDB) *GameModel {
	return &GameModel{
		Collection: mongoDB.Database.Collection("games"),
	}
}

const (
	gameUpcomingPaginationLimit = 20
	gamePaginationLimit         = 20
)

/* TODO Endpoints
* [x] Get upcoming by popularity etc.
* [x] Get games by release date, popularity, genre, platform etc.
* [x] Get game details
 */

func (gameModel *GameModel) GetUpcomingGamesBySort(data requests.SortUpcoming) ([]responses.Game, p.PaginationData, error) {
	var (
		sortType            string
		sortOrder           int8
		hasReleaseDateOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "rawg_rating"
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
		"$or": bson.A{
			bson.M{"tba": true},
			bson.M{"release_date": bson.M{"$gte": utils.GetCurrentDate()}},
		},
	}}

	addFields := bson.M{"$addFields": bson.M{
		"has_release_date": bson.M{
			"$and": bson.A{
				bson.M{
					"$ne": bson.A{"$release_date", nil},
				},
				bson.M{
					"$ne": bson.A{"$release_date", ""},
				},
			},
		},
	}}

	fmt.Println(time.Now().UTC(), utils.GetCurrentDate())

	paginatedData, err := p.New(gameModel.Collection).Context(context.TODO()).Limit(gameUpcomingPaginationLimit).
		Page(data.Page).Sort("has_release_date", hasReleaseDateOrder).Sort(sortType, sortOrder).Sort("_id", 1).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate upcoming games: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get upcoming games.")
	}

	var upcomingGames []responses.Game
	for _, raw := range paginatedData.Data {
		var game *responses.Game
		if marshalErr := bson.Unmarshal(raw, &game); marshalErr == nil {
			upcomingGames = append(upcomingGames, *game)
		}
	}

	return upcomingGames, paginatedData.Pagination, nil
}

func (gameModel *GameModel) GetGamesByFilterAndSort(data requests.SortFilterGame) ([]responses.Game, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		if data.TBA != nil && *data.TBA {
			sortType = "rawg_rating"
		} else {
			sortType = "metacritic_score"
		}
		sortOrder = -1
	case "new":
		sortType = "release_date"
		sortOrder = -1
	case "old":
		sortType = "release_date"
		sortOrder = 1
	}

	match := bson.M{}
	if data.Genres != nil || data.Platform != nil || data.TBA != nil {
		if data.Genres != nil {
			match["genres.name"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.Platform != nil {
			match["platforms"] = bson.M{
				"$in": bson.A{data.Platform},
			}
		}

		if data.TBA != nil {
			match["tba"] = data.TBA
		}
	}

	var games []responses.Game
	paginatedData, err := p.New(gameModel.Collection).Context(context.TODO()).Limit(gamePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&games).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate games by sort and filter", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get games.")
	}

	return games, paginatedData.Pagination, nil
}

func (gameModel *GameModel) GetGameDetails(data requests.ID) (responses.Game, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	result := gameModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectID,
	})

	var game responses.Game
	if err := result.Decode(&game); err != nil {
		logrus.WithFields(logrus.Fields{
			"game_id": data.ID,
		}).Error("failed to find game details by id: ", err)

		return responses.Game{}, fmt.Errorf("Failed to find game by id.")
	}

	return game, nil
}

func (gameModel *GameModel) GetGameDetailsWithPlayList(data requests.ID, uuid string) (responses.GameDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	match := bson.M{"$match": bson.M{
		"_id": objectID,
	}}

	set := bson.M{"$set": bson.M{
		"game_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from": "movie-watch-lists",
		"let": bson.M{
			"uuid":    uuid,
			"game_id": "$game_id",
			"rawg_id": "$rawg_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{
								"$or": bson.A{
									bson.M{"$eq": bson.A{"$game_id", "$$game_id"}},
									bson.M{"$eq": bson.A{"$game_rawg_id", "$$rawg_id"}},
								},
							},
							bson.M{"$eq": bson.A{"$user_id", "$$uuid"}},
						},
					},
				},
			},
		},
		"as": "game_list",
	}}

	unwindWatchList := bson.M{"$unwind": bson.M{
		"path":                       "$game_list",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	cursor, err := gameModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate game details: ", err)

		return responses.GameDetails{}, fmt.Errorf("Failed to aggregate game details with watch list.")
	}

	var gameDetails []responses.GameDetails
	if err = cursor.All(context.TODO(), &gameDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to decode game details: ", err)

		return responses.GameDetails{}, fmt.Errorf("Failed to decode game details.")
	}

	if len(gameDetails) > 0 {
		return gameDetails[0], nil
	}

	return responses.GameDetails{}, nil
}
