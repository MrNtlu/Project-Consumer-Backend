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
