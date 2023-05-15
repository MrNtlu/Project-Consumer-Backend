package models

import (
	"app/db"
	"app/requests"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserInteractionModel struct {
	ConsumeLaterCollection *mongo.Collection
}

func NewUserInteractionModel(mongoDB *db.MongoDB) *UserInteractionModel {
	return &UserInteractionModel{
		ConsumeLaterCollection: mongoDB.Database.Collection("consume-laters"),
	}
}

/**
* Consume Later
* Recommendation
* ?	- Agree/Disagree to recommendation. Collection or array?
* Suggest Similar Content
* Suggest Genre/Tag
* *Later
* - Level System
**/

type ConsumeLaterList struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID               string             `bson:"user_id" json:"user_id"`
	ContentID            string             `bson:"content_id" json:"content_id"`
	ContentExternalID    *string            `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64             `bson:"content_external_int_id" json:"content_external_int_id"`
	ContentType          string             `bson:"content_type" json:"content_type"` // anime, movie, tvseries or game
	SelfNote             *string            `bson:"self_note" json:"self_note"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
}

func createConsumeLaterObject(userID, contentID, contentType string, contentExternalID, selfNote *string, contentExternalIntID *int64) *ConsumeLaterList {
	return &ConsumeLaterList{
		UserID:               userID,
		ContentID:            contentID,
		ContentType:          contentType,
		ContentExternalID:    contentExternalID,
		ContentExternalIntID: contentExternalIntID,
		SelfNote:             selfNote,
		CreatedAt:            time.Now().UTC(),
	}
}

// TODO Get consume later list by x order

func (userInteractionModel *UserInteractionModel) CreateConsumeLater(uid string, data requests.CreateConsumeLater) error {
	consumeLater := createConsumeLaterObject(
		uid,
		data.ContentID,
		data.ContentType,
		data.ContentExternalID,
		data.SelfNote,
		data.ContentExternalIntID,
	)

	if _, err := userInteractionModel.ConsumeLaterCollection.InsertOne(context.TODO(), consumeLater); err != nil {
		logrus.WithFields(logrus.Fields{
			"consume_later": consumeLater,
		}).Error("failed to create new consume later: ", err)

		return fmt.Errorf("Failed to create consume later.")
	}

	return nil
}

func (userInteractionModel *UserInteractionModel) UpdateConsumeLaterSelfNote(data requests.UpdateConsumeLater, consumeLater ConsumeLaterList) error {
	objectConsumeLaterID, _ := primitive.ObjectIDFromHex(data.ID)

	consumeLater.SelfNote = data.SelfNote

	if _, err := userInteractionModel.ConsumeLaterCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectConsumeLaterID,
	}, bson.M{"$set": consumeLater}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectConsumeLaterID,
			"data": data,
		}).Error("failed to update consume later: ", err)

		return fmt.Errorf("Failed to update consume later.")
	}

	return nil
}

func (userInteractionModel *UserInteractionModel) DeleteConsumeLaterByID(uid, consumeLaterID string) (bool, error) {
	objectConsumeLaterID, _ := primitive.ObjectIDFromHex(consumeLaterID)

	count, err := userInteractionModel.ConsumeLaterCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectConsumeLaterID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":              uid,
			"consume_later_id": consumeLaterID,
		}).Error("failed to delete consume later: ", err)

		return false, fmt.Errorf("Failed to delete consume later.")
	}

	return count.DeletedCount > 0, nil
}

func (userInteractionModel *UserInteractionModel) DeleteAllConsumeLaterByUserID(uid string) error {
	if _, err := userInteractionModel.ConsumeLaterCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all consume laters by user id: ", err)

		return fmt.Errorf("Failed to delete all consume laters by user.")
	}

	return nil
}
