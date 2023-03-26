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

type UserListModel struct {
	UserListCollection          *mongo.Collection
	AnimeListCollection         *mongo.Collection
	GameListCollection          *mongo.Collection
	MovieWatchListCollection    *mongo.Collection
	TVSeriesWatchListCollection *mongo.Collection
}

func NewUserListModel(mongoDB *db.MongoDB) *UserListModel {
	return &UserListModel{
		UserListCollection:          mongoDB.Database.Collection("user-lists"),
		AnimeListCollection:         mongoDB.Database.Collection("anime-lists"),
		GameListCollection:          mongoDB.Database.Collection("game-lists"),
		MovieWatchListCollection:    mongoDB.Database.Collection("movie-watch-lists"),
		TVSeriesWatchListCollection: mongoDB.Database.Collection("tvseries-watch-lists"),
	}
}

/**
* !Features
* Combination of all lists.
* Calculate total amount of episodes/games/movies watched etc.
* Mean/Median etc. scores
* Create slug for sharing
*
* !Premium Features
* ? - Not decided yet
**/
type UserList struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID   string             `bson:"user_id" json:"user_id"`
	Slug     string             `bson:"slug" json:"slug"`
	IsPublic bool               `bson:"is_public" json:"is_public"`
}

type AnimeList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	Status          string             `bson:"status" json:"status"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	Score           *float32           `bson:"score" json:"score"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"-"`
}

type GameList struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	GameID            string             `bson:"game_id" json:"game_id"`
	Status            string             `bson:"status" json:"status"`
	Score             *float32           `bson:"score" json:"score"`
	AchievementStatus *float32           `bson:"achievement_status" json:"achievement_status"`
	TimesFinished     int                `bson:"times_finished" json:"times_finished"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"-"`
}

type MovieWatchList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	MovieID       string             `bson:"movie_id" json:"movie_id"`
	Status        string             `bson:"status" json:"status"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"-"`
}

// TODO
// This has to be different
// Unlike anime, seasons are not separated.
type TVSeriesWatchList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	TvID            string             `bson:"tv_id" json:"tv_id"`
	Status          string             `bson:"status" json:"status"`
	Score           *float32           `bson:"score" json:"score"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	WatchedSeasons  int                `bson:"watched_seasons" json:"watched_seasons"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"-"`
}

func createUserListObject(userID, slug string) *UserList {
	return &UserList{
		UserID:   userID,
		Slug:     slug,
		IsPublic: true,
	}
}

func createAnimeListObject(userID, animeID, status string, watchedEpisodes int, score *float32) *AnimeList {
	return &AnimeList{
		UserID:          userID,
		AnimeID:         animeID,
		Status:          status,
		WatchedEpisodes: watchedEpisodes,
		Score:           score,
		TimesFinished:   handleTimesFinished(status),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

func createGameListObject(userID, gameID, status string, score, achievementStatus *float32) *GameList {
	return &GameList{
		UserID:            userID,
		GameID:            gameID,
		Status:            status,
		Score:             score,
		AchievementStatus: achievementStatus,
		TimesFinished:     handleTimesFinished(status),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
}

func createMovieWatchListObject(userID, movieID, status string, score *float32) *MovieWatchList {
	return &MovieWatchList{
		UserID:        userID,
		MovieID:       movieID,
		Status:        status,
		Score:         score,
		TimesFinished: handleTimesFinished(status),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

func createTVSeriesWatchListObject(userID, tvID, status string, watchedEpisodes, watchedSeasons int, score *float32) *TVSeriesWatchList {
	return &TVSeriesWatchList{
		UserID:          userID,
		TvID:            tvID,
		Status:          status,
		Score:           score,
		WatchedEpisodes: watchedEpisodes,
		WatchedSeasons:  watchedSeasons,
		TimesFinished:   handleTimesFinished(status),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

func handleTimesFinished(status string) int {
	if status == "finished" {
		return 1
	}
	return 0
}

/* TODO
* [x] Create user list when they register
* [x] Add xx list
* [] Get user list and others
* [] Update xx list by ID
* [x] Delete xx list by ID
* [x] Delete all by user list
 */

//! Create
func (userListModel *UserListModel) CreateUserList(uid, slug string) {
	userListObject := createUserListObject(uid, slug)

	if _, err := userListModel.UserListCollection.InsertOne(context.TODO(), userListObject); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"slug": slug,
		}).Error("failed to create new user list: ", err)
	}
}

func (userListModel *UserListModel) CreateAnimeList(uid string, data requests.CreateAnimeList, anime responses.Anime) error {
	animeList := createAnimeListObject(
		uid, data.AnimeID, data.Status,
		*data.WatchedEpisodes, data.Score,
	)

	if anime.Episodes != nil {
		if *data.WatchedEpisodes > int(*anime.Episodes) {
			animeList.WatchedEpisodes = int(*anime.Episodes)
		} else if data.Status == "finished" && *data.WatchedEpisodes < int(*anime.Episodes) {
			animeList.WatchedEpisodes = int(*anime.Episodes)
		}
	}

	if _, err := userListModel.AnimeListCollection.InsertOne(context.TODO(), animeList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new anime list: ", err)

		return fmt.Errorf("Failed to create new anime list.")
	}

	return nil
}

func (userListModel *UserListModel) CreateGameList(uid string, data requests.CreateGameList) error {
	gameList := createGameListObject(
		uid, data.GameID, data.Status,
		data.Score, data.AchievementStatus,
	)

	if _, err := userListModel.GameListCollection.InsertOne(context.TODO(), gameList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new game list: ", err)

		return fmt.Errorf("Failed to create new game list.")
	}

	return nil
}

func (userListModel *UserListModel) CreateMovieWatchList(uid string, data requests.CreateMovieWatchList) error {
	movieWatchList := createMovieWatchListObject(uid, data.MovieID, data.Status, data.Score)

	if _, err := userListModel.MovieWatchListCollection.InsertOne(context.TODO(), movieWatchList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new movie watch list: ", err)

		return fmt.Errorf("Failed to create new movie watch list.")
	}

	return nil
}

func (userListModel *UserListModel) CreateTVSeriesWatchList(uid string, data requests.CreateTVSeriesWatchList) error {
	tvSeriesWatchList := createTVSeriesWatchListObject(
		uid, data.TvID, data.Status, *data.WatchedEpisodes,
		*data.WatchedSeasons, data.Score,
	)

	if _, err := userListModel.TVSeriesWatchListCollection.InsertOne(context.TODO(), tvSeriesWatchList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new tv series watch list: ", err)

		return fmt.Errorf("Failed to create new tv series watch list.")
	}

	return nil
}

//! Update

//! Get
func (userListModel *UserListModel) GetAnimeListByUserID(uid string, data requests.SortList) ([]responses.AnimeList, error) {
	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "popularity":
		sortType = "popularity"
		sortOrder = -1
	case "new":
		sortType = "created_at"
		sortOrder = -1
	case "old":
		sortType = "created_at"
		sortOrder = 1
	case "score":
		sortType = "score"
		sortOrder = -1
	}

	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}

	addFields := bson.M{"$addFields": bson.M{
		"anime_obj_id": bson.M{
			"$toObjectId": "$anime_id",
		},
		"status_priority": bson.M{
			"$switch": bson.M{
				"branches": bson.A{
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$status", "active"},
						},
						"then": 0,
					},
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$status", "finished"},
						},
						"then": 1,
					},
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$status", "dropped"},
						},
						"then": 2,
					},
					bson.M{
						"case": bson.M{
							"$eq": bson.A{"$status", "planto"},
						},
						"then": 3,
					},
				},
				"default": 4,
			},
		},
	}}

	lookup := bson.M{"$lookup": bson.M{
		"from":         "animes",
		"localField":   "anime_obj_id",
		"foreignField": "_id",
		"as":           "anime",
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$anime",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": false,
	}}

	sort := bson.M{"$sort": bson.M{
		"status_priority": 1,
		sortType:          sortOrder,
	}}

	cursor, err := userListModel.AnimeListCollection.Aggregate(context.TODO(), bson.A{
		match, addFields, lookup, unwind, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to aggregate anime list: ", err)

		return nil, fmt.Errorf("Failed to aggregate anime list.")
	}

	var animeList []responses.AnimeList
	if err = cursor.All(context.TODO(), &animeList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to decode anime list: ", err)

		return nil, fmt.Errorf("Failed to decode anime list.")
	}

	return animeList, nil
}

//! Delete
func (userListModel *UserListModel) DeleteListByUserIDAndType(uid string, data requests.DeleteList) (bool, error) {
	objectListID, _ := primitive.ObjectIDFromHex(data.ID)

	var collection mongo.Collection
	switch data.Type {
	case "anime":
		collection = *userListModel.AnimeListCollection
	case "game":
		collection = *userListModel.GameListCollection
	case "movie":
		collection = *userListModel.MovieWatchListCollection
	case "tv":
		collection = *userListModel.TVSeriesWatchListCollection
	}

	count, err := collection.DeleteOne(context.TODO(), bson.M{
		"_id":     objectListID,
		"user_id": uid,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to delete list by user id: ", err)

		return false, fmt.Errorf("Failed to delete list.")
	}

	return count.DeletedCount > 0, nil
}

func (userListModel *UserListModel) DeleteUserListByUserID(uid string) {
	if _, err := userListModel.UserListCollection.DeleteOne(context.TODO(), bson.M{
		"user_id": uid,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to delete user list by user id: ", err)
	}

	collections := [4]mongo.Collection{
		*userListModel.AnimeListCollection,
		*userListModel.GameListCollection,
		*userListModel.MovieWatchListCollection,
		*userListModel.TVSeriesWatchListCollection,
	}

	for i := 0; i < len(collections); i++ {
		go collections[i].DeleteMany(context.TODO(), bson.M{
			"user_id": uid,
		})
	}
}
