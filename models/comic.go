package models

import (
	"app/db"

	"go.mongodb.org/mongo-driver/mongo"
)

type ComicModel struct {
	Collection *mongo.Collection
}

func NewComicModel(mongoDB *db.MongoDB) *ComicModel {
	return &ComicModel{
		Collection: mongoDB.Database.Collection("comic-books"),
	}
}

const (
	comicSearchLimit     = 50
	comicPaginationLimit = 40
)
