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
	paginatedData, err := p.New(gameModel.Collection).Context(context.TODO()).Limit(gameUpcomingPaginationLimit).
		Page(data.Page).Sort(sortType, sortOrder).Filter(match).Decode(&games).Find()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate games by sort and filter", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get games.")
	}

	return games, paginatedData.Pagination, nil
}

//TODO Get user's is listed etc. values
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
