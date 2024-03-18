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

type LogsModel struct {
	LogsCollection *mongo.Collection
}

func NewLogsModel(mongoDB *db.MongoDB) *LogsModel {
	return &LogsModel{
		LogsCollection: mongoDB.Database.Collection("logs"),
	}
}

const (
	UserListLogType     = "userlist"
	ConsumeLaterLogType = "later"
	ReviewLogType       = "review"

	AddLogAction    = "add"
	UpdateLogAction = "update"
	DeleteLogAction = "delete"

	FinishedActionDetails = "finished"
	ActiveActionDetails   = "active"
	DroppedActionDetails  = "dropped"
)

type Log struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID           string             `bson:"user_id" json:"user_id"`
	LogType          string             `bson:"log_type" json:"log_type"`                     //User List, Consume Later
	LogAction        string             `bson:"log_action" json:"log_action"`                 //add, update, delete
	LogActionDetails string             `bson:"log_action_details" json:"log_action_details"` //1 Episode added etc.
	ContentTitle     string             `bson:"content_title" json:"content_title"`
	ContentImage     string             `bson:"content_image" json:"content_image"`
	ContentType      string             `bson:"content_type" json:"content_type"`
	ContentID        string             `bson:"content_id" json:"content_id"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

func createLogObject(userID, logType, logAction, logActionDetails, contentTitle, contentImage, contentType, contentID string) *Log {
	return &Log{
		UserID:           userID,
		LogType:          logType,
		LogAction:        logAction,
		LogActionDetails: logActionDetails,
		ContentTitle:     contentTitle,
		ContentImage:     contentImage,
		ContentType:      contentType,
		ContentID:        contentID,
		CreatedAt:        time.Now().UTC(),
	}
}

func (logsModel *LogsModel) CreateLog(uid string, data requests.CreateLog) {
	log := createLogObject(
		uid,
		data.LogType,
		data.LogAction,
		data.LogActionDetails,
		data.ContentTitle,
		data.ContentImage,
		data.ContentType,
		data.ContentID,
	)

	if _, err := logsModel.LogsCollection.InsertOne(context.TODO(), log); err != nil {
		logrus.WithFields(logrus.Fields{
			"log": log,
		}).Error("failed to create new log: ", err)
	}
}

func (logsModel *LogsModel) MostLikedGenresByLogs(uid string) ([]responses.MostLikedGenres, error) {
	match := bson.M{"$match": bson.M{
		"user_id":            uid,
		"log_type":           "userlist",
		"log_action_details": "finished",
	}}

	duplicateClearGroup := bson.M{"$group": bson.M{
		"_id": "$content_id",
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"content_id": bson.M{
			"$first": "$content_id",
		},
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"content_type": 1,
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
						"content_id": "$content_id",
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
								"genres": 1,
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
						"content_id": "$content_id",
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
								"genres": 1,
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
									"$eq": bson.A{"$_id", "$$content_id"},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"genres": "$genres.name",
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
						"content_id": "$content_id",
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
								"genres": 1,
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	projectLogs := bson.M{"$project": bson.M{
		"logs": bson.M{
			"$concatArrays": bson.A{
				"$movies",
				"$tv",
				"$anime",
				"$games",
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$logs",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$logs",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	unwindGenres := bson.M{"$unwind": bson.M{
		"path":                       "$content.genres",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"genres": "$content.genres",
			"type":   "$content_type",
		},
		"genre": bson.M{
			"$first": "$content.genres",
		},
		"type": bson.M{
			"$first": "$content_type",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"type":  1,
		"count": -1,
	}}

	groupType := bson.M{"$group": bson.M{
		"_id": "$type",
		"type": bson.M{
			"$first": "$type",
		},
		"genre": bson.M{
			"$first": "$genre",
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, duplicateClearGroup, project, facet, projectLogs, unwind, replaceRoot, unwindContent, unwindGenres, group, sort, groupType,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find most liked genres by logs: ", err)

		return nil, fmt.Errorf("Failed to find most liked genres by logs.")
	}

	var mostLikedGenres []responses.MostLikedGenres
	if err := cursor.All(context.TODO(), &mostLikedGenres); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode most liked genres by logs: ", err)

		return nil, fmt.Errorf("Failed to decode most liked genres by logs.")
	}

	return mostLikedGenres, nil
}

func (logsModel *LogsModel) MostLikedCountryByLogs(uid string) ([]responses.MostLikedCountry, error) {
	match := bson.M{"$match": bson.M{
		"user_id":            uid,
		"log_type":           "userlist",
		"log_action_details": "finished",
		"content_type": bson.M{
			"$ne": "game",
		},
	}}

	duplicateClearGroup := bson.M{"$group": bson.M{
		"_id": "$content_id",
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"content_id": bson.M{
			"$first": "$content_id",
		},
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"content_type": 1,
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
						"content_id": "$content_id",
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
								"country": "$production_companies.origin_country",
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
						"content_id": "$content_id",
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
								"country": "$production_companies.origin_country",
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
									"$eq": bson.A{"$_id", "$$content_id"},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"country": "$demographics.name",
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	projectLogs := bson.M{"$project": bson.M{
		"logs": bson.M{
			"$concatArrays": bson.A{
				"$movies",
				"$tv",
				"$anime",
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$logs",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$logs",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	unwindGenres := bson.M{"$unwind": bson.M{
		"path":                       "$content.country",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"country": "$content.country",
			"type":    "$content_type",
		},
		"country": bson.M{
			"$first": "$content.country",
		},
		"type": bson.M{
			"$first": "$content_type",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"type":  1,
		"count": -1,
	}}

	groupType := bson.M{"$group": bson.M{
		"_id": "$type",
		"type": bson.M{
			"$first": "$type",
		},
		"country": bson.M{
			"$first": "$country",
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, duplicateClearGroup, project, facet, projectLogs, unwind, replaceRoot, unwindContent, unwindGenres, group, sort, groupType,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find most liked genres by logs: ", err)

		return nil, fmt.Errorf("Failed to find most liked genres by logs.")
	}

	var mostLikedCountry []responses.MostLikedCountry
	if err := cursor.All(context.TODO(), &mostLikedCountry); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode most liked genres by logs: ", err)

		return nil, fmt.Errorf("Failed to decode most liked genres by logs.")
	}

	return mostLikedCountry, nil
}

func (logsModel *LogsModel) MostWatchedActors(uid string) ([]responses.MostWatchedActors, error) {
	match := bson.M{"$match": bson.M{
		"user_id":            uid,
		"log_type":           "userlist",
		"log_action_details": "finished",
		"content_type": bson.M{
			"$nin": bson.A{"game", "anime"},
		},
	}}

	duplicateClearGroup := bson.M{"$group": bson.M{
		"_id": "$content_id",
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"content_id": bson.M{
			"$first": "$content_id",
		},
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"content_type": 1,
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
						"content_id": "$content_id",
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
								"actors": bson.M{
									"$slice": bson.A{"$actors", 3},
								},
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
						"content_id": "$content_id",
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
								"actors": bson.M{
									"$slice": bson.A{"$actors", 3},
								},
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	projectLogs := bson.M{"$project": bson.M{
		"logs": bson.M{
			"$concatArrays": bson.A{
				"$movies",
				"$tv",
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$logs",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$logs",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	unwindGenres := bson.M{"$unwind": bson.M{
		"path":                       "$content.actors",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"name": "$content.actors.name",
			"type": "$content_type",
		},
		"id": bson.M{
			"$first": "$content.actors.tmdb_id",
		},
		"actor": bson.M{
			"$first": "$content.actors.name",
		},
		"image": bson.M{
			"$first": "$content.actors.image",
		},
		"type": bson.M{
			"$first": "$content_type",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"type":  1,
		"count": -1,
	}}

	groupType := bson.M{"$group": bson.M{
		"_id": "$type",
		"actors": bson.M{
			"$addToSet": bson.M{
				"id":    "$id",
				"name":  "$actor",
				"image": "$image",
				"count": "$count",
			},
		},
		"type": bson.M{
			"$first": "$type",
		},
	}}

	set := bson.M{"$set": bson.M{
		"actors": bson.M{
			"$slice": bson.A{
				bson.M{
					"$sortArray": bson.M{
						"input":  "$actors",
						"sortBy": bson.M{"count": -1},
					},
				},
				3,
			},
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, duplicateClearGroup, project, facet, projectLogs, unwind,
		replaceRoot, unwindContent, unwindGenres, group, sort, groupType, set,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find most watched actors by logs: ", err)

		return nil, fmt.Errorf("Failed to find most watched actors by logs.")
	}

	var mostWatchedActors []responses.MostWatchedActors
	if err := cursor.All(context.TODO(), &mostWatchedActors); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode most watched actors by logs: ", err)

		return nil, fmt.Errorf("Failed to decode most watched actors by logs.")
	}

	return mostWatchedActors, nil
}

func (logsModel *LogsModel) MostLikedStudios(uid string) ([]responses.MostLikedStudios, error) {
	match := bson.M{"$match": bson.M{
		"user_id":            uid,
		"log_type":           "userlist",
		"log_action_details": "finished",
		"content_type": bson.M{
			"$nin": bson.A{"tv", "movie"},
		},
	}}

	duplicateClearGroup := bson.M{"$group": bson.M{
		"_id": "$content_id",
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"content_id": bson.M{
			"$first": "$content_id",
		},
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"content_type": 1,
	}}

	facet := bson.M{"$facet": bson.M{
		"anime": bson.A{
			bson.M{
				"$match": bson.M{"content_type": "anime"},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"content_id": "$content_id",
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
								"studios": "$studios.name",
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
						"content_id": "$content_id",
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
								"studios": "$publishers",
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	projectLogs := bson.M{"$project": bson.M{
		"logs": bson.M{
			"$concatArrays": bson.A{
				"$games",
				"$anime",
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$logs",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$logs",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	unwindGenres := bson.M{"$unwind": bson.M{
		"path":                       "$content.studios",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"name": "$content.studios",
			"type": "$content_type",
		},
		"name": bson.M{
			"$first": "$content.studios",
		},
		"type": bson.M{
			"$first": "$content_type",
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"type":  1,
		"count": -1,
	}}

	groupType := bson.M{"$group": bson.M{
		"_id": "$type",
		"studios": bson.M{
			"$addToSet": bson.M{
				"name":  "$name",
				"count": "$count",
			},
		},
		"type": bson.M{
			"$first": "$type",
		},
	}}

	set := bson.M{"$set": bson.M{
		"studios": bson.M{
			"$slice": bson.A{
				bson.M{
					"$sortArray": bson.M{
						"input":  "$studios",
						"sortBy": bson.M{"count": -1},
					},
				},
				3,
			},
		},
	}}

	projectSchema := bson.M{"$project": bson.M{
		"type":    1,
		"studios": "$studios.name",
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, duplicateClearGroup, project, facet, projectLogs, unwind,
		replaceRoot, unwindContent, unwindGenres, group, sort, groupType,
		set, projectSchema,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to find most liked studios by logs: ", err)

		return nil, fmt.Errorf("Failed to find most liked studios by logs.")
	}

	var mostLikedStudios []responses.MostLikedStudios
	if err := cursor.All(context.TODO(), &mostLikedStudios); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode most liked studios by logs: ", err)

		return nil, fmt.Errorf("Failed to decode most liked studios by logs.")
	}

	return mostLikedStudios, nil
}

func (logsModel *LogsModel) FinishedLogStats(uid string, data requests.LogStatInterval) ([]responses.FinishedLogStats, error) {
	var intervalDate time.Time

	switch data.Interval {
	case "weekly":
		intervalDate = time.Now().UTC().AddDate(0, 0, -7)
	case "monthly":
		intervalDate = time.Now().UTC().AddDate(0, -1, 0)
	case "3months":
		intervalDate = time.Now().UTC().AddDate(0, -3, 0)
	}

	match := bson.M{"$match": bson.M{
		"user_id":            uid,
		"log_type":           "userlist",
		"log_action_details": "finished",
		"created_at": bson.M{
			"$gte": intervalDate,
		},
	}}

	duplicateClearGroup := bson.M{"$group": bson.M{
		"_id": "$content_id",
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"content_id": bson.M{
			"$first": "$content_id",
		},
	}}

	project := bson.M{"$project": bson.M{
		"content_id": bson.M{
			"$toObjectId": "$content_id",
		},
		"content_type": 1,
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
						"content_id": "$content_id",
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
								"length": 1,
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
						"content_id": "$content_id",
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
								"total_episodes": 1,
								"total_seasons":  1,
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
									"$eq": bson.A{"$_id", "$$content_id"},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"total_episodes": "$episodes",
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
						"content_id": "$content_id",
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
								"metacritic_score": 1,
							},
						},
					},
					"as": "content",
				},
			},
		},
	}}

	projectLogs := bson.M{"$project": bson.M{
		"logs": bson.M{
			"$concatArrays": bson.A{
				"$movies",
				"$tv",
				"$anime",
				"$games",
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$logs",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$logs",
	}}

	unwindContent := bson.M{"$unwind": bson.M{
		"path":                       "$content",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$content_type",
		"content_type": bson.M{
			"$first": "$content_type",
		},
		"length": bson.M{
			"$sum": "$content.length",
		},
		"total_episodes": bson.M{
			"$sum": "$content.total_episodes",
		},
		"total_seasons": bson.M{
			"$sum": "$content.total_seasons",
		},
		"count": bson.M{
			"$sum": 1,
		},
		"metacritic_score": bson.M{
			"$sum": "$content.metacritic_score",
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, duplicateClearGroup, project, facet, projectLogs, unwind, replaceRoot, unwindContent, group,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find user stats by logs: ", err)

		return nil, fmt.Errorf("Failed to find user stats by logs.")
	}

	var finishedLogStats []responses.FinishedLogStats
	if err := cursor.All(context.TODO(), &finishedLogStats); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode user stats by logs: ", err)

		return nil, fmt.Errorf("Failed to decode user stats by logs.")
	}

	return finishedLogStats, nil
}

func (logsModel *LogsModel) LogStatisticsChart(uid string, data requests.LogStatInterval) ([]responses.ChartLogs, error) {
	var intervalDate time.Time

	switch data.Interval {
	case "weekly":
		intervalDate = time.Now().UTC().AddDate(0, 0, -7)
	case "monthly":
		intervalDate = time.Now().UTC().AddDate(0, -1, 0)
	case "3months":
		intervalDate = time.Now().UTC().AddDate(0, -3, 0)
	}

	match := bson.M{"$match": bson.M{
		"user_id": uid,
		"created_at": bson.M{
			"$gte": intervalDate,
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": bson.M{
			"$dateToString": bson.M{
				"format": "%Y-%m-%d",
				"date":   "$created_at",
			},
		},
		"count": bson.M{
			"$sum": 1,
		},
	}}

	sort := bson.M{"$sort": bson.M{
		"_id": -1,
	}}

	set := bson.M{"$set": bson.M{
		"created_at": bson.M{
			"$toDate": "$_id",
		},
		"day_of_week": bson.M{
			"$dayOfWeek": bson.M{
				"$toDate": "$_id",
			},
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, group, sort, set,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to find user chart logs: ", err)

		return nil, fmt.Errorf("Failed to find user chart logs.")
	}

	var chartLogs []responses.ChartLogs
	if err := cursor.All(context.TODO(), &chartLogs); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode user chart logs: ", err)

		return nil, fmt.Errorf("Failed to decode user chart logs.")
	}

	return chartLogs, nil
}

func (logsModel *LogsModel) GetLogStreak(uid string) (int, int) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}

	set := bson.M{"$set": bson.M{
		"created_at": bson.M{
			"$dateToString": bson.M{
				"format": "%Y-%m-%d",
				"date":   "$created_at",
			},
		},
	}}

	setToDate := bson.M{"$set": bson.M{
		"created_at": bson.M{
			"$toDate": "$created_at",
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$user_id",
		"dates": bson.M{
			"$addToSet": "$created_at",
		},
	}}

	setSort := bson.M{"$set": bson.M{
		"dates": bson.M{
			"$sortArray": bson.M{
				"input":  "$dates",
				"sortBy": 1,
			},
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, set, setToDate, group, setSort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate logs: ", err)

		return 0, 0
	}

	var logs []responses.LogDates
	if err = cursor.All(context.TODO(), &logs); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode logs: ", err)

		return 0, 0
	}

	if len(logs) > 0 {
		return calculateStreak(logs[0])
	}

	return 0, 0
}

func (logsModel *LogsModel) GetLogsByDateRange(uid string, data requests.LogsByDateRange) ([]responses.LogsByRange, error) {
	dateString := "2006-01-02"

	fromDate, _ := time.Parse(dateString, data.From)
	toDate, _ := time.Parse(dateString, data.To)

	match := bson.M{"$match": bson.M{
		"user_id": uid,
		"created_at": bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		},
	}}

	set := bson.M{"$set": bson.M{
		"created_at_str": bson.M{
			"$dateToString": bson.M{
				"format": "%Y-%m-%d",
				"date":   "$created_at",
			},
		},
	}}

	group := bson.M{"$group": bson.M{
		"_id": "$created_at_str",
		"data": bson.M{
			"$push": "$$ROOT",
		},
		"count": bson.M{
			"$count": bson.M{},
		},
	}}

	sortArray := bson.M{"$set": bson.M{
		"date": "$_id",
		"data": bson.M{
			"$sortArray": bson.M{
				"input": "$data",
				"sortBy": bson.M{
					"created_at": -1,
				},
			},
		},
	}}

	cursor, err := logsModel.LogsCollection.Aggregate(context.TODO(), bson.A{
		match, set, group, sortArray,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to aggregate logs: ", err)

		return nil, fmt.Errorf("Failed to aggregate logs.")
	}

	var logs []responses.LogsByRange
	if err = cursor.All(context.TODO(), &logs); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode logs: ", err)

		return nil, fmt.Errorf("Failed to decode logs.")
	}

	return logs, nil
}

func (logsModel *LogsModel) DeleteLogsByUserID(uid string) {
	if _, err := logsModel.LogsCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete logs by user id: ", err)
	}
}

func calculateStreak(logs responses.LogDates) (int, int) {
	var (
		logDates      []time.Time
		maxStreak     int
		currentStreak int
	)

	logDates = logs.Dates
	maxStreak = 0
	currentStreak = 0

	for i := 1; i < len(logDates); i++ {
		prevDate := logDates[i-1]
		currentDate := logDates[i]

		currentTime := time.Now()
		today, _ := time.Parse("2006-01-02", currentTime.Format("2006-01-02"))

		if currentDate.Sub(prevDate).Hours() == 24 {
			currentStreak = currentStreak + 1
		}

		if maxStreak < currentStreak {
			maxStreak = currentStreak
		}

		if currentDate.Sub(prevDate).Hours() != 24 {
			currentStreak = 0
		}

		if i == len(logDates) && today.Sub(currentDate).Hours() > 47 {
			currentStreak = 0
		}

		if i == (len(logDates)-1) && today.Sub(currentDate).Hours() == 24 {
			currentStreak = currentStreak + 1
		}
	}

	return maxStreak, currentStreak
}
