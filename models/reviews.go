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

const reviewPagination = 25

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
	ContentType          string             `bson:"content_type" json:"content_type"` // anime, movie, tv or game
	Star                 int8               `bson:"star" json:"star"`
	Review               string             `bson:"review" json:"review"`
	Likes                []string           `bson:"likes" json:"likes"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

func createReviewObject(
	userID, contentID, contentType, review string, contentExternalID *string,
	contentExternalIntID *int64, star int8,
) *Review {
	return &Review{
		UserID:               userID,
		ContentID:            contentID,
		ContentExternalID:    contentExternalID,
		ContentExternalIntID: contentExternalIntID,
		ContentType:          contentType,
		Star:                 star,
		Review:               review,
		Likes:                []string{},
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
}

func convertReviewModelToResponse(review Review) responses.Review {
	return responses.Review{
		ID:                   review.ID,
		UserID:               review.UserID,
		ContentID:            review.ContentID,
		ContentExternalID:    review.ContentExternalID,
		ContentExternalIntID: review.ContentExternalIntID,
		Star:                 review.Star,
		IsAuthor:             true,
		IsLiked:              false,
		Review:               review.Review,
		Popularity:           int64(len(review.Likes)),
		Likes:                review.Likes,
		CreatedAt:            review.CreatedAt,
		UpdatedAt:            review.UpdatedAt,
	}
}

func convertReviewResponseToModel(review responses.Review) Review {
	return Review{
		ID:                   review.ID,
		UserID:               review.UserID,
		ContentID:            review.ContentID,
		ContentExternalID:    review.ContentExternalID,
		ContentExternalIntID: review.ContentExternalIntID,
		Star:                 review.Star,
		Review:               review.Review,
		Likes:                review.Likes,
		CreatedAt:            review.CreatedAt,
		UpdatedAt:            review.UpdatedAt,
	}
}

func (reviewModel *ReviewModel) CreateReview(uid string, data requests.CreateReview) (responses.Review, error) {
	review := createReviewObject(
		uid,
		data.ContentID,
		data.ContentType,
		data.Review,
		data.ContentExternalID,
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

		return responses.Review{}, fmt.Errorf("Failed to create review.")
	}

	review.ID = insertedID.InsertedID.(primitive.ObjectID)

	return convertReviewModelToResponse(*review), nil
}

func (reviewModel *ReviewModel) GetReviewSummaryForDetails(contentID, uid string, contentExternalID *string, contentExternalIntID *int64) (responses.ReviewSummary, error) {
	var (
		cID  string
		ceID int64
	)

	if contentExternalID != nil {
		cID = *contentExternalID
	} else {
		cID = "-1"
	}

	if contentExternalIntID != nil {
		ceID = *contentExternalIntID
	} else {
		ceID = -1
	}

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"content_id": contentID,
			},
			bson.M{
				"content_external_id": cID,
			},
			bson.M{
				"content_external_int_id": ceID,
			},
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$content_id",
		"avg_star": bson.M{
			"$avg": "$star",
		},
		"total_votes": bson.M{
			"$sum": 1,
		},
		"is_reviewed": bson.M{
			"$sum": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{"$user_id", uid}},
					"then": 1,
					"else": 0,
				},
			},
		},
		"one_star": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$star", 1},
					}, 1, 0,
				},
			},
		},
		"two_star": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$star", 2},
					}, 1, 0,
				},
			},
		},
		"three_star": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$star", 3},
					}, 1, 0,
				},
			},
		},
		"four_star": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$star", 4},
					}, 1, 0,
				},
			},
		},
		"five_star": bson.M{
			"$sum": bson.M{
				"$cond": bson.A{
					bson.M{
						"$eq": bson.A{"$star", 5},
					}, 1, 0,
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"avg_star":    1,
		"total_votes": 1,
		"is_reviewed": bson.M{
			"$cond": bson.M{
				"if":   bson.M{"$gte": bson.A{"$is_reviewed", 1}},
				"then": true,
				"else": false,
			},
		},
		"star_counts": bson.M{
			"one_star":   "$one_star",
			"two_star":   "$two_star",
			"three_star": "$three_star",
			"four_star":  "$four_star",
			"five_star":  "$five_star",
		},
	}}

	cursor, err := reviewModel.ReviewCollection.Aggregate(context.TODO(), bson.A{match, group, project})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"content_id":              contentID,
			"content_external_id":     contentExternalID,
			"context_external_int_id": contentExternalIntID,
		}).Error("failed to aggregate review summary: ", err)

		return responses.ReviewSummary{}, fmt.Errorf("Failed to aggregate review summary.")
	}

	var reviewSummary []responses.ReviewSummary
	if err = cursor.All(context.TODO(), &reviewSummary); err != nil {
		logrus.WithFields(logrus.Fields{
			"content_id":              contentID,
			"content_external_id":     contentExternalID,
			"context_external_int_id": contentExternalIntID,
		}).Error("failed to decode review summary: ", err)

		return responses.ReviewSummary{}, fmt.Errorf("Failed to decode review summary.")
	}

	if len(reviewSummary) > 0 {
		return reviewSummary[0], nil
	}

	return responses.ReviewSummary{}, nil
}

func (reviewModel *ReviewModel) GetReviewsByContentID(data requests.SortReviewByContentID) ([]responses.Review, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "popularity"
		sortOrder = -1
	case "latest":
		sortType = "created_at"
		sortOrder = -1
	case "oldest":
		sortType = "created_at"
		sortOrder = 1
	}

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"content_id": data.ContentID,
			},
			bson.M{
				"content_external_id": data.ContentExternalID,
			},
			bson.M{
				"content_external_int_id": data.ContentExternalIntID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"popularity": bson.M{
			"$size": "$likes",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "obj_user_id",
		"foreignField": "_id",
		"as":           "author",
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$author",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	paginatedData, err := p.New(reviewModel.ReviewCollection).Context(context.TODO()).Limit(reviewPagination).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(
		match, set, lookup, unwind,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate reviews", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get reviews.")
	}

	var reviews []responses.Review
	for _, raw := range paginatedData.Data {
		var review *responses.Review
		if marshalErr := bson.Unmarshal(raw, &review); marshalErr == nil {
			reviews = append(reviews, *review)
		}
	}

	return reviews, paginatedData.Pagination, nil
}

func (reviewModel *ReviewModel) GetReviewsByContentIDAndUserID(
	uid string, data requests.SortReviewByContentID,
) ([]responses.Review, p.PaginationData, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "popularity"
		sortOrder = -1
	case "latest":
		sortType = "created_at"
		sortOrder = -1
	case "oldest":
		sortType = "created_at"
		sortOrder = 1
	}

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"content_id": data.ContentID,
			},
			bson.M{
				"content_external_id": data.ContentExternalID,
			},
			bson.M{
				"content_external_int_id": data.ContentExternalIntID,
			},
		},
	}}

	set := bson.M{"$set": bson.M{
		"is_author": bson.M{
			"$eq": bson.A{
				"$user_id", uid,
			},
		},
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"popularity": bson.M{
			"$size": "$likes",
		},
		"is_liked": bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$likes",
					},
				},
				"then": true,
				"else": false,
			},
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "obj_user_id",
		"foreignField": "_id",
		"as":           "author",
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$author",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	paginatedData, err := p.New(reviewModel.ReviewCollection).Context(context.TODO()).Limit(reviewPagination).
		Page(data.Page).Sort("is_author", -1).Sort(sortType, sortOrder).Aggregate(
		match, set, lookup, unwind,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":     uid,
			"request": data,
		}).Error("failed to aggregate reviews by user id", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get reviews by user id.")
	}

	var reviews []responses.Review
	for _, raw := range paginatedData.Data {
		var review *responses.Review
		if marshalErr := bson.Unmarshal(raw, &review); marshalErr == nil {
			reviews = append(reviews, *review)
		}
	}

	return reviews, paginatedData.Pagination, nil
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

func (reviewModel *ReviewModel) GetBaseReview(uid, reviewID string) (Review, error) {
	objectID, _ := primitive.ObjectIDFromHex(reviewID)

	result := reviewModel.ReviewCollection.FindOne(context.TODO(), bson.M{
		"_id":     objectID,
		"user_id": uid,
	})

	var review Review
	if err := result.Decode(&review); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
			"id":      reviewID,
		}).Error("failed to find review by id: ", err)

		return Review{}, fmt.Errorf("Failed to find review by id.")
	}

	return review, nil
}

func (reviewModel *ReviewModel) GetBaseReviewResponse(uid, reviewID string) (responses.Review, error) {
	objectID, _ := primitive.ObjectIDFromHex(reviewID)

	match := bson.M{"$match": bson.M{
		"_id": objectID,
	}}

	set := bson.M{"$set": bson.M{
		"is_author": bson.M{
			"$eq": bson.A{
				"$user_id", uid,
			},
		},
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"popularity": bson.M{
			"$size": "$likes",
		},
		"is_liked": bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$likes",
					},
				},
				"then": true,
				"else": false,
			},
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "obj_user_id",
		"foreignField": "_id",
		"as":           "author",
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$author",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	cursor, err := reviewModel.ReviewCollection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwind,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":  reviewID,
			"uid": uid,
		}).Error("failed to aggregate review: ", err)

		return responses.Review{}, fmt.Errorf("Failed to aggregate review.")
	}

	var review []responses.Review
	if err = cursor.All(context.TODO(), &review); err != nil {
		logrus.WithFields(logrus.Fields{
			"id":  reviewID,
			"uid": uid,
		}).Error("failed to decode review: ", err)

		return responses.Review{}, fmt.Errorf("Failed to decode review.")
	}

	if len(review) > 0 {
		return review[0], nil
	}

	return responses.Review{}, nil
}

func (reviewModel *ReviewModel) VoteReview(uid string, data requests.ID, review responses.Review) (responses.Review, error) {
	objectReviewID, _ := primitive.ObjectIDFromHex(data.ID)

	var isAlreadyLiked = false

	for _, like := range review.Likes {
		if like == uid {
			isAlreadyLiked = true
			review.IsLiked = false
			review.Popularity = review.Popularity - 1
			review.Likes = removeElement(review.Likes, uid)
		}
	}

	if !isAlreadyLiked {
		review.Likes = append(review.Likes, uid)
		review.Popularity = review.Popularity + 1
		review.IsLiked = true
	}

	updatedReview := convertReviewResponseToModel(review)
	if _, err := reviewModel.ReviewCollection.UpdateOne(context.TODO(), bson.M{
		"_id": objectReviewID,
	}, bson.M{"$set": updatedReview}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectReviewID,
			"data": data,
		}).Error("failed to update like review: ", err)

		return responses.Review{}, fmt.Errorf("Failed to like.")
	}

	return review, nil
}

func (reviewModel *ReviewModel) UpdateReview(uid string, data requests.UpdateReview, review responses.Review) (responses.Review, error) {
	objectReviewID, _ := primitive.ObjectIDFromHex(data.ID)

	if data.Review != nil {
		review.Review = *data.Review
	}

	if data.Star != nil {
		review.Star = *data.Star
	}

	updatedReview := convertReviewResponseToModel(review)
	if _, err := reviewModel.ReviewCollection.UpdateOne(context.TODO(), bson.M{
		"_id":     objectReviewID,
		"user_id": uid,
	}, bson.M{"$set": updatedReview}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectReviewID,
			"data": data,
		}).Error("failed to update review: ", err)

		return responses.Review{}, fmt.Errorf("Failed to update review.")
	}

	return review, nil
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

func removeElement(slice []string, element string) []string {
	index := -1
	for i, item := range slice {
		if item == element {
			index = i
			break
		}
	}

	if index == -1 {
		return slice
	}

	return append(slice[:index], slice[index+1:]...)
}
