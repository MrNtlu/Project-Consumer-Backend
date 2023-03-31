package models

import (
	"app/db"
	"app/responses"
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MovieModel struct {
	Collection *mongo.Collection
}

func NewMovieModel(mongoDB *db.MongoDB) *MovieModel {
	return &MovieModel{
		Collection: mongoDB.Database.Collection("movies"),
	}
}

// const (
// 	movieUpcomingPaginationLimit = 20
// )

func (movieModel *MovieModel) GetMovies() []responses.Movie {
	cursor, _ := movieModel.Collection.Aggregate(context.TODO(), bson.A{})

	var movies []responses.Movie
	if err := cursor.All(context.TODO(), &movies); err != nil {
		logrus.Error("failed to decode movies", err)

		return nil
	}

	return movies
}
