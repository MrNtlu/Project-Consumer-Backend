package models

import (
	"app/db"
	"app/requests"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewModel struct {
	ReviewCollection *mongo.Collection
}

func NewReviewModel(mongoDB *db.MongoDB) *ReviewModel {
	return &ReviewModel{
		ReviewCollection: mongoDB.Database.Collection("reviews"),
	}
}

type Review struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID               string             `bson:"user_id" json:"user_id"`
	ContentID            string             `bson:"content_id" json:"content_id"`
	ContentExternalID    *string            `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64             `bson:"content_external_int_id" json:"content_external_int_id"`
	Star                 int8               `bson:"star" json:"star"`
	Review               *string            `bson:"review" json:"review"`
	Likes                []string           `bson:"likes" json:"likes"`
	Dislikes             []string           `bson:"dislikes" json:"dislikes"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

func createReviewObject(
	userID, contentID string, contentExternalID, review *string,
	contentExternalIntID *int64, star int8,
) *Review {
	emptyList := []string{}

	return &Review{
		UserID:               userID,
		ContentID:            contentID,
		ContentExternalID:    contentExternalID,
		ContentExternalIntID: contentExternalIntID,
		Star:                 star,
		Review:               review,
		Likes:                emptyList,
		Dislikes:             emptyList,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
}

func (reviewModel *ReviewModel) CreateReview(uid string, data requests.CreateReview) (Review, error) {
	review := createReviewObject(
		uid,
		data.ContentID,
		data.ContentExternalID,
		data.Review,
		data.ContentExternalIntID,
		data.Star,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = reviewModel.ReviewCollection.InsertOne(context.TODO(), review); err != nil {
		logrus.WithFields(logrus.Fields{
			"review": review,
		}).Error("failed to create new review: ", err)

		return Review{}, fmt.Errorf("Failed to create review.")
	}

	review.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *review, nil
}
