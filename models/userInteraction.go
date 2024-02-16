package models

import (
	"app/db"
	"app/requests"
	"app/responses"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//lint:file-ignore ST1005 Ignore all

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
	ContentType          string             `bson:"content_type" json:"content_type"` // anime, movie, tv or game
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

func (userInteractionModel *UserInteractionModel) CreateConsumeLater(uid string, data requests.CreateConsumeLater) (ConsumeLaterList, error) {
	consumeLater := createConsumeLaterObject(
		uid,
		data.ContentID,
		data.ContentType,
		data.ContentExternalID,
		data.SelfNote,
		data.ContentExternalIntID,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = userInteractionModel.ConsumeLaterCollection.InsertOne(context.TODO(), consumeLater); err != nil {
		logrus.WithFields(logrus.Fields{
			"consume_later": consumeLater,
		}).Error("failed to create new consume later: ", err)

		return ConsumeLaterList{}, fmt.Errorf("Failed to create consume later.")
	}

	consumeLater.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *consumeLater, nil
}

func (userInteractionModel *UserInteractionModel) GetConsumeLaterCount(uid string) int64 {
	consumeLaterCount, err := userInteractionModel.ConsumeLaterCollection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count consume later list: ", err)

		return -1
	}

	return consumeLaterCount
}

func (userInteractionModel *UserInteractionModel) GetBaseConsumeLater(uid, id string) (ConsumeLaterList, error) {
	objectID, _ := primitive.ObjectIDFromHex(id)

	result := userInteractionModel.ConsumeLaterCollection.FindOne(context.TODO(), bson.M{
		"user_id": uid,
		"_id":     objectID,
	})

	var consumeLater ConsumeLaterList
	if err := result.Decode(&consumeLater); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to find consume later by user id: ", err)

		return ConsumeLaterList{}, fmt.Errorf("Failed to find consume later by user id.")
	}

	return consumeLater, nil
}

func (userInteractionModel *UserInteractionModel) GetConsumeLater(uid string, data requests.SortFilterConsumeLater) ([]responses.ConsumeLater, error) {
	var (
		match     bson.M
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "new":
		sortType = "created_at"
		sortOrder = -1
	case "old":
		sortType = "created_at"
		sortOrder = 1
	case "alphabetical":
		sortType = "content.title_original"
		sortOrder = 1
	case "unalphabetical":
		sortType = "content.title_original"
		sortOrder = -1
	}

	sort := bson.M{"$sort": bson.M{
		sortType: sortOrder,
	}}

	if data.ContentType != nil {
		match = bson.M{"$match": bson.M{
			"user_id":      uid,
			"content_type": data.ContentType,
		}}
	} else {
		match = bson.M{"$match": bson.M{
			"user_id": uid,
		}}
	}

	set := bson.M{"$set": bson.M{
		"content_id": bson.M{
			"$toObjectId": "$content_id",
		},
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
								"description":    1,
								"score":          "$tmdb_vote",
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
								"description":    1,
								"score":          "$tmdb_vote",
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
								"description":    1,
								"score":          "$mal_score",
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
								"description":    1,
								"score":          "$rawg_rating",
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"consume-laters": bson.M{
			"$concatArrays": bson.A{"$movies", "$tv", "$anime", "$games"},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$consume-laters",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$consume-laters",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	cursor, err := userInteractionModel.ConsumeLaterCollection.Aggregate(context.TODO(), bson.A{
		match, set, facet, project, unwind, replaceRoot, unwindContent, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find consume later by user id: ", err)

		return nil, fmt.Errorf("Failed to find consume later by user id.")
	}

	var consumeLaterList []responses.ConsumeLater
	if err := cursor.All(context.TODO(), &consumeLaterList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode consume later by user id: ", err)

		return nil, fmt.Errorf("Failed to decode consume later by user id.")
	}

	return consumeLaterList, nil
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
