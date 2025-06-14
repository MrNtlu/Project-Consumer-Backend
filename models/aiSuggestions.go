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
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore ST1005 Ignore all

type AISuggestionsModel struct {
	Collection              *mongo.Collection
	NotInterestedCollection *mongo.Collection
}

func NewAISuggestionsModel(mongoDB *db.MongoDB) *AISuggestionsModel {
	return &AISuggestionsModel{
		Collection:              mongoDB.Database.Collection("ai-suggestions"),
		NotInterestedCollection: mongoDB.Database.Collection("ai-suggestions-not-interested"),
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

type NotInterested struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	ContentID   string             `bson:"content_id" json:"content_id"`
	ContentType string             `bson:"content_type" json:"content_type"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

func createNotInterestedObject(userID, contentID, contentType string) *NotInterested {
	return &NotInterested{
		UserID:      userID,
		ContentID:   contentID,
		ContentType: contentType,
		CreatedAt:   time.Now().UTC(),
	}
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

func (aiSuggestionsModel *AISuggestionsModel) CreateNotInterested(uid, contentID, contentType string) error {
	notInterested := createNotInterestedObject(uid, contentID, contentType)
	if _, err := aiSuggestionsModel.NotInterestedCollection.InsertOne(context.TODO(), notInterested); err != nil {
		logrus.WithFields(logrus.Fields{
			"not_interested": notInterested,
		}).Error("failed to create not interested: ", err)
		return fmt.Errorf("Failed to create not interested.")
	}

	return nil
}

func (aiSuggestionsModel *AISuggestionsModel) DeleteNotInterested(uid, contentID, contentType string) error {
	if _, err := aiSuggestionsModel.NotInterestedCollection.DeleteOne(context.TODO(), bson.M{
		"user_id":      uid,
		"content_id":   contentID,
		"content_type": contentType,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":          uid,
			"content_id":   contentID,
			"content_type": contentType,
		}).Error("failed to delete not interested: ", err)

		return fmt.Errorf("Failed to delete not interested.")
	}

	return nil
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

func (aiSuggestionsModel *AISuggestionsModel) GetAllNotInterestedByUserID(uid string) ([]NotInterested, error) {
	cursor, err := aiSuggestionsModel.NotInterestedCollection.Find(context.TODO(), bson.M{
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to find not interested by user id: ", err)
		return nil, fmt.Errorf("Failed to find not interested by user id.")
	}

	var notInterested []NotInterested
	if err := cursor.All(context.TODO(), &notInterested); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to find not interested by user id: ", err)

		return nil, fmt.Errorf("Failed to find not interested by user id.")
	}

	return notInterested, nil
}

func (aiSuggestionsModel *AISuggestionsModel) GetAISuggestions(uid string) (AISuggestions, error) {
	opts := options.FindOne().SetSort(bson.M{"created_at": -1})
	result := aiSuggestionsModel.Collection.FindOne(context.TODO(), bson.M{
		"user_id": uid,
	}, opts)

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
