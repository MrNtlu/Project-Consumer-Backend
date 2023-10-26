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

func (reviewModel *ReviewModel) GetReviewSummaryForDetails(contentID string, uid, contentExternalID *string, contentExternalIntID *int64) (responses.ReviewSummary, error) {
	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"content_id": contentID,
			},
			bson.M{
				"content_external_id": contentExternalID,
			},
			bson.M{
				"content_external_int_id": contentExternalIntID,
			},
		},
	}}

	var userID string
	if uid != nil {
		userID = *uid
	} else {
		userID = "-1"
	}

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
					"if":   bson.M{"$eq": bson.A{"$user_id", userID}},
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
				"if":   bson.M{"$eq": bson.A{"$is_reviewed", 1}},
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
		"like_count": bson.M{
			"$size": "$likes",
		},
		"dislike_count": bson.M{
			"$size": "$dislikes",
		},
	}}

	setPopularity := bson.M{"$set": bson.M{
		"popularity": bson.M{
			"$sum": bson.A{"$like_count", "$dislike_count"},
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
		match, set, setPopularity, lookup, unwind,
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
		"like_count": bson.M{
			"$size": "$likes",
		},
		"dislike_count": bson.M{
			"$size": "$dislikes",
		},
	}}

	setPopularity := bson.M{"$set": bson.M{
		"popularity": bson.M{
			"$sum": bson.A{"$like_count", "$dislike_count"},
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
		match, set, setPopularity, lookup, unwind,
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

func (reviewModel *ReviewModel) VoteReview(uid string, data requests.VoteReview, review Review) (Review, error) {
	objectReviewID, _ := primitive.ObjectIDFromHex(data.ID)

	var (
		isAlreadyLiked    bool
		isAlreadyDisliked bool
	)

	for _, like := range review.Likes {
		if like == uid {
			isAlreadyLiked = true
			review.Likes = removeElement(review.Likes, uid)
		}
	}

	for _, dislike := range review.Dislikes {
		if dislike == uid {
			isAlreadyDisliked = true
			review.Dislikes = removeElement(review.Dislikes, uid)
		}
	}

	if isAlreadyLiked || isAlreadyDisliked {
		if isAlreadyLiked && !data.IsLike {
			review.Dislikes = append(review.Dislikes, uid)
		} else if isAlreadyDisliked && data.IsLike {
			review.Likes = append(review.Likes, uid)
		}
	} else {
		if data.IsLike {
			review.Likes = append(review.Likes, uid)
		} else {
			review.Dislikes = append(review.Dislikes, uid)
		}
	}

	if _, err := reviewModel.ReviewCollection.UpdateOne(context.TODO(), bson.M{
		"_id":     objectReviewID,
		"user_id": uid,
	}, bson.M{"$set": review}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectReviewID,
			"data": data,
		}).Error("failed to update review: ", err)

		return Review{}, fmt.Errorf("Failed to like/dislike.")
	}

	return review, nil
}

func (reviewModel *ReviewModel) UpdateReview(uid string, data requests.UpdateReview, review Review) (Review, error) {
	objectReviewID, _ := primitive.ObjectIDFromHex(data.ID)

	review.Review = data.Review

	if data.Star != nil {
		review.Star = *data.Star
	}

	if _, err := reviewModel.ReviewCollection.UpdateOne(context.TODO(), bson.M{
		"_id":     objectReviewID,
		"user_id": uid,
	}, bson.M{"$set": review}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectReviewID,
			"data": data,
		}).Error("failed to update review: ", err)

		return Review{}, fmt.Errorf("Failed to update review.")
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
