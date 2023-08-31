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
	gameUpcomingPaginationLimit = 40
	gamePaginationLimit         = 40
)

func (gameModel *GameModel) GetUpcomingGamesBySort(data requests.Pagination) ([]responses.Game, p.PaginationData, error) {
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

	addPopularityFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$rawg_rating", "$rawg_rating_count",
			},
		},
	}}

	paginatedData, err := p.New(gameModel.Collection).Context(context.TODO()).Limit(gameUpcomingPaginationLimit).
		Page(data.Page).Sort("has_release_date", -1).Sort("popularity", -1).Sort("_id", 1).Aggregate(match, addFields, addPopularityFields)
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
	addFields := bson.M{"$addFields": bson.M{
		"popularity": bson.M{
			"$multiply": bson.A{
				"$rawg_rating", "$rawg_rating_count",
			},
		},
	}}

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "popularity"
		sortOrder = -1
	case "top":
		sortType = "metacritic_score"
		sortOrder = -1
	case "new":
		sortType = "release_date"
		sortOrder = -1
	case "old":
		sortType = "release_date"
		sortOrder = 1
	}

	matchFields := bson.M{}
	if data.Genres != nil || data.Platform != nil || data.TBA != nil {
		if data.Genres != nil {
			matchFields["genres"] = bson.M{
				"$in": bson.A{data.Genres},
			}
		}

		if data.Platform != nil {
			matchFields["platforms"] = bson.M{
				"$in": bson.A{data.Platform},
			}
		}

		if data.TBA != nil {
			matchFields["tba"] = data.TBA
		}
	}

	match := bson.M{"$match": matchFields}

	paginatedData, err := p.New(gameModel.Collection).Context(context.TODO()).Limit(gamePaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(match, addFields)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
			"match":   match,
		}).Error("failed to aggregate games by sort and filter", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get games.")
	}

	var games []responses.Game
	for _, raw := range paginatedData.Data {
		var game *responses.Game
		if marshalErr := bson.Unmarshal(raw, &game); marshalErr == nil {
			games = append(games, *game)
		}
	}

	return games, paginatedData.Pagination, nil
}

func (gameModel *GameModel) GetGameDetails(data requests.ID) (responses.GameDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)
	rawgID, _ := strconv.Atoi(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"rawg_id": rawgID,
			},
		},
	}}

	relatedGamesLookup := bson.M{"$lookup": bson.M{
		"from": "games",
		"let": bson.M{
			"rawg_id": "$related_games.rawg_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$in": bson.A{
							"$rawg_id", "$$rawg_id",
						},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id":            1,
					"title":          1,
					"title_original": 1,
					"rawg_id":        1,
					"image_url":      1,
					"platforms":      1,
					"release_date":   1,
				},
			},
		},
		"as": "related_games",
	}}

	sortRelatedGames := bson.M{"$set": bson.M{
		"related_games": bson.M{
			"$sortArray": bson.M{
				"input": "$related_games",
				"sortBy": bson.M{
					"release_date": -1,
				},
			},
		},
	}}

	cursor, err := gameModel.Collection.Aggregate(context.TODO(), bson.A{
		match, relatedGamesLookup, sortRelatedGames,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to aggregate game details: ", err)

		return responses.GameDetails{}, fmt.Errorf("Failed to aggregate game details with watch list.")
	}

	var gameDetails []responses.GameDetails
	if err = cursor.All(context.TODO(), &gameDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": data.ID,
		}).Error("failed to decode game details: ", err)

		return responses.GameDetails{}, fmt.Errorf("Failed to decode game details.")
	}

	if len(gameDetails) > 0 {
		return gameDetails[0], nil
	}

	return responses.GameDetails{}, nil
}

func (gameModel *GameModel) GetGameDetailsWithPlayList(data requests.ID, uuid string) (responses.GameDetails, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)
	rawgID, _ := strconv.Atoi(data.ID)

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"_id": objectID,
			},
			bson.M{
				"rawg_id": rawgID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"game_id": bson.M{
			"$toString": "$_id",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from": "game-lists",
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

	lookupWatchLater := bson.M{"$lookup": bson.M{
		"from": "consume-laters",
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
									bson.M{"$eq": bson.A{"$content_id", "$$game_id"}},
									bson.M{"$eq": bson.A{"$content_external_id", "$$rawg_id"}},
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

	relatedGamesLookup := bson.M{"$lookup": bson.M{
		"from": "games",
		"let": bson.M{
			"rawg_id": "$related_games.rawg_id",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$in": bson.A{
							"$rawg_id", "$$rawg_id",
						},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id":            1,
					"title":          1,
					"title_original": 1,
					"rawg_id":        1,
					"image_url":      1,
					"platforms":      1,
					"release_date":   1,
				},
			},
		},
		"as": "related_games",
	}}

	sortRelatedGames := bson.M{"$set": bson.M{
		"related_games": bson.M{
			"$sortArray": bson.M{
				"input": "$related_games",
				"sortBy": bson.M{
					"release_date": -1,
				},
			},
		},
	}}

	cursor, err := gameModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwindWatchList, lookupWatchLater, unwindWatchLater,
		relatedGamesLookup, sortRelatedGames,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to aggregate authenticated game details: ", err)

		return responses.GameDetails{}, fmt.Errorf("Failed to aggregate game details with play list.")
	}

	var gameDetails []responses.GameDetails
	if err = cursor.All(context.TODO(), &gameDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uuid,
			"id":  data.ID,
		}).Error("failed to decode authenticated game details: ", err)

		return responses.GameDetails{}, fmt.Errorf("Failed to decode game details with play list.")
	}

	if len(gameDetails) > 0 {
		return gameDetails[0], nil
	}

	return responses.GameDetails{}, nil
}
