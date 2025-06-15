package models

import (
	"app/db"
	"app/responses"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementModel struct {
	AchievementCollection     *mongo.Collection
	UserAchievementCollection *mongo.Collection
}

func NewAchievementModel(mongoDB *db.MongoDB) *AchievementModel {
	return &AchievementModel{
		AchievementCollection:     mongoDB.Database.Collection("achievements"),
		UserAchievementCollection: mongoDB.Database.Collection("user-achievements"),
	}
}

type Achievement struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title       string             `bson:"title" json:"title"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type UserAchievement struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	AchievementID string             `bson:"achievement_id" json:"achievement_id"`
	UnlockedAt    time.Time          `bson:"unlocked_at" json:"unlocked_at"`
}

func (achievementModel *AchievementModel) GetAllAchievements() ([]responses.Achievement, error) {
	pipeline := []bson.M{
		{
			"$sort": bson.M{
				"created_at": 1,
			},
		},
	}

	cursor, err := achievementModel.AchievementCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		logrus.Error("failed to get all achievements: ", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var achievements []responses.Achievement
	if err = cursor.All(context.TODO(), &achievements); err != nil {
		logrus.Error("failed to decode achievements: ", err)
		return nil, err
	}

	return achievements, nil
}

func (achievementModel *AchievementModel) GetUserAchievements(uid string) ([]responses.UserAchievementResponse, error) {
	// MongoDB aggregation pipeline to get all achievements with user unlock status
	pipeline := []bson.M{
		// Lookup user achievements for this specific user
		{
			"$lookup": bson.M{
				"from": "user-achievements",
				"let":  bson.M{"achievement_id": "$_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{"$eq": []interface{}{"$user_id", uid}},
									{"$eq": []interface{}{"$achievement_id", bson.M{"$toString": "$$achievement_id"}}},
								},
							},
						},
					},
				},
				"as": "user_achievement",
			},
		},
		// Add unlocked field and unlocked_at if exists
		{
			"$addFields": bson.M{
				"unlocked": bson.M{
					"$gt": []interface{}{bson.M{"$size": "$user_achievement"}, 0},
				},
				"unlocked_at": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$gt": []interface{}{bson.M{"$size": "$user_achievement"}, 0}},
						"then": bson.M{"$arrayElemAt": []interface{}{"$user_achievement.unlocked_at", 0}},
						"else": nil,
					},
				},
			},
		},
		// Remove the user_achievement array from output
		{
			"$project": bson.M{
				"user_achievement": 0,
			},
		},
		// Sort by creation date
		{
			"$sort": bson.M{
				"created_at": 1,
			},
		},
	}

	cursor, err := achievementModel.AchievementCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to aggregate user achievements: ", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var achievements []responses.UserAchievementResponse
	if err = cursor.All(context.TODO(), &achievements); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to decode user achievements: ", err)
		return nil, err
	}

	return achievements, nil
}

func (achievementModel *AchievementModel) UnlockAchievement(uid, achievementID string) error {
	// Check if user already has this achievement
	count, err := achievementModel.UserAchievementCollection.CountDocuments(
		context.TODO(),
		bson.M{
			"user_id":        uid,
			"achievement_id": achievementID,
		},
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id":        uid,
			"achievement_id": achievementID,
		}).Error("failed to check existing user achievement: ", err)
		return err
	}

	// If user already has this achievement, don't add it again
	if count > 0 {
		return nil
	}

	// Create new user achievement
	userAchievement := UserAchievement{
		UserID:        uid,
		AchievementID: achievementID,
		UnlockedAt:    time.Now().UTC(),
	}

	_, err = achievementModel.UserAchievementCollection.InsertOne(context.TODO(), userAchievement)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id":        uid,
			"achievement_id": achievementID,
		}).Error("failed to unlock achievement: ", err)
		return err
	}

	return nil
}

func (achievementModel *AchievementModel) GetUserAchievementCount(uid string) (int64, error) {
	count, err := achievementModel.UserAchievementCollection.CountDocuments(
		context.TODO(),
		bson.M{"user_id": uid},
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to get user achievement count: ", err)
		return 0, err
	}

	return count, nil
}

func (achievementModel *AchievementModel) DeleteUserAchievementsByUserID(uid string) {
	if _, err := achievementModel.UserAchievementCollection.DeleteMany(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete user achievements by user id: ", err)
	}
}
