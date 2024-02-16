package models

import (
	"app/db"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//lint:file-ignore ST1005 Ignore all

type AISuggestionsModel struct {
	Collection *mongo.Collection
}

func NewAISuggestionsModel(mongoDB *db.MongoDB) *AISuggestionsModel {
	return &AISuggestionsModel{
		Collection: mongoDB.Database.Collection("ai-suggestions"),
	}
}

type AISuggestions struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Movies    []string           `bson:"movies" json:"movies"`
	TVSeries  []string           `bson:"tv" json:"tv"`
	Anime     []string           `bson:"anime" json:"anime"`
	Games     []string           `bson:"game" json:"game"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

func createAISuggestionsObject(userID string, movies, tvSeries, anime, games []string) *AISuggestions {
	return &AISuggestions{
		UserID:    userID,
		Movies:    movies,
		TVSeries:  tvSeries,
		Anime:     anime,
		Games:     games,
		CreatedAt: time.Now().UTC(),
	}
}

func (aiSuggestionsModel *AISuggestionsModel) CreateAISuggestions(uid string, movies, tvSeries, anime, games []string) error {
	aiSuggestions := createAISuggestionsObject(uid, movies, tvSeries, anime, games)

	if _, err := aiSuggestionsModel.Collection.InsertOne(context.TODO(), aiSuggestions); err != nil {
		logrus.WithFields(logrus.Fields{
			"ai_suggestions": aiSuggestions,
		}).Error("failed to create new ai suggestions: ", err)

		return fmt.Errorf("Failed to create ai suggestions.")
	}

	return nil
}

func (aiSuggestionsModel *AISuggestionsModel) GetAISuggestions(uid string) (AISuggestions, error) {
	result := aiSuggestionsModel.Collection.FindOne(context.TODO(), bson.M{
		"user_id": uid,
	})

	var suggestions AISuggestions
	if err := result.Decode(&suggestions); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to find suggestions by user id: ", err)

		return AISuggestions{}, fmt.Errorf("Failed to find suggestions by user id.")
	}

	return suggestions, nil
}

func (aiSuggestionsModel *AISuggestionsModel) DeleteAISuggestionsByUserID(uid string) (bool, error) {
	count, err := aiSuggestionsModel.Collection.DeleteOne(context.TODO(), bson.M{
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete ai suggestion: ", err)

		return false, fmt.Errorf("Failed to delete ai suggestion.")
	}

	return count.DeletedCount > 0, nil
}

func (aiSuggestionsModel *AISuggestionsModel) DeleteAllAISuggestionsByUserID(uid string) error {
	if _, err := aiSuggestionsModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete ai suggestions: ", err)

		return fmt.Errorf("Failed to delete ai suggestions.")
	}

	return nil
}
