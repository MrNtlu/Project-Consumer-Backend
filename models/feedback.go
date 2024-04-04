package models

import (
	"app/db"
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FeedbackModel struct {
	FeedbackCollection *mongo.Collection
}

func NewFeedbackModel(mongoDB *db.MongoDB) *FeedbackModel {
	return &FeedbackModel{
		FeedbackCollection: mongoDB.Database.Collection("feedbacks"),
	}
}

type Feedback struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID string             `bson:"user_id" json:"user_id"`
}

func createFeedback(userID string) *Feedback {
	return &Feedback{
		UserID: userID,
	}
}

func (feedbackModel *FeedbackModel) CreateFeedback(uid string) error {
	log := createFeedback(
		uid,
	)

	if _, err := feedbackModel.FeedbackCollection.InsertOne(context.TODO(), log); err != nil {
		logrus.WithFields(logrus.Fields{
			"log": log,
		}).Error("failed to create new feedback: ", err)

		return err
	}

	return nil
}
