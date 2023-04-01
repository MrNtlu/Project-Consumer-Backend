package models

import (
	"app/db"

	"go.mongodb.org/mongo-driver/mongo"
)

type TVModel struct {
	Collection *mongo.Collection
}

func NewTVModel(mongoDB *db.MongoDB) *TVModel {
	return &TVModel{
		Collection: mongoDB.Database.Collection("tv-series"),
	}
}
