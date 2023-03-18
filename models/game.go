package models

import (
	"app/db"

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

/* TODO Endpoints
* [] Get upcoming by popularity etc.
* [] Get games by release date, popularity, genre, platform etc.
* [] Get game details
 */
