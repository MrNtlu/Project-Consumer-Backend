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
	Order       int                `bson:"order" json:"order"`
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
				"order": 1,
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
		// Sort by order (0 to ++)
		{
			"$sort": bson.M{
				"order": 1,
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

// Achievement definitions with their IDs and conditions
var achievementMap = map[string]struct {
	ID        string
	Condition func(*AchievementModel, string) bool
}{
	// First activity
	"first_steps": {
		ID: "684e727dd9ef9c69eb3e0aa7",
		Condition: func(am *AchievementModel, uid string) bool {
			return true // Always unlock on first activity
		},
	},
	// Review achievements
	"reviewer": {
		ID: "684e727dd9ef9c69eb3e0aa8",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserReviewCount(uid) >= 1
		},
	},
	"critic": {
		ID: "684e727dd9ef9c69eb3e0aa9",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserReviewCount(uid) >= 10
		},
	},
	// Movie achievements
	"silver_screen_scout": {
		ID: "684e727dd9ef9c69eb3e0aaa",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "movie") >= 10
		},
	},
	"reel_explorer": {
		ID: "684e727dd9ef9c69eb3e0aab",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "movie") >= 25
		},
	},
	"cinema_sage": {
		ID: "684e727dd9ef9c69eb3e0aac",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "movie") >= 50
		},
	},
	// TV achievements
	"pilot_hunter": {
		ID: "684e727dd9ef9c69eb3e0aad",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "tv") >= 10
		},
	},
	"seasoned_binger": {
		ID: "684e727dd9ef9c69eb3e0aae",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "tv") >= 25
		},
	},
	"episodic_legend": {
		ID: "684e727dd9ef9c69eb3e0aaf",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "tv") >= 50
		},
	},
	// Game achievements
	"rookie_adventurer": {
		ID: "684e727dd9ef9c69eb3e0ab0",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "game") >= 10
		},
	},
	"pixel_challenger": {
		ID: "684e727dd9ef9c69eb3e0ab1",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "game") >= 25
		},
	},
	"digital_conqueror": {
		ID: "684e727dd9ef9c69eb3e0ab2",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "game") >= 50
		},
	},
	// Anime achievements
	"newcomer_to_nippon": {
		ID: "684e727dd9ef9c69eb3e0ab3",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "anime") >= 10
		},
	},
	"story_arc_wanderer": {
		ID: "684e727dd9ef9c69eb3e0ab4",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "anime") >= 25
		},
	},
	"legend_of_the_otaku": {
		ID: "684e727dd9ef9c69eb3e0ab5",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserFinishedCount(uid, "anime") >= 50
		},
	},
	// Special achievements
	"devoted_soul": {
		ID: "684e727dd9ef9c69eb3e0ab6",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserMaxTimesFinished(uid) >= 10
		},
	},
	// Watch Later achievements
	"future_watcher": {
		ID: "684e727dd9ef9c69eb3e0ab7",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserWatchLaterCount(uid) >= 10
		},
	},
	"content_collector": {
		ID: "684e727dd9ef9c69eb3e0ab8",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserWatchLaterCount(uid) >= 25
		},
	},
	"archiver_of_anticipation": {
		ID: "684e727dd9ef9c69eb3e0ab9",
		Condition: func(am *AchievementModel, uid string) bool {
			return am.getUserWatchLaterCount(uid) >= 50
		},
	},
}

// CheckAndUnlockAchievements checks and unlocks achievements for a user based on their activity
// This runs asynchronously to avoid impacting response times
func (achievementModel *AchievementModel) CheckAndUnlockAchievements(uid string, activityType string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"user_id":       uid,
					"activity_type": activityType,
					"panic":         r,
				}).Error("panic in achievement checker")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Check relevant achievements based on activity type
		var achievementsToCheck []string

		switch activityType {
		case "first_activity":
			achievementsToCheck = []string{"first_steps"}
		case "review":
			achievementsToCheck = []string{"reviewer", "critic"}
		case "movie_finished":
			achievementsToCheck = []string{"silver_screen_scout", "reel_explorer", "cinema_sage", "devoted_soul"}
		case "tv_finished":
			achievementsToCheck = []string{"pilot_hunter", "seasoned_binger", "episodic_legend", "devoted_soul"}
		case "game_finished":
			achievementsToCheck = []string{"rookie_adventurer", "pixel_challenger", "digital_conqueror", "devoted_soul"}
		case "anime_finished":
			achievementsToCheck = []string{"newcomer_to_nippon", "story_arc_wanderer", "legend_of_the_otaku", "devoted_soul"}
		case "watch_later":
			achievementsToCheck = []string{"future_watcher", "content_collector", "archiver_of_anticipation"}
		default:
			return
		}

		for _, achievementKey := range achievementsToCheck {
			achievement, exists := achievementMap[achievementKey]
			if !exists {
				continue
			}

			// Check if user already has this achievement
			count, err := achievementModel.UserAchievementCollection.CountDocuments(ctx, bson.M{
				"user_id":        uid,
				"achievement_id": achievement.ID,
			})
			if err != nil || count > 0 {
				continue // Skip if error or already unlocked
			}

			// Check if condition is met
			if achievement.Condition(achievementModel, uid) {
				err := achievementModel.UnlockAchievement(uid, achievement.ID)
				if err == nil {
					logrus.WithFields(logrus.Fields{
						"user_id":        uid,
						"achievement_id": achievement.ID,
						"achievement":    achievementKey,
					}).Info("achievement unlocked")
				}
			}
		}
	}()
}

// Helper functions to check user statistics
func (achievementModel *AchievementModel) getUserReviewCount(uid string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := achievementModel.AchievementCollection.Database().Collection("reviews").CountDocuments(ctx, bson.M{
		"user_id": uid,
	})
	if err != nil {
		return 0
	}
	return count
}

func (achievementModel *AchievementModel) getUserFinishedCount(uid, contentType string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var collectionName string
	switch contentType {
	case "movie":
		collectionName = "movie-watch-lists"
	case "tv":
		collectionName = "tvseries-watch-lists"
	case "game":
		collectionName = "game-lists"
	case "anime":
		collectionName = "anime-lists"
	default:
		return 0
	}

	count, err := achievementModel.AchievementCollection.Database().Collection(collectionName).CountDocuments(ctx, bson.M{
		"user_id": uid,
		"status":  "finished",
	})
	if err != nil {
		return 0
	}
	return count
}

func (achievementModel *AchievementModel) getUserMaxTimesFinished(uid string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collections := []string{"movie-watch-lists", "tvseries-watch-lists", "game-lists", "anime-lists"}
	maxTimes := int64(0)

	for _, collectionName := range collections {
		pipeline := []bson.M{
			{"$match": bson.M{"user_id": uid}},
			{"$group": bson.M{
				"_id":       nil,
				"max_times": bson.M{"$max": "$times_finished"},
			}},
		}

		cursor, err := achievementModel.AchievementCollection.Database().Collection(collectionName).Aggregate(ctx, pipeline)
		if err != nil {
			continue
		}

		var result []bson.M
		if err := cursor.All(ctx, &result); err == nil && len(result) > 0 {
			if times, ok := result[0]["max_times"].(int32); ok && int64(times) > maxTimes {
				maxTimes = int64(times)
			}
		}
		cursor.Close(ctx)
	}

	return maxTimes
}

func (achievementModel *AchievementModel) getUserWatchLaterCount(uid string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := achievementModel.AchievementCollection.Database().Collection("consume-laters").CountDocuments(ctx, bson.M{
		"user_id": uid,
	})
	if err != nil {
		return 0
	}
	return count
}
