package models

import (
	"app/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LogsModel struct {
	LogsCollection *mongo.Collection
}

func NewLogsModel(mongoDB *db.MongoDB) *LogsModel {
	return &LogsModel{
		LogsCollection: mongoDB.Database.Collection("logs"),
	}
}

type Log struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID string             `bson:"user_id" json:"user_id"`
}

type Diary struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID string             `bson:"user_id" json:"user_id"`
}
