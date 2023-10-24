package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"
	"time"

	p "github.com/gobeam/mongo-go-pagination"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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

func (reviewModel *ReviewModel) GetReviewsByContentID(
	uid, contentID string, contentExternalID *string,
	contentExternalIntID *int64, data requests.SortUpcoming,
) ([]responses.Review, p.PaginationData, error) {
	// TODO Aggregation, get author lookup
	// if author is uid, mark it with a boolean
	// get like dislikes
	// get current user with uid liked/disliked ?

	return nil, p.PaginationData{}, nil
}

func (reviewModel *ReviewModel) GetReviewsByUserID(
	uid string, data requests.SortUpcoming,
) ([]responses.Review, p.PaginationData, error) {
	// TODO Aggregation
	// if author return true
	// get like dislikes
	// get current user with uid liked/disliked ?

	return nil, p.PaginationData{}, nil
}

func (reviewModel *ReviewModel) GetBaseReview(reviewID string) (Review, error) {
	objectID, _ := primitive.ObjectIDFromHex(reviewID)

	result := reviewModel.ReviewCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var review Review
	if err := result.Decode(&review); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": reviewID,
		}).Error("failed to find review by id: ", err)

		return Review{}, fmt.Errorf("Failed to find review by id.")
	}

	return review, nil
}

func (reviewModel *ReviewModel) UpdateReview(data requests.UpdateReview, review Review) error {
	objectReviewID, _ := primitive.ObjectIDFromHex(data.ID)

	review.Review = data.Review

	if data.Star != nil {
		review.Star = *data.Star
	}

	if _, err := reviewModel.ReviewCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectReviewID,
	}, bson.M{"$set": review}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectReviewID,
			"data": data,
		}).Error("failed to update review: ", err)

		return fmt.Errorf("Failed to update review.")
	}

	return nil
}

func (reviewModel *ReviewModel) DeleteReviewByID(uid, reviewID string) (bool, error) {
	objectReviewID, _ := primitive.ObjectIDFromHex(reviewID)

	count, err := reviewModel.ReviewCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectReviewID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":       uid,
			"review_id": reviewID,
		}).Error("failed to delete review: ", err)

		return false, fmt.Errorf("Failed to delete review.")
	}

	return count.DeletedCount > 0, nil
}
