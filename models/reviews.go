package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"
	"time"
	"unicode"

	goaway "github.com/TwiN/go-away"
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
	IsSpoiler            bool               `bson:"is_spoiler" json:"is_spoiler"`
	Likes                []string           `bson:"likes" json:"likes"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

func containsNonASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return true
		}
	}
	return false
}

func createReviewObject(
	userID, contentID, contentType, review string, contentExternalID *string,
	contentExternalIntID *int64, star int8, isSpoiler bool,
) *Review {
	return &Review{
		UserID:               userID,
		ContentID:            contentID,
		ContentExternalID:    contentExternalID,
		ContentExternalIntID: contentExternalIntID,
		ContentType:          contentType,
		Star:                 star,
		Review:               review,
		IsSpoiler:            isSpoiler,
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
		ContentType:          review.ContentType,
		Star:                 review.Star,
		IsAuthor:             true,
		IsLiked:              false,
		Review:               review.Review,
		IsSpoiler:            review.IsSpoiler,
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
		ContentType:          review.ContentType,
		Star:                 review.Star,
		Review:               review.Review,
		IsSpoiler:            review.IsSpoiler,
		Likes:                review.Likes,
		CreatedAt:            review.CreatedAt,
		UpdatedAt:            review.UpdatedAt,
	}
}

func (reviewModel *ReviewModel) CreateReview(uid string, data requests.CreateReview) (responses.Review, error) {
	var reviewStr string
	if containsNonASCII(data.Review) {
		reviewStr = data.Review
	} else {
		reviewStr = goaway.Censor(data.Review)
	}

	review := createReviewObject(
		uid,
		data.ContentID,
		data.ContentType,
		reviewStr,
		data.ContentExternalID,
		data.ContentExternalIntID,
		data.Star,
		*data.IsSpoiler,
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

func (reviewModel *ReviewModel) GetReviewDetails(uid *string, reviewID string) (responses.ReviewDetails, error) {
	var (
		uidAggregation  primitive.M
		likeAggregation primitive.M
	)

	objectID, _ := primitive.ObjectIDFromHex(reviewID)

	if uid != nil {
		uidAggregation = bson.M{
			"$eq": bson.A{
				"$user_id", uid,
			},
		}

		likeAggregation = bson.M{
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
		}
	} else {
		uidAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}

		likeAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}
	}

	match := bson.M{"$match": bson.M{
		"_id": objectID,
	}}

	set := bson.M{"$set": bson.M{
		"is_author": uidAggregation,
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"obj_content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"obj_likes": bson.M{
			"$map": bson.M{
				"input": bson.M{
					"$slice": bson.A{"$likes", 10},
				},
				"as": "likes",
				"in": bson.M{"$toObjectId": "$$likes"},
			},
		},
		"popularity": bson.M{
			"$size": "$likes",
		},
		"is_liked": likeAggregation,
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

	facet := bson.M{"$facet": bson.M{
		"movies": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "movie"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$tmdb_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       1,
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
		"tv": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "tv"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$tmdb_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       1,
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
		"anime": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "anime"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_int_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$mal_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       1,
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
		"games": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "game"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_int_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$rawg_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       "$title",
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"reviews": bson.M{
			"$concatArrays": bson.A{"$movies", "$tv", "$anime", "$games"},
		},
	}}

	unwindReviews := bson.M{"$unwind": bson.M{
		"path":                       "$reviews",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$reviews",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	unwindObjLikes := bson.M{"$unwind": bson.M{
		"path":                       "$obj_likes",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	likesLookup := bson.M{"$lookup": bson.M{
		"from": "users",
		"let": bson.M{
			"user_id": "$obj_likes",
		},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$_id", "$$user_id"},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id":        1,
					"username":   1,
					"email":      1,
					"image":      1,
					"is_premium": 1,
				},
			},
		},
		"as": "likes",
	}}

	unwindLikes := bson.M{"$unwind": bson.M{
		"path":                       "$likes",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"user_id": bson.M{
			"$first": "$user_id",
		},
		"content_id": bson.M{
			"$first": "$content_id",
		},
		"content_external_id": bson.M{
			"$first": "$content_external_id",
		},
		"content_external_int_id": bson.M{
			"$first": "$content_external_int_id",
		},
		"review": bson.M{
			"$first": "$review",
		},
		"author": bson.M{
			"$first": "$author",
		},
		"popularity": bson.M{
			"$first": "$popularity",
		},
		"content": bson.M{
			"$first": "$content",
		},
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"is_author": bson.M{
			"$first": "$is_author",
		},
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"is_spoiler": bson.M{
			"$first": "$is_spoiler",
		},
		"star": bson.M{
			"$first": "$star",
		},
		"likes": bson.M{
			"$push": "$likes",
		},
		"created_at": bson.M{
			"$first": "$created_at",
		},
		"updated_at": bson.M{
			"$first": "$updated_at",
		},
	}}

	cursor, err := reviewModel.ReviewCollection.Aggregate(context.TODO(), bson.A{
		match, set, lookup, unwind, facet, project, unwindReviews, replaceRoot,
		unwindContent, unwindObjLikes, likesLookup, unwindLikes, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"review_id": reviewID,
			"uid":       uid,
		}).Error("failed to aggregate review details: ", err)

		return responses.ReviewDetails{}, fmt.Errorf("Failed to aggregate review details.")
	}

	var reviewDetails []responses.ReviewDetails
	if err = cursor.All(context.TODO(), &reviewDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"review_id": reviewID,
			"uid":       uid,
		}).Error("failed to decode review details: ", err)

		return responses.ReviewDetails{}, fmt.Errorf("Failed to decode review details.")
	}

	if len(reviewDetails) > 0 {
		return reviewDetails[0], nil
	}

	return responses.ReviewDetails{}, nil
}

func (reviewModel *ReviewModel) GetReviewSummaryForDetails(contentID, uid string, contentExternalID *string, contentExternalIntID *int64) (responses.ReviewSummary, error) {
	var (
		cID  string
		ceID int64
	)

	if contentExternalID != nil {
		cID = *contentExternalID
	} else {
		cID = "-999"
	}

	if contentExternalIntID != nil {
		ceID = *contentExternalIntID
	} else {
		ceID = -999
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
		sortType             string
		sortOrder            int8
		contentExternalID    string
		contentExternalIntID int64
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

	if data.ContentExternalID != nil {
		contentExternalID = *data.ContentExternalID
	} else {
		contentExternalID = "-999"
	}

	if data.ContentExternalIntID != nil {
		contentExternalIntID = *data.ContentExternalIntID
	} else {
		contentExternalIntID = -999
	}

	logrus.WithFields(logrus.Fields{
		"request": data,
	}).Info("Request")

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"content_id": data.ContentID,
			},
			bson.M{
				"content_external_id": contentExternalID,
			},
			bson.M{
				"content_external_int_id": contentExternalIntID,
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
		sortType             string
		sortOrder            int8
		contentExternalID    string
		contentExternalIntID int64
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

	if data.ContentExternalID != nil {
		contentExternalID = *data.ContentExternalID
	} else {
		contentExternalID = "-999"
	}

	if data.ContentExternalIntID != nil {
		contentExternalIntID = *data.ContentExternalIntID
	} else {
		contentExternalIntID = -999
	}

	match := bson.M{"$match": bson.M{
		"$or": bson.A{
			bson.M{
				"content_id": data.ContentID,
			},
			bson.M{
				"content_external_id": contentExternalID,
			},
			bson.M{
				"content_external_int_id": contentExternalIntID,
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

func (reviewModel *ReviewModel) GetReviewsByUserID(uid *string, data requests.SortReviewByUserID) ([]responses.ReviewWithContent, p.PaginationData, error) {
	var (
		sortType        string
		sortOrder       int8
		uidAggregation  primitive.M
		likeAggregation primitive.M
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

	if uid != nil {
		uidAggregation = bson.M{
			"$eq": bson.A{
				"$user_id", uid,
			},
		}

		likeAggregation = bson.M{
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
		}
	} else {
		uidAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}

		likeAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}
	}

	match := bson.M{"$match": bson.M{
		"user_id": data.UserID,
	}}

	set := bson.M{"$set": bson.M{
		"is_author": uidAggregation,
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"obj_content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"popularity": bson.M{
			"$size": "$likes",
		},
		"is_liked": likeAggregation,
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

	facet := bson.M{"$facet": bson.M{
		"movies": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "movie"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$tmdb_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       1,
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
		"tv": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "tv"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$tmdb_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       1,
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
		"anime": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "anime"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_int_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$mal_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       1,
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
		"games": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "game"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"content_id":  "$content_id",
						"external_id": "$content_external_int_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{"$eq": bson.A{"$_id", "$$content_id"}},
										bson.M{"$eq": bson.A{"$rawg_id", "$$external_id"}},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"title_en":       "$title",
								"title_original": 1,
								"image_url":      1,
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"reviews": bson.M{
			"$concatArrays": bson.A{"$movies", "$tv", "$anime", "$games"},
		},
	}}

	unwindReviews := bson.M{"$unwind": bson.M{
		"path":                       "$reviews",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$reviews",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	paginatedData, err := p.New(reviewModel.ReviewCollection).Context(context.TODO()).Limit(reviewPagination).
		Page(data.Page).Sort("is_author", -1).Sort(sortType, sortOrder).Aggregate(
		match, set, lookup, unwind, facet, project, unwindReviews, replaceRoot, unwindContent,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":     uid,
			"request": data,
		}).Error("failed to aggregate reviews by user id", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get reviews by user id.")
	}

	var reviews []responses.ReviewWithContent
	for _, raw := range paginatedData.Data {
		var review *responses.ReviewWithContent
		if marshalErr := bson.Unmarshal(raw, &review); marshalErr == nil {
			reviews = append(reviews, *review)
		}
	}

	return reviews, paginatedData.Pagination, nil
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

func (reviewModel *ReviewModel) GetBaseReviewResponseByUserIDAndContentID(contentID, uid string) (responses.Review, error) {
	result := reviewModel.ReviewCollection.FindOne(context.TODO(), bson.M{
		"content_id": contentID,
		"user_id":    uid,
	})

	var review Review
	if err := result.Decode(&review); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id":    uid,
			"content_id": contentID,
		}).Error("failed to find review by content id: ", err)

		return responses.Review{}, fmt.Errorf("Failed to find review by id.")
	}

	return convertReviewModelToResponse(review), nil
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
		if containsNonASCII(*data.Review) {
			review.Review = *data.Review
		} else {
			review.Review = goaway.Censor(*data.Review)
		}
	}

	if data.Star != nil {
		review.Star = *data.Star
	}

	if data.IsSpoiler != nil {
		review.IsSpoiler = *data.IsSpoiler
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
