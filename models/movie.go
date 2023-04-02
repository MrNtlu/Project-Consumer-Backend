package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

/* TODO Endpoints
* [ ] Get upcoming movies by popularity etc.
* [ ] Get movies by release date, popularity, genre etc. (sort & filter)
* [ ] Get movie details
* [ ] Get top movies by every decade 1980's 1990's etc.
* [ ] Get top movies by every genre (?)
* [ ]
 */

func (movieModel *MovieModel) GetMovieDetails(data requests.ID) (responses.Movie, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	result := movieModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectID,
	})

	var movie responses.Movie
	if err := result.Decode(&movie); err != nil {
		logrus.WithFields(logrus.Fields{
			"game_id": data.ID,
		}).Error("failed to find movie details by id: ", err)

		return responses.Movie{}, fmt.Errorf("Failed to find movie by id.")
	}

	return movie, nil
}
