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
	Bookmarks   []string            `bson:"bookmarks" json:"bookmarks"`
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
*/

func createCustomListObject(userID, name string, description *string, isPrivate bool, content []requests.CustomListContent) *CustomList {
	return &CustomList{
		UserID:      userID,
		Name:        name,
		Description: description,
		Likes:       []string{},
		Bookmarks:   []string{},
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

func (customListModel *CustomListModel) BookmarkCustomList(uid string, data requests.ID, customList responses.CustomList) (responses.CustomList, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(data.ID)

	var isAlreadyBookmarked = false

	for _, bookmark := range customList.Bookmarks {
		if bookmark == uid {
			isAlreadyBookmarked = true
			customList.IsBookmarked = false
			customList.BookmarkCount = customList.BookmarkCount - 1
			customList.Bookmarks = removeElement(customList.Bookmarks, uid)
		}
	}

	if !isAlreadyBookmarked {
		customList.Bookmarks = append(customList.Bookmarks, uid)
		customList.BookmarkCount = customList.BookmarkCount + 1
		customList.IsBookmarked = true
	}

	updatedReview := convertCustomListResponseToModel(customList)
	if _, err := customListModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id": objectCustomListID,
	}, bson.M{"$set": updatedReview}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectCustomListID,
			"data": data,
		}).Error("failed to bookmark custom list: ", err)

		return responses.CustomList{}, fmt.Errorf("Failed to bookmark.")
	}

	return customList, nil
}

func (customListModel *CustomListModel) LikeCustomList(uid string, data requests.ID, customList responses.CustomList) (responses.CustomList, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(data.ID)

	var isAlreadyLiked = false

	for _, like := range customList.Likes {
		if like == uid {
			isAlreadyLiked = true
			customList.IsLiked = false
			customList.Popularity = customList.Popularity - 1
			customList.Likes = removeElement(customList.Likes, uid)
		}
	}

	if !isAlreadyLiked {
		customList.Likes = append(customList.Likes, uid)
		customList.Popularity = customList.Popularity + 1
		customList.IsLiked = true
	}

	updatedReview := convertCustomListResponseToModel(customList)
	if _, err := customListModel.Collection.UpdateOne(context.TODO(), bson.M{
		"_id": objectCustomListID,
	}, bson.M{"$set": updatedReview}); err != nil {
		logrus.WithFields(logrus.Fields{
			"_id":  objectCustomListID,
			"data": data,
		}).Error("failed to update like custom list: ", err)

		return responses.CustomList{}, fmt.Errorf("Failed to like.")
	}

	return customList, nil
}

func (customListModel *CustomListModel) UpdateCustomList(uid string, data requests.UpdateCustomList, customList CustomList) (CustomList, error) {
	objectCustomListID, _ := primitive.ObjectIDFromHex(data.ID)

	customList.Name = data.Name
	customList.Description = data.Description
	customList.IsPrivate = *data.IsPrivate
	customList.Content = convertRequestToModel(data.Content)

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

func (customListModel *CustomListModel) DeleteAllCustomListsByUserID(uid string) error {
	if _, err := customListModel.Collection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete all custom list: ", err)

		return fmt.Errorf("Failed to delete all custom list.")
	}

	return nil
}

func (customListModel *CustomListModel) GetCustomListsByUserID(uid *string, data requests.SortCustomListUID, hidePrivate, isSocial bool) ([]responses.CustomList, error) {
	var (
		sortType            string
		sortOrder           int8
		likeAggregation     primitive.M
		bookmarkAggregation primitive.M
		match               bson.M
	)

	if uid != nil && data.UserID == "" {
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
		bookmarkAggregation = bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$bookmarks",
					},
				},
				"then": true,
				"else": false,
			},
		}
		if isSocial {
			match = bson.M{"$match": bson.M{
				"is_private": false,
			}}
		} else if hidePrivate {
			match = bson.M{"$match": bson.M{
				"user_id":    *uid,
				"is_private": false,
			}}
		} else {
			match = bson.M{"$match": bson.M{
				"user_id": *uid,
			}}
		}
	} else if uid != nil && data.UserID != "" {
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
		bookmarkAggregation = bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$bookmarks",
					},
				},
				"then": true,
				"else": false,
			},
		}
		if hidePrivate {
			match = bson.M{"$match": bson.M{
				"user_id":    data.UserID,
				"is_private": false,
			}}
		} else {
			match = bson.M{"$match": bson.M{
				"user_id": data.UserID,
			}}
		}
	} else {
		likeAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}
		bookmarkAggregation = bson.M{
			"$eq": bson.A{
				-1, 1,
			},
		}
		if data.UserID != "" {
			match = bson.M{"$match": bson.M{
				"user_id":    data.UserID,
				"is_private": false,
			}}
		} else {
			match = bson.M{"$match": bson.M{
				"is_private": false,
			}}
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

	set := bson.M{"$set": bson.M{
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"obj_content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"popularity": bson.M{
			"$sum": bson.A{
				bson.M{"$size": "$likes"},
				bson.M{"$multiply": bson.A{bson.M{"$size": "$bookmarks"}, 3}},
			},
		},
		"bookmark_count": bson.M{
			"$size": "$bookmarks",
		},
		"is_liked":      likeAggregation,
		"is_bookmarked": bookmarkAggregation,
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
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"bookmarks": bson.M{
			"$first": "$bookmarks",
		},
		"is_bookmarked": bson.M{
			"$first": "$is_bookmarked",
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
		"bookmark_count": bson.M{
			"$first": "$bookmark_count",
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
		match, set, unwindContent, facet, project, unwind, replaceRoot, unwindContentAgain, group, lookup, unwindAuthor, sort,
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

func (customListModel *CustomListModel) GetCustomLists(uid *string, data requests.SortCustomList) ([]responses.CustomList, error) {
	var (
		sortType            string
		sortOrder           int8
		likeAggregation     primitive.M
		bookmarkAggregation primitive.M
	)

	match := bson.M{"$match": bson.M{
		"is_private": false,
	}}

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
		bookmarkAggregation = bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$bookmarks",
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
		bookmarkAggregation = bson.M{
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

	set := bson.M{"$set": bson.M{
		"obj_user_id": bson.M{
			"$toObjectId": "$user_id",
		},
		"obj_content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"popularity": bson.M{
			"$sum": bson.A{
				bson.M{"$size": "$likes"},
				bson.M{"$multiply": bson.A{bson.M{"$size": "$bookmarks"}, 3}},
			},
		},
		"bookmark_count": bson.M{
			"$size": "$bookmarks",
		},
		"is_liked":      likeAggregation,
		"is_bookmarked": bookmarkAggregation,
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
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"bookmarks": bson.M{
			"$first": "$bookmarks",
		},
		"is_bookmarked": bson.M{
			"$first": "$is_bookmarked",
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
		"bookmark_count": bson.M{
			"$first": "$bookmark_count",
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
		match, set, unwindContent, facet, project, unwind, replaceRoot, unwindContentAgain, group, lookup, unwindAuthor, sort,
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

func (customListModel *CustomListModel) GetLikedCustomLists(uid string, data requests.SortLikeBookmarkCustomList) ([]responses.CustomList, error) {
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
	case "alphabetical":
		sortType = "name"
		sortOrder = 1
	case "unalphabetical":
		sortType = "name"
		sortOrder = -1
	}

	var bookmarkAggregation = bson.M{
		"$cond": bson.M{
			"if": bson.M{
				"$in": bson.A{
					uid,
					"$bookmarks",
				},
			},
			"then": true,
			"else": false,
		},
	}

	match := bson.M{"$match": bson.M{
		"likes": bson.M{
			"$in": bson.A{uid},
		},
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
		"bookmark_count": bson.M{
			"$size": "$bookmarks",
		},
		"is_bookmarked": bookmarkAggregation,
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
		"is_liked": bson.M{
			"$first": true,
		},
		"bookmarks": bson.M{
			"$first": "$bookmarks",
		},
		"is_bookmarked": bson.M{
			"$first": "$is_bookmarked",
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
		"bookmark_count": bson.M{
			"$first": "$bookmark_count",
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
		match, set, unwindContent, facet, project, unwind, replaceRoot,
		unwindContentAgain, group, lookup, unwindAuthor, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find liked custom list: ", err)

		return nil, fmt.Errorf("Failed to find liked custom list.")
	}

	var customListDetails []responses.CustomList
	if err := cursor.All(context.TODO(), &customListDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode liked custom list: ", err)

		return nil, fmt.Errorf("Failed to decode liked custom list.")
	}

	return customListDetails, nil
}

func (customListModel *CustomListModel) GetBookmarkedCustomLists(uid string, data requests.SortLikeBookmarkCustomList) ([]responses.CustomList, error) {
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
	case "alphabetical":
		sortType = "name"
		sortOrder = 1
	case "unalphabetical":
		sortType = "name"
		sortOrder = -1
	}

	var likeAggregation = bson.M{
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

	match := bson.M{"$match": bson.M{
		"likes": bson.M{
			"$in": bson.A{uid},
		},
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
		"bookmark_count": bson.M{
			"$size": "$bookmarks",
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
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"bookmarks": bson.M{
			"$first": "$bookmarks",
		},
		"is_bookmarked": bson.M{
			"$first": true,
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
		"bookmark_count": bson.M{
			"$first": "$bookmark_count",
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
		match, set, unwindContent, facet, project, unwind, replaceRoot,
		unwindContentAgain, group, lookup, unwindAuthor, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find liked custom list: ", err)

		return nil, fmt.Errorf("Failed to find liked custom list.")
	}

	var customListDetails []responses.CustomList
	if err := cursor.All(context.TODO(), &customListDetails); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode liked custom list: ", err)

		return nil, fmt.Errorf("Failed to decode liked custom list.")
	}

	return customListDetails, nil
}

func (customListModel *CustomListModel) GetCustomListDetails(uid *string, customListID string) (responses.CustomList, error) {
	var (
		likeAggregation     primitive.M
		bookmarkAggregation primitive.M
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
		bookmarkAggregation = bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$bookmarks",
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
		bookmarkAggregation = bson.M{
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
		"bookmark_count": bson.M{
			"$size": "$bookmarks",
		},
		"is_liked":      likeAggregation,
		"is_bookmarked": bookmarkAggregation,
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
		"is_liked": bson.M{
			"$first": "$is_liked",
		},
		"bookmarks": bson.M{
			"$first": "$bookmarks",
		},
		"is_bookmarked": bson.M{
			"$first": "$is_bookmarked",
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
		"bookmark_count": bson.M{
			"$first": "$bookmark_count",
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

func (customListModel *CustomListModel) GetBaseCustomListResponse(uid *string, customListID string) (responses.CustomList, error) {
	var (
		likeAggregation     primitive.M
		bookmarkAggregation primitive.M
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
		bookmarkAggregation = bson.M{
			"$cond": bson.M{
				"if": bson.M{
					"$in": bson.A{
						uid,
						"$bookmarks",
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
		bookmarkAggregation = bson.M{
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
		"bookmark_count": bson.M{
			"$size": "$bookmarks",
		},
		"is_liked":      likeAggregation,
		"is_bookmarked": bookmarkAggregation,
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
		match, set, lookup, unwindAuthor,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":            uid,
			"custom_list_id": customListID,
		}).Error("failed to find custom list response: ", err)

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

func convertCustomListResponseToModel(customList responses.CustomList) CustomList {
	return CustomList{
		ID:          customList.ID,
		UserID:      customList.UserID,
		Likes:       customList.Likes,
		Bookmarks:   customList.Bookmarks,
		CreatedAt:   customList.CreatedAt,
		Name:        customList.Name,
		Description: customList.Description,
		IsPrivate:   customList.IsPrivate,
		Content:     convertCustomListContentResponseToModel(customList.Content),
	}
}

func convertCustomListContentResponseToModel(customList []responses.CustomListContent) []CustomListContent {
	var customListContent []CustomListContent
	for i := 0; i < len(customList); i++ {
		customListContent = append(customListContent, CustomListContent{
			Order:                customList[i].Order,
			ContentID:            customList[i].ContentID,
			ContentExternalID:    customList[i].ContentExternalID,
			ContentExternalIntID: customList[i].ContentExternalIntID,
			ContentType:          customList[i].ContentType,
		})
	}

	return customListContent
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
