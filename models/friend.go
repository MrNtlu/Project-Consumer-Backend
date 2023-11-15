package models

import (
	"app/db"
	"app/responses"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendModel struct {
	RequestCollection      *mongo.Collection
	ShareAccountCollection *mongo.Collection
}

func NewFriendModel(mongoDB *db.MongoDB) *FriendModel {
	return &FriendModel{
		RequestCollection:      mongoDB.Database.Collection("friend-requests"),
		ShareAccountCollection: mongoDB.Database.Collection("share-accounts"),
	}
}

type FriendRequest struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	SenderID         string             `bson:"sender_id" json:"sender_id"`
	SenderUsername   string             `bson:"sender_username" json:"sender_username"`
	ReceiverID       string             `bson:"receiver_id" json:"receiver_id"`
	ReceiverUsername string             `bson:"receiver_username" json:"receiver_username"`
	IsIgnored        bool               `bson:"is_ignored" json:"is_ignored"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

func createFriendRequestObject(senderId, senderUsername, receiverId, receiverUsername string) *FriendRequest {
	return &FriendRequest{
		SenderID:         senderId,
		SenderUsername:   senderUsername,
		ReceiverID:       receiverId,
		ReceiverUsername: receiverUsername,
		IsIgnored:        false,
		CreatedAt:        time.Now().UTC(),
	}
}

func (friendModel *FriendModel) CreateFriendRequest(senderId, senderUsername, receiverId, receiverUsername string) error {
	friendRequest := createFriendRequestObject(senderId, senderUsername, receiverId, receiverUsername)

	_, err := friendModel.RequestCollection.InsertOne(context.TODO(), friendRequest)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"friend_request": friendRequest,
		}).Error("failed to create friend request: ", err)

		return fmt.Errorf("Failed to send friend request.")
	}

	return nil
}

func (friendModel *FriendModel) GetFriendRequests(uid string) ([]responses.FriendRequest, error) {
	match := bson.M{"$match": bson.M{
		"receiver_id": uid,
	}}

	set := bson.M{"$set": bson.M{
		"receiver_obj_id": bson.M{
			"$toObjectId": "$receiver_id",
		},
		"sender_obj_id": bson.M{
			"$toObjectId": "$sender_id",
		},
	}}

	senderLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "sender_obj_id",
		"foreignField": "_id",
		"as":           "sender",
	}}

	senderUnwind := bson.M{"$unwind": bson.M{
		"path":                       "$sender",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	receiverLookup := bson.M{"$lookup": bson.M{
		"from":         "users",
		"localField":   "receiver_obj_id",
		"foreignField": "_id",
		"as":           "receiver",
	}}

	receiverUnwind := bson.M{"$unwind": bson.M{
		"path":                       "$receiver",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	sort := bson.M{"$sort": bson.M{
		"is_ignored": -1,
		"created_at": -1,
	}}

	cursor, err := friendModel.RequestCollection.Aggregate(context.TODO(), bson.A{
		match, set, senderLookup, senderUnwind, receiverLookup, receiverUnwind, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate friend requests: ", err)

		return nil, fmt.Errorf("Failed to aggregate friend requests.")
	}

	var friendRequests []responses.FriendRequest
	if err = cursor.All(context.TODO(), &friendRequests); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode friend request: ", err)

		return nil, fmt.Errorf("Failed to decode friend requests.")
	}

	return friendRequests, nil
}

func (friendModel *FriendModel) IsFriendRequestSent(senderId, receiverId string) (bool, error) {
	count, err := friendModel.RequestCollection.CountDocuments(context.TODO(), bson.M{
		"sender_id":   senderId,
		"receiver_id": receiverId,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"sender_id":   senderId,
			"receiver_id": receiverId,
		}).Error("failed to check friend request status: ", err)

		return false, fmt.Errorf("Failed to check friend request.")
	}

	return count > 0, nil
}

func (friendModel *FriendModel) FriendRequestCount(receiverId string) (int64, error) {
	count, err := friendModel.RequestCollection.CountDocuments(context.TODO(), bson.M{
		"receiver_id": receiverId,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"receiver_id": receiverId,
		}).Error("failed to check friend request status: ", err)

		return 0, fmt.Errorf("Failed to check friend request.")
	}

	return count, nil
}
