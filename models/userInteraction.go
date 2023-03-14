package models

import (
	"app/db"
	"time"

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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	ContentID   string             `bson:"content_id" json:"content_id"`
	ContentType string             `bson:"content_type" json:"content_type"` // anime, movie, tvseries or game
	SelfNote    *string            `bson:"self_note" json:"self_note"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

func createConsumeLaterObject(userID, contentID, contentType string, selfNote *string) *ConsumeLaterList {
	return &ConsumeLaterList{
		UserID:      userID,
		ContentID:   contentID,
		ContentType: contentType,
		SelfNote:    selfNote,
		CreatedAt:   time.Now().UTC(),
	}
}

/* TODO
* Add new consume later
* Get consume later list by x order
* Update self note
* Delete consume later
* Erase all list
 */
