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

type CustomListModel struct {
	Collection *mongo.Collection
}

func NewCustomListModel(mongoDB *db.MongoDB) *CustomListModel {
	return &CustomListModel{
		Collection: mongoDB.Database.Collection("custom-lists"),
	}
}

type CustomList struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	UserID      string              `bson:"user_id" json:"user_id"`
	Name        string              `bson:"name" json:"name"`
	Description *string             `bson:"description" json:"description"`
	Likes       []string            `bson:"likes" json:"likes"`
	IsPrivate   bool                `bson:"is_private" json:"is_private"`
	Content     []CustomListContent `bson:"content" json:"content"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
}

type CustomListContent struct {
	Order                int     `bson:"order" json:"order"`
	ContentID            string  `bson:"content_id" json:"content_id"`
	ContentExternalID    *string `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64  `bson:"content_external_int_id" json:"content_external_int_id"`
	ContentType          string  `bson:"content_type" json:"content_type"` // anime, movie, tv or game
}

const CustomListFreeLimit = 5
const CustomListPremiumLimit = 20

/* TODO Endpoints
* - Reorder custom list
	- https://github.com/MrNtlu/Asset-Manager/blob/master/models/favouriteInvesting.go
* - Like Custom List
*/

func createCustomListObject(userID, name string, description *string, isPrivate bool, content []requests.CustomListContent) *CustomList {
	return &CustomList{
		UserID:      userID,
		Name:        name,
		Description: description,
		Likes:       []string{},
		IsPrivate:   isPrivate,
		Content:     convertRequestToModel(content),
		CreatedAt:   time.Now().UTC(),
	}
}

func (customListModel *CustomListModel) CreateCustomList(uid string, data requests.CreateCustomList) (CustomList, error) {
	customList := createCustomListObject(
		uid,
		data.Name,
		data.Description,
		*data.IsPrivate,
		data.Content,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = customListModel.Collection.InsertOne(context.TODO(), customList); err != nil {
		logrus.WithFields(logrus.Fields{
			"custom_list": customList,
		}).Error("failed to create new custom list: ", err)

		return CustomList{}, fmt.Errorf("Failed to create custom list.")
	}

	customList.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *customList, nil
}

func (customListModel *CustomListModel) UpdateCustomList(uid string, data requests.UpdateCustomList, customList CustomList) (CustomList, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(data.ID)

	customList.Name = data.Name
	customList.Description = data.Description
	customList.IsPrivate = *data.IsPrivate

	if _, err := customListModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id":     objectCustomListID,
		"user_id": uid,
	}, bson.M{"$set": customList}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectCustomListID,
			"data": data,
		}).Error("failed to update custom list: ", err)

		return CustomList{}, fmt.Errorf("Failed to update custom list.")
	}

	return customList, nil
}

func (customListModel *CustomListModel) UpdateAddContentToCustomList(uid string, data requests.AddToCustomList, customList CustomList) (CustomList, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(data.ID)

	customList.Content = append(customList.Content, CustomListContent{
		Order:                len(customList.Content) + 1,
		ContentID:            data.Content.ContentID,
		ContentExternalID:    data.Content.ContentExternalID,
		ContentExternalIntID: data.Content.ContentExternalIntID,
		ContentType:          data.Content.ContentType,
	})

	if _, err := customListModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id":     objectCustomListID,
		"user_id": uid,
	}, bson.M{"$set": customList}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectCustomListID,
			"data": data,
		}).Error("failed to add content to custom list: ", err)

		return CustomList{}, fmt.Errorf("Failed to add content to custom list.")
	}

	return customList, nil
}

func (customListModel *CustomListModel) DeleteBulkContentFromCustomListByID(uid string, data requests.BulkDeleteCustomList) (bool, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(data.ID)

	count, err := customListModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id":     objectCustomListID,
		"user_id": uid,
	}, bson.M{"$pull": bson.M{
		"content": bson.M{
			"content_id": bson.M{
				"$in": data.Content,
			},
		},
	}})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"custom_list_id": objectCustomListID,
		}).Error("failed to bulk delete contents from custom list: ", err)

		return false, fmt.Errorf("Failed to bulk delete contents from custom list.")
	}

	return count.ModifiedCount > 0, nil
}

func (customListModel *CustomListModel) DeleteCustomListByID(uid, customID string) (bool, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(customID)

	count, err := customListModel.Collection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectCustomListID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"custom_list_id": objectCustomListID,
		}).Error("failed to delete custom list: ", err)

		return false, fmt.Errorf("Failed to delete custom list.")
	}

	return count.DeletedCount > 0, nil
}

func (customListModel *CustomListModel) GetCustomListsByUserID(uid *string, data requests.SortCustomList) ([]responses.CustomList, error) {
	var (
		sortType        string
		sortOrder       int8
		likeAggregation primitive.M
		matchId         string
	)

	if uid != nil {
		matchId = *uid
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
		matchId = data.UserID
		likeAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}
	}

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
	case "alphabetical":
		sortType = "name"
		sortOrder = 1
	case "unalphabetical":
		sortType = "name"
		sortOrder = -1
	}

	match := bson.M{"$match": bson.M{
		"user_id": matchId,
	}}

	set := bson.M{"$set": bson.M{
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
		"content": bson.M{
			"$map": bson.M{
				"input": "$content",
				"as":    "content",
				"in": bson.M{
					"order": "$$content.order",
					"content_obj_id": bson.M{
						"$toObjectId": "$$content.content_id",
					},
					"content_id":              "$$content.content_id",
					"content_type":            "$$content.content_type",
					"content_external_id":     "$$content.content_external_id",
					"content_external_int_id": "$$content.content_external_int_id",
				},
			},
		},
	}}

	setLimit := bson.M{"$set": bson.M{
		"content": bson.M{
			"$slice": bson.A{"$content", 5},
		},
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	facet := bson.M{"$facet": bson.M{
		"movies": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "movie"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"content_id":          "$content.content_id",
						"external_id":         "$content.content_external_id",
						"content_type":        "$content.content_type",
						"order":               "$content.order",
						"content_external_id": "$content.content_external_id",
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
								"content_external_id": "$$content_external_id",
								"content_type":        "$$content_type",
								"content_id":          "$$content_id",
								"order":               "$$order",
								"title_en":            1,
								"title_original":      1,
								"image_url":           1,
								"description":         1,
								"score":               "$tmdb_vote",
							},
						},
					},
					"as": "content",
				},
			},
		},
		"tv": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "tv"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"content_id":          "$content.content_id",
						"content_type":        "$content.content_type",
						"external_id":         "$content.content_external_id",
						"order":               "$content.order",
						"content_external_id": "$content.content_external_id",
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
								"content_id":          "$$content_id",
								"content_external_id": "$$content_external_id",
								"content_type":        "$$content_type",
								"order":               "$$order",
								"title_en":            1,
								"title_original":      1,
								"image_url":           1,
								"description":         1,
								"score":               "$tmdb_vote",
							},
						},
					},
					"as": "content",
				},
			},
		},
		"anime": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "anime"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"content_id":              "$content.content_id",
						"content_type":            "$content.content_type",
						"external_id":             "$content.content_external_int_id",
						"order":                   "$content.order",
						"content_external_int_id": "$content.content_external_int_id",
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
								"content_id":              "$$content_id",
								"order":                   "$$order",
								"content_external_int_id": "$$content_external_int_id",
								"content_type":            "$$content_type",
								"title_en":                1,
								"title_original":          1,
								"image_url":               1,
								"description":             1,
								"score":                   "$mal_score",
							},
						},
					},
					"as": "content",
				},
			},
		},
		"games": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "game"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"content_id":              "$content.content_id",
						"content_type":            "$content.content_type",
						"external_id":             "$content.content_external_int_id",
						"order":                   "$content.order",
						"content_external_int_id": "$content.content_external_int_id",
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
								"content_id":              "$$content_id",
								"order":                   "$$order",
								"content_external_int_id": "$$content_external_int_id",
								"content_type":            "$$content_type",
								"title_en":                "$title",
								"title_original":          1,
								"image_url":               1,
								"description":             1,
								"score":                   "$rawg_rating",
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"custom_list_contents": bson.M{
			"$concatArrays": bson.A{"$movies", "$tv", "$anime", "$games"},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$custom_list_contents",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$custom_list_contents",
	}}

	unwindContentAgain := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"user_id": bson.M{
			"$first": "$user_id",
		},
		"name": bson.M{
			"$first": "$name",
		},
		"description": bson.M{
			"$first": "$description",
		},
		"likes": bson.M{
			"$first": "$likes",
		},
		"is_private": bson.M{
			"$first": "$is_private",
		},
		"created_at": bson.M{
			"$first": "$created_at",
		},
		"popularity": bson.M{
			"$first": "$popularity",
		},
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"obj_user_id": bson.M{
			"$first": "$obj_user_id",
		},
		"content": bson.M{
			"$push": "$content",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
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

	sort := bson.M{"$sort": bson.M{
		sortType: sortOrder,
	}}

	cursor, err := customListModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, setLimit, unwindContent, facet, project, unwind, replaceRoot, unwindContentAgain, group, lookup, unwindAuthor, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find custom list by user id: ", err)

		return nil, fmt.Errorf("Failed to find custom list by user id.")
	}

	var customLists []responses.CustomList
	if err := cursor.All(context.TODO(), &customLists); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode custom list by user id: ", err)

		return nil, fmt.Errorf("Failed to decode custom list by user id.")
	}

	return customLists, nil
}

func (customListModel *CustomListModel) GetCustomListDetails(uid *string, customListID string) (responses.CustomList, error) {
	var (
		likeAggregation primitive.M
	)

	objectID, _ := primitive.ObjectIDFromHex(customListID)

	if uid != nil {
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
		"content": bson.M{
			"$map": bson.M{
				"input": "$content",
				"as":    "content",
				"in": bson.M{
					"order": "$$content.order",
					"content_obj_id": bson.M{
						"$toObjectId": "$$content.content_id",
					},
					"content_type":            "$$content.content_type",
					"content_external_id":     "$$content.content_external_id",
					"content_external_int_id": "$$content.content_external_int_id",
				},
			},
		},
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	facet := bson.M{"$facet": bson.M{
		"movies": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "movie"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"content_id":          "$content.content_id",
						"external_id":         "$content.content_external_id",
						"content_type":        "$content.content_type",
						"order":               "$content.order",
						"content_external_id": "$content.content_external_id",
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
								"content_external_id": "$$content_external_id",
								"content_type":        "$$content_type",
								"order":               "$$order",
								"title_en":            1,
								"title_original":      1,
								"image_url":           1,
								"description":         1,
								"score":               "$tmdb_vote",
							},
						},
					},
					"as": "content",
				},
			},
		},
		"tv": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "tv"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"content_id":          "$content.content_id",
						"content_type":        "$content.content_type",
						"external_id":         "$content.content_external_id",
						"order":               "$content.order",
						"content_external_id": "$content.content_external_id",
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
								"content_external_id": "$$content_external_id",
								"content_type":        "$$content_type",
								"order":               "$$order",
								"title_en":            1,
								"title_original":      1,
								"image_url":           1,
								"description":         1,
								"score":               "$tmdb_vote",
							},
						},
					},
					"as": "content",
				},
			},
		},
		"anime": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "anime"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"content_id":              "$content.content_id",
						"content_type":            "$content.content_type",
						"external_id":             "$content.content_external_int_id",
						"order":                   "$content.order",
						"content_external_int_id": "$content.content_external_int_id",
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
								"order":                   "$$order",
								"content_external_int_id": "$$content_external_int_id",
								"content_type":            "$$content_type",
								"title_en":                1,
								"title_original":          1,
								"image_url":               1,
								"description":             1,
								"score":                   "$mal_score",
							},
						},
					},
					"as": "content",
				},
			},
		},
		"games": bson.A{
			bson.M{
				"$match": bson.M{"content.content_type": "game"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"content_id":              "$content.content_id",
						"content_type":            "$content.content_type",
						"external_id":             "$content.content_external_int_id",
						"order":                   "$content.order",
						"content_external_int_id": "$content.content_external_int_id",
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
								"order":                   "$$order",
								"content_external_int_id": "$$content_external_int_id",
								"content_type":            "$$content_type",
								"title_en":                "$title",
								"title_original":          1,
								"image_url":               1,
								"description":             1,
								"score":                   "$rawg_rating",
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	project := bson.M{"$project": bson.M{
		"custom_list_contents": bson.M{
			"$concatArrays": bson.A{"$movies", "$tv", "$anime", "$games"},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$custom_list_contents",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$custom_list_contents",
	}}

	unwindContentAgain := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$_id",
		"user_id": bson.M{
			"$first": "$user_id",
		},
		"name": bson.M{
			"$first": "$name",
		},
		"description": bson.M{
			"$first": "$description",
		},
		"likes": bson.M{
			"$first": "$likes",
		},
		"is_private": bson.M{
			"$first": "$is_private",
		},
		"created_at": bson.M{
			"$first": "$created_at",
		},
		"popularity": bson.M{
			"$first": "$popularity",
		},
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"obj_user_id": bson.M{
			"$first": "$obj_user_id",
		},
		"content": bson.M{
			"$push": "$content",
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
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

	cursor, err := customListModel.Collection.Aggregate(context.TODO(), bson.A{
		match, set, unwindContent, facet, project, unwind, replaceRoot, unwindContentAgain, group, lookup, unwindAuthor,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"custom_list_id": customListID,
		}).Error("failed to find custom list details: ", err)

		return responses.CustomList{}, fmt.Errorf("Failed to find custom list details.")
	}

	var customListDetails []responses.CustomList
	if err := cursor.All(context.TODO(), &customListDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"custom_list_id": customListID,
		}).Error("failed to decode custom list details: ", err)

		return responses.CustomList{}, fmt.Errorf("Failed to decode custom list details.")
	}

	if len(customListDetails) > 0 {
		return customListDetails[0], nil
	}

	return responses.CustomList{}, nil
}

func (customListModel *CustomListModel) GetCustomListCount(uid string) int64 {
	count, err := customListModel.Collection.CountDocuments(context.TODO(), bson.M{"user_id": uid})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to count user custom list: ", err)

		return CustomListPremiumLimit
	}

	return count
}

func (customListModel *CustomListModel) GetBaseCustomList(uid, customListID string) (CustomList, error) {
	objectID, _ := primitive.ObjectIDFromHex(customListID)

	result := customListModel.Collection.FindOne(context.TODO(), bson.M{
		"_id":     objectID,
		"user_id": uid,
	})

	var customList CustomList
	if err := result.Decode(&customList); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
			"id":      customListID,
		}).Error("failed to find custom list by id: ", err)

		return CustomList{}, fmt.Errorf("Failed to find custom list by id.")
	}

	return customList, nil
}

func convertRequestToModel(content []requests.CustomListContent) []CustomListContent {
	var modelList []CustomListContent

	for i := 0; i < len(content); i++ {
		modelList = append(modelList, CustomListContent{
			Order:                content[i].Order,
			ContentID:            content[i].ContentID,
			ContentExternalID:    content[i].ContentExternalID,
			ContentExternalIntID: content[i].ContentExternalIntID,
			ContentType:          content[i].ContentType,
		})
	}

	return modelList
}
