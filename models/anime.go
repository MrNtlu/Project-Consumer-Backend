package models

import (
	"app/db"

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

/* TODO Endpoints
* [] Get upcoming by popularity etc.
* [] Get by season
* [] Get currently airing animes by day
* [] Get anime by popularity, genre etc.
* [] Get anime details
 */
