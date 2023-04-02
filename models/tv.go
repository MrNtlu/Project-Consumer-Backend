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

type TVModel struct {
	Collection *mongo.Collection
}

func NewTVModel(mongoDB *db.MongoDB) *TVModel {
	return &TVModel{
		Collection: mongoDB.Database.Collection("tv-series"),
	}
}

/* TODO Endpoints
* [ ] Get upcoming tv series by popularity etc.
* [ ] Get tv series by release date, popularity, genre etc. (sort & filter)
* [ ] Get tv series details
* [ ] Get top tv series by every decade 1980's 1990's etc.
* [ ] Get top tv series by every genre (?)
* [ ]
 */

func (tvModel *TVModel) GetTVSeriesDetails(data requests.ID) (responses.TVSeries, error) {
	objectID, _ := primitive.ObjectIDFromHex(data.ID)

	result := tvModel.Collection.FindOne(context.TODO(), bson.M{
		"_id": objectID,
	})

	var tvSeries responses.TVSeries
	if err := result.Decode(&tvSeries); err != nil {
		logrus.WithFields(logrus.Fields{
			"game_id": data.ID,
		}).Error("failed to find tv series details by id: ", err)

		return responses.TVSeries{}, fmt.Errorf("Failed to find tv series by id.")
	}

	return tvSeries, nil
}
