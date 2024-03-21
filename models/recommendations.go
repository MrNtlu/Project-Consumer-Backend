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

//lint:file-ignore ST1005 Ignore all

type RecommendationModel struct {
	RecommendationCollection *mongo.Collection
}

const recommendationPagination = 25

func NewRecommendationModel(mongoDB *db.MongoDB) *RecommendationModel {
	return &RecommendationModel{
		RecommendationCollection: mongoDB.Database.Collection("recomendations"),
	}
}

type Recommendation struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID           string             `bson:"user_id" json:"user_id"`
	ContentID        string             `bson:"content_id" json:"content_id"`
	ContentType      string             `bson:"content_type" json:"content_type"` // anime, movie, tv or game
	RecommendationID string             `bson:"recommendation_id" json:"recommendation_id"`
	Reason           *string            `bson:"reason" json:"reason"`
	Likes            []string           `bson:"likes" json:"likes"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

func createRecommendationObject(
	userID, contentID, contentType,
	recommendationID string, reason *string,
) *Recommendation {
	return &Recommendation{
		UserID:           userID,
		ContentID:        contentID,
		ContentType:      contentType,
		RecommendationID: recommendationID,
		Reason:           reason,
		Likes:            []string{},
		CreatedAt:        time.Now().UTC(),
	}
}

/*
- Change Recommendation
- Like/Dislike Recommendation
- Liked Recommendations
- Recommendations independent from content
*/

func (recommendationModel *RecommendationModel) CreateRecommendation(
	uid string, data requests.CreateRecommendation,
) (Recommendation, error) {
	recommendation := createRecommendationObject(
		uid,
		data.ContentID,
		data.ContentType,
		data.RecommendationID,
		data.Reason,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = recommendationModel.RecommendationCollection.InsertOne(context.TODO(), recommendation); err != nil {
		logrus.WithFields(logrus.Fields{
			"recommendation": recommendation,
		}).Error("failed to create new recommendation: ", err)

		return Recommendation{}, fmt.Errorf("Failed to create recommendation.")
	}

	recommendation.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *recommendation, nil
}

func (recommendationModel *RecommendationModel) GetRecommendationsByContentID(
	uid string, isUIDNull bool, data requests.SortRecommendation,
) ([]responses.RecommendationWithContent, p.PaginationData, error) {
	var (
		sortType         string
		sortOrder        int8
		lookupCollection string
		set              bson.M
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

	switch data.ContentType {
	case "anime":
		lookupCollection = "animes"
	case "game":
		lookupCollection = "games"
	case "movie":
		lookupCollection = "movies"
	case "tv":
		lookupCollection = "tv-series"
	}

	match := bson.M{"$match": bson.M{
		"content_id": data.ContentID,
	}}

	if !isUIDNull {
		set = bson.M{"$set": bson.M{
			"is_author": bson.M{
				"$eq": bson.A{"$user_id", uid},
			},
			"obj_user_id": bson.M{
				"$toObjectId": "$user_id",
			},
			"obj_content_id": bson.M{
				"$toObjectId": "$content_id",
			},
			"obj_recommend_id": bson.M{
				"$toObjectId": "$recommendation_id",
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
	} else {
		set = bson.M{"$set": bson.M{
			"is_author": false,
			"obj_user_id": bson.M{
				"$toObjectId": "$user_id",
			},
			"obj_content_id": bson.M{
				"$toObjectId": "$content_id",
			},
			"obj_recommend_id": bson.M{
				"$toObjectId": "$recommendation_id",
			},
			"popularity": bson.M{
				"$size": "$likes",
			},
			"is_liked": false,
		}}
	}

	authorLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "obj_user_id",
		"foreignField": "_id",
		"as":           "author",
	}}

	unwindAuthor := bson.M{"$unwind": bson.M{
		"path":                       "$author",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	contentLookup := bson.M{"$lookup": bson.M{
		"from":         lookupCollection,
		"localField":   "obj_content_id",
		"foreignField": "_id",
		"as":           "content",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	recommendationLookup := bson.M{"$lookup": bson.M{
		"from":         lookupCollection,
		"localField":   "obj_recommend_id",
		"foreignField": "_id",
		"as":           "recommendation_content",
	}}

	unwindRecommendation := bson.M{"$unwind": bson.M{
		"path":                       "$recommendation_content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	paginatedData, err := p.New(recommendationModel.RecommendationCollection).Context(context.TODO()).
		Limit(recommendationPagination).Page(data.Page).Sort(sortType, sortOrder).Aggregate(
		match, set, authorLookup, unwindAuthor, contentLookup,
		unwindContent, recommendationLookup, unwindRecommendation,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"request": data,
		}).Error("failed to aggregate recommendations", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to get recommendations.")
	}

	var recommendations []responses.RecommendationWithContent
	for _, raw := range paginatedData.Data {
		var recommendation *responses.RecommendationWithContent
		if marshalErr := bson.Unmarshal(raw, &recommendation); marshalErr == nil {
			recommendations = append(recommendations, *recommendation)
		}
	}

	return recommendations, paginatedData.Pagination, nil
}

func (recommendationModel *RecommendationModel) GetRecommendationsByUserID(
	uid string, data requests.SortRecommendationByUserID,
) ([]responses.RecommendationWithContent, p.PaginationData, error) {
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
		"user_id": data.UserID,
	}}

	set := bson.M{"$set": bson.M{
		"is_author": bson.M{
			"$eq": bson.A{"$user_id", uid},
		},
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"obj_content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"obj_recommendation_id": bson.M{
			"$toObjectId": "$recommendation_id",
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

	facet := bson.M{"$facet": bson.M{
		"movies": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "movie"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"content_id": "$obj_content_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"content_id": "$obj_recommendation_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
					"as": "recommendation_content",
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
						"content_id": "$obj_content_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"content_id": "$obj_content_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
					"as": "recommendation_content",
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
						"content_id": "$obj_content_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"content_id": "$obj_recommendation_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
					"as": "recommendation_content",
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
						"content_id": "$obj_content_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"content_id": "$obj_recommendation_id",
					},
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$eq": bson.A{"$_id", "$$content_id"},
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
					"as": "recommendation_content",
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"recommendations": bson.M{
			"$concatArrays": bson.A{"$movies", "$tv", "$anime", "$games"},
		},
	}}

	unwindRecommendations := bson.M{"$unwind": bson.M{
		"path":                       "$recommendations",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$recommendations",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	unwindRecommendationContent := bson.M{"$unwind": bson.M{
		"path":                       "$recommendation_content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	paginatedData, err := p.New(recommendationModel.RecommendationCollection).Context(context.TODO()).Limit(recommendationPagination).
		Page(data.Page).Sort(sortType, sortOrder).Aggregate(
		match, set, lookup, unwind, facet, project,
		unwindRecommendations, replaceRoot,
		unwindContent, unwindRecommendationContent,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"data": data,
			"uid":  uid,
		}).Error("failed to aggregate recommendations: ", err)

		return nil, p.PaginationData{}, fmt.Errorf("Failed to aggregate recommendations.")
	}

	var recommendations []responses.RecommendationWithContent
	for _, raw := range paginatedData.Data {
		var recommendation *responses.RecommendationWithContent
		if marshalErr := bson.Unmarshal(raw, &recommendation); marshalErr == nil {
			recommendations = append(recommendations, *recommendation)
		}
	}

	return recommendations, paginatedData.Pagination, nil
}

func (recommendationModel *RecommendationModel) DeleteRecommendationByID(uid, recommendationId string) (bool, error) {
	objectReviewID, _ := primitive.ObjectIDFromHex(recommendationId)

	count, err := recommendationModel.RecommendationCollection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectReviewID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":               uid,
			"recommendation_id": recommendationId,
		}).Error("failed to delete recommendation: ", err)

		return false, fmt.Errorf("Failed to delete recommendation.")
	}

	return count.DeletedCount > 0, nil
}

func (recommendationModel *RecommendationModel) DeleteAllRecommendationByUserID(uid string) error {
	if _, err := recommendationModel.RecommendationCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete recommendations: ", err)

		return fmt.Errorf("Failed to delete recommendations.")
	}

	return nil
}
