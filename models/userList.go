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
	AnimeMALID      int64              `bson:"anime_mal_id" json:"anime_mal_id"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	Status          string             `bson:"status" json:"status"`
	WatchedEpisodes int64              `bson:"watched_episodes" json:"watched_episodes"`
	Score           *float32           `bson:"score" json:"score"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"-"`
}

type GameList struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	GameRAWGID        int64              `bson:"game_rawg_id" json:"game_rawg_id"`
	GameID            string             `bson:"game_id" json:"game_id"`
	Status            string             `bson:"status" json:"status"`
	Score             *float32           `bson:"score" json:"score"`
	AchievementStatus *float32           `bson:"achievement_status" json:"achievement_status"`
	TimesFinished     int                `bson:"times_finished" json:"times_finished"`
	HoursPlayed       *int               `bson:"hours_played" json:"hours_played"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"-"`
}

type MovieWatchList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	MovieTmdbID   string             `bson:"movie_tmdb_id" json:"movie_tmdb_id"`
	MovieID       string             `bson:"movie_id" json:"movie_id"`
	Status        string             `bson:"status" json:"status"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"-"`
}

type TVSeriesWatchList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	TvTmdbID        string             `bson:"tv_tmdb_id" json:"tv_tmdb_id"`
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

func createAnimeListObject(userID, animeID, status string, animeMALID, watchedEpisodes int64, score *float32, timesFinished *int) *AnimeList {
	return &AnimeList{
		UserID:          userID,
		AnimeID:         animeID,
		AnimeMALID:      animeMALID,
		Status:          status,
		WatchedEpisodes: watchedEpisodes,
		Score:           score,
		TimesFinished:   handleTimesFinished(status, timesFinished),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

func createGameListObject(userID, gameID, status string, gameRAWGID int64, score, achievementStatus *float32, timesFinished, hoursPlayed *int) *GameList {
	return &GameList{
		UserID:            userID,
		GameRAWGID:        gameRAWGID,
		GameID:            gameID,
		Status:            status,
		Score:             score,
		AchievementStatus: achievementStatus,
		TimesFinished:     handleTimesFinished(status, timesFinished),
		HoursPlayed:       hoursPlayed,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
}

func createMovieWatchListObject(userID, movieTmdbID, movieID, status string, score *float32, timesFinished *int) *MovieWatchList {
	return &MovieWatchList{
		UserID:        userID,
		MovieTmdbID:   movieTmdbID,
		MovieID:       movieID,
		Status:        status,
		Score:         score,
		TimesFinished: handleTimesFinished(status, timesFinished),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

func createTVSeriesWatchListObject(userID, tvTmdbID, tvID, status string, watchedEpisodes, watchedSeasons int, score *float32, timesFinished *int) *TVSeriesWatchList {
	return &TVSeriesWatchList{
		UserID:          userID,
		TvTmdbID:        tvTmdbID,
		TvID:            tvID,
		Status:          status,
		Score:           score,
		WatchedEpisodes: watchedEpisodes,
		WatchedSeasons:  watchedSeasons,
		TimesFinished:   handleTimesFinished(status, timesFinished),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

func handleTimesFinished(status string, timesFinished *int) int {
	if timesFinished == nil && status == "finished" {
		return 1
	} else if timesFinished == nil {
		return 0
	}
	return *timesFinished
}

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

func (userListModel *UserListModel) CreateAnimeList(uid string, data requests.CreateAnimeList, anime responses.Anime) (AnimeList, error) {
	animeList := createAnimeListObject(
		uid, data.AnimeID, data.Status, data.AnimeMALID,
		*data.WatchedEpisodes, data.Score, data.TimesFinished,
	)

	if anime.Episodes != nil {
		if *data.WatchedEpisodes > *anime.Episodes {
			animeList.WatchedEpisodes = *anime.Episodes
		} else if data.Status == "finished" && *data.WatchedEpisodes < *anime.Episodes {
			animeList.WatchedEpisodes = *anime.Episodes
		}
	}

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = userListModel.AnimeListCollection.InsertOne(context.TODO(), animeList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new anime list: ", err)

		return AnimeList{}, fmt.Errorf("Failed to create new anime list.")
	}

	animeList.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *animeList, nil
}

func (userListModel *UserListModel) CreateGameList(uid string, data requests.CreateGameList) (GameList, error) {
	gameList := createGameListObject(
		uid, data.GameID, data.Status, data.GameRAWGID,
		data.Score, data.AchievementStatus,
		data.TimesFinished, data.HoursPlayed,
	)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = userListModel.GameListCollection.InsertOne(context.TODO(), gameList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new game list: ", err)

		return GameList{}, fmt.Errorf("Failed to create new game list.")
	}

	gameList.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *gameList, nil
}

func (userListModel *UserListModel) CreateMovieWatchList(uid string, data requests.CreateMovieWatchList) (MovieWatchList, error) {
	movieWatchList := createMovieWatchListObject(uid, data.MovieTmdbID, data.MovieID, data.Status, data.Score, data.TimesFinished)

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = userListModel.MovieWatchListCollection.InsertOne(context.TODO(), movieWatchList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new movie watch list: ", err)

		return MovieWatchList{}, fmt.Errorf("Failed to create new movie watch list.")
	}

	movieWatchList.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *movieWatchList, nil
}

func (userListModel *UserListModel) CreateTVSeriesWatchList(
	uid string, data requests.CreateTVSeriesWatchList, tvSeries responses.TVSeries,
) (TVSeriesWatchList, error) {
	tvSeriesWatchList := createTVSeriesWatchListObject(
		uid, data.TvTmdbID, data.TvID, data.Status,
		*data.WatchedEpisodes, *data.WatchedSeasons, data.Score,
		data.TimesFinished,
	)

	if (*data.WatchedEpisodes > tvSeries.TotalEpisodes) || (data.Status == "finished" && *data.WatchedEpisodes < tvSeries.TotalEpisodes) {
		tvSeriesWatchList.WatchedEpisodes = tvSeries.TotalEpisodes
	}

	if (*data.WatchedSeasons > tvSeries.TotalSeasons) || (data.Status == "finished" && *data.WatchedSeasons < tvSeries.TotalSeasons) {
		tvSeriesWatchList.WatchedSeasons = tvSeries.TotalSeasons
	}

	var (
		insertedID *mongo.InsertOneResult
		err        error
	)

	if insertedID, err = userListModel.TVSeriesWatchListCollection.InsertOne(context.TODO(), tvSeriesWatchList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"data": data,
		}).Error("failed to create new tv series watch list: ", err)

		return TVSeriesWatchList{}, fmt.Errorf("Failed to create new tv series watch list.")
	}

	tvSeriesWatchList.ID = insertedID.InsertedID.(primitive.ObjectID)

	return *tvSeriesWatchList, nil
}

//! Update
func (userListModel *UserListModel) UpdateUserListPublicVisibility(userList UserList, data requests.UpdateUserList) error {
	if data.IsPublic != userList.IsPublic {
		if _, err := userListModel.UserListCollection.UpdateOne(context.TODO(), bson.M{
			"_id": userList.ID,
		}, bson.M{"$set": bson.M{
			"is_public": data.IsPublic,
		}}); err != nil {
			logrus.WithFields(logrus.Fields{
				"user_list_id": userList.ID,
				"data":         data,
			}).Error("failed to update user list visibility: ", err)

			return fmt.Errorf("Failed to update user list visibility.")
		}
	}

	return nil
}

func (userListModel *UserListModel) UpdateAnimeListByID(animeList AnimeList, data requests.UpdateAnimeList) (AnimeList, error) {
	if data.IsUpdatingScore || data.TimesFinished != nil ||
		data.Status != nil || data.WatchedEpisodes != nil {
		set := bson.M{}

		if data.IsUpdatingScore && animeList.Score != data.Score {
			set["score"] = data.Score
			animeList.Score = data.Score
		}

		if data.TimesFinished != nil && animeList.TimesFinished != *data.TimesFinished {
			set["times_finished"] = *data.TimesFinished
			animeList.TimesFinished = *data.TimesFinished
		}

		if data.Status != nil && animeList.Status != *data.Status {
			set["status"] = *data.Status
			animeList.Status = *data.Status
		}

		if data.WatchedEpisodes != nil && animeList.WatchedEpisodes != *data.WatchedEpisodes {
			set["watched_episodes"] = *data.WatchedEpisodes
			animeList.WatchedEpisodes = *data.WatchedEpisodes
		}

		if _, err := userListModel.AnimeListCollection.UpdateOne(context.TODO(), bson.M{
			"_id": animeList.ID,
		}, bson.M{"$set": set}); err != nil {
			logrus.WithFields(logrus.Fields{
				"anime_list_id": animeList.ID,
				"data":          data,
			}).Error("failed to update anime list: ", err)

			return AnimeList{}, fmt.Errorf("Failed to update anime list.")
		}
	}

	return animeList, nil
}

func (userListModel *UserListModel) UpdateGameListByID(gameList GameList, data requests.UpdateGameList) (GameList, error) {
	if data.IsUpdatingScore || data.TimesFinished != nil ||
		data.Status != nil || data.AchievementStatus != nil {
		set := bson.M{}

		if data.IsUpdatingScore && gameList.Score != data.Score {
			set["score"] = data.Score
			gameList.Score = data.Score
		}

		if data.TimesFinished != nil && gameList.TimesFinished != *data.TimesFinished {
			set["times_finished"] = *data.TimesFinished
			gameList.TimesFinished = *data.TimesFinished
		}

		if data.Status != nil && gameList.Status != *data.Status {
			set["status"] = *data.Status
			gameList.Status = *data.Status
		}

		if data.HoursPlayed != nil && gameList.HoursPlayed != data.HoursPlayed {
			set["hours_played"] = data.HoursPlayed
			gameList.HoursPlayed = data.HoursPlayed
		}

		if data.AchievementStatus != nil && gameList.AchievementStatus != data.AchievementStatus {
			set["achievement_status"] = data.AchievementStatus
			gameList.AchievementStatus = data.AchievementStatus
		}

		if _, err := userListModel.GameListCollection.UpdateOne(context.TODO(), bson.M{
			"_id": gameList.ID,
		}, bson.M{"$set": set}); err != nil {
			logrus.WithFields(logrus.Fields{
				"game_list_id": gameList.ID,
				"data":         data,
			}).Error("failed to update game list: ", err)

			return GameList{}, fmt.Errorf("Failed to update game list.")
		}
	}

	return gameList, nil
}

func (userListModel *UserListModel) UpdateMovieListByID(movieList MovieWatchList, data requests.UpdateMovieList) (MovieWatchList, error) {
	if data.IsUpdatingScore || data.TimesFinished != nil || data.Status != nil {
		set := bson.M{}

		if data.IsUpdatingScore && movieList.Score != data.Score {
			set["score"] = data.Score
			movieList.Score = data.Score
		}

		if data.TimesFinished != nil && movieList.TimesFinished != *data.TimesFinished {
			set["times_finished"] = *data.TimesFinished
			movieList.TimesFinished = *data.TimesFinished
		}

		if data.Status != nil && movieList.Status != *data.Status {
			set["status"] = *data.Status
			movieList.Status = *data.Status
		}

		if _, err := userListModel.MovieWatchListCollection.UpdateOne(context.TODO(), bson.M{
			"_id": movieList.ID,
		}, bson.M{"$set": set}); err != nil {
			logrus.WithFields(logrus.Fields{
				"movie_list_id": movieList.ID,
				"data":          data,
			}).Error("failed to update movie list: ", err)

			return MovieWatchList{}, fmt.Errorf("Failed to update movie watch list.")
		}
	}

	return movieList, nil
}

func (userListModel *UserListModel) UpdateTVSeriesListByID(tvList TVSeriesWatchList, data requests.UpdateTVSeriesList) (TVSeriesWatchList, error) {
	if data.IsUpdatingScore || data.TimesFinished != nil ||
		data.Status != nil || data.WatchedEpisodes != nil ||
		data.WatchedSeasons != nil {
		set := bson.M{}

		if data.IsUpdatingScore && tvList.Score != data.Score {
			set["score"] = data.Score
			tvList.Score = data.Score
		}

		if data.TimesFinished != nil && tvList.TimesFinished != *data.TimesFinished {
			set["times_finished"] = *data.TimesFinished
			tvList.TimesFinished = *data.TimesFinished
		}

		if data.Status != nil && tvList.Status != *data.Status {
			set["status"] = *data.Status
			tvList.Status = *data.Status
		}

		if data.WatchedEpisodes != nil && tvList.WatchedEpisodes != *data.WatchedEpisodes {
			set["watched_episodes"] = *data.WatchedEpisodes
			tvList.WatchedEpisodes = *data.WatchedEpisodes
		}

		if data.WatchedSeasons != nil && tvList.WatchedSeasons != *data.WatchedSeasons {
			set["watched_seasons"] = *data.WatchedSeasons
			tvList.WatchedSeasons = *data.WatchedSeasons
		}

		if _, err := userListModel.TVSeriesWatchListCollection.UpdateOne(context.TODO(), bson.M{
			"_id": tvList.ID,
		}, bson.M{"$set": set}); err != nil {
			logrus.WithFields(logrus.Fields{
				"tv_list_id": tvList.ID,
				"data":       data,
			}).Error("failed to update tv list: ", err)

			return TVSeriesWatchList{}, fmt.Errorf("Failed to update tv series watch list.")
		}
	}

	return tvList, nil
}

//! Get
func (userListModel *UserListModel) GetBaseUserListByUserID(uid string) (UserList, error) {
	result := userListModel.UserListCollection.FindOne(context.TODO(), bson.M{"user_id": uid})

	var userList UserList
	if err := result.Decode(&userList); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
		}).Error("failed to find user list by user id: ", err)

		return UserList{}, fmt.Errorf("Failed to find user list by user id.")
	}

	return userList, nil
}

func (userListModel *UserListModel) GetBaseAnimeListByID(animeListID string) (AnimeList, error) {
	objectID, _ := primitive.ObjectIDFromHex(animeListID)

	result := userListModel.AnimeListCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var animeList AnimeList
	if err := result.Decode(&animeList); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": animeListID,
		}).Error("failed to find anime list by id: ", err)

		return AnimeList{}, fmt.Errorf("Failed to find anime list by id.")
	}

	return animeList, nil
}

func (userListModel *UserListModel) GetBaseGameListByID(gameListID string) (GameList, error) {
	objectID, _ := primitive.ObjectIDFromHex(gameListID)

	result := userListModel.GameListCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var gameList GameList
	if err := result.Decode(&gameList); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": gameListID,
		}).Error("failed to find game list by id: ", err)

		return GameList{}, fmt.Errorf("Failed to find game list by id.")
	}

	return gameList, nil
}

func (userListModel *UserListModel) GetBaseMovieListByID(movieID string) (MovieWatchList, error) {
	objectID, _ := primitive.ObjectIDFromHex(movieID)

	result := userListModel.MovieWatchListCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var movieList MovieWatchList
	if err := result.Decode(&movieList); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": movieID,
		}).Error("failed to find movie list by id: ", err)

		return MovieWatchList{}, fmt.Errorf("Failed to find movie watch list by id.")
	}

	return movieList, nil
}

func (userListModel *UserListModel) GetBaseTVSeriesListByID(movieID string) (TVSeriesWatchList, error) {
	objectID, _ := primitive.ObjectIDFromHex(movieID)

	result := userListModel.TVSeriesWatchListCollection.FindOne(context.TODO(), bson.M{"_id": objectID})

	var tvList TVSeriesWatchList
	if err := result.Decode(&tvList); err != nil {
		logrus.WithFields(logrus.Fields{
			"id": movieID,
		}).Error("failed to find tv series list by id: ", err)

		return TVSeriesWatchList{}, fmt.Errorf("Failed to find tv series watch list by id.")
	}

	return tvList, nil
}

func (userListModel *UserListModel) GetUserListByUserID(uid string, data requests.SortList) (responses.UserList, error) {
	match := bson.M{"$match": bson.M{
		"user_id": uid,
	}}

	animeListLookup := bson.M{"$lookup": bson.M{
		"from":         "anime-lists",
		"localField":   "user_id",
		"foreignField": "user_id",
		"as":           "anime_list",
	}}

	gameListLookup := bson.M{"$lookup": bson.M{
		"from":         "game-lists",
		"localField":   "user_id",
		"foreignField": "user_id",
		"as":           "game_list",
	}}

	movieListLookup := bson.M{"$lookup": bson.M{
		"from":         "movie-watch-lists",
		"localField":   "user_id",
		"foreignField": "user_id",
		"as":           "movie_watch_list",
	}}

	tvListLookup := bson.M{"$lookup": bson.M{
		"from":         "tvseries-watch-lists",
		"localField":   "user_id",
		"foreignField": "user_id",
		"as":           "tv_watch_list",
	}}

	addFields := bson.M{"$addFields": bson.M{
		"anime_count": bson.M{
			"$size": "$anime_list",
		},
		"game_count": bson.M{
			"$size": "$game_list",
		},
		"movie_count": bson.M{
			"$size": "$movie_watch_list",
		},
		"tv_count": bson.M{
			"$size": "$tv_watch_list",
		},
		"anime_total_watched_episodes": bson.M{
			"$sum": "$anime_list.watched_episodes",
		},
		"tv_total_watched_episodes": bson.M{
			"$sum": "$tv_watch_list.watched_episodes",
		},
		"anime_total_finished": bson.M{
			"$sum": "$anime_list.times_finished",
		},
		"game_total_finished": bson.M{
			"$sum": "$game_list.times_finished",
		},
		"movie_total_finished": bson.M{
			"$sum": "$movie_watch_list.times_finished",
		},
		"tv_total_finished": bson.M{
			"$sum": "$tv_watch_list.times_finished",
		},
		"anime_avg_score": bson.M{
			"$divide": bson.A{
				bson.M{
					"$sum": "$anime_list.score",
				},
				bson.M{
					"$cond": bson.A{
						bson.M{
							"$gt": bson.A{
								bson.M{"$size": "$anime_list"},
								0,
							},
						},
						bson.M{
							"$size": "$anime_list",
						},
						1,
					},
				},
			},
		},
		"game_avg_score": bson.M{
			"$divide": bson.A{
				bson.M{
					"$sum": "$game_list.score",
				},
				bson.M{
					"$cond": bson.A{
						bson.M{
							"$gt": bson.A{
								bson.M{"$size": "$game_list"},
								0,
							},
						},
						bson.M{
							"$size": "$game_list",
						},
						1,
					},
				},
			},
		},
		"movie_avg_score": bson.M{
			"$divide": bson.A{
				bson.M{
					"$sum": "$movie_watch_list.score",
				},
				bson.M{
					"$cond": bson.A{
						bson.M{
							"$gt": bson.A{
								bson.M{"$size": "$movie_watch_list"},
								0,
							},
						},
						bson.M{
							"$size": "$movie_watch_list",
						},
						1,
					},
				},
			},
		},
		"tv_avg_score": bson.M{
			"$divide": bson.A{
				bson.M{
					"$sum": "$tv_watch_list.score",
				},
				bson.M{
					"$cond": bson.A{
						bson.M{
							"$gt": bson.A{
								bson.M{"$size": "$tv_watch_list"},
								0,
							},
						},
						bson.M{
							"$size": "$tv_watch_list",
						},
						1,
					},
				},
			},
		},
	}}

	facet := bson.M{"$facet": bson.M{
		"lookups": bson.A{
			bson.M{
				"$lookup": bson.M{
					"from": "movies",
					"let": bson.M{
						"movie_id":       "$movie_watch_list.movie_id",
						"tmdb_id":        "$movie_watch_list.movie_tmdb_id",
						"id":             "$movie_watch_list._id",
						"status":         "$movie_watch_list.status",
						"score":          "$movie_watch_list.score",
						"times_finished": "$movie_watch_list.times_finished",
					},
					"pipeline": bson.A{
						bson.M{
							"$addFields": bson.M{
								"movie_id": bson.M{
									"$toString": "$_id",
								},
							},
						},
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{
											"$in": bson.A{"$movie_id", "$$movie_id"},
										},
										bson.M{
											"$in": bson.A{"$tmdb_id", "$$tmdb_id"},
										},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title_en":       1,
								"status":         1,
								"movie_id":       1,
								"tmdb_id":        1,
								"_id": bson.M{
									"$arrayElemAt": bson.A{
										"$$id",
										bson.M{
											"$indexOfArray": bson.A{
												"$$movie_id",
												"$movie_id",
											},
										},
									},
								},
								"status_sort": bson.M{
									"$switch": bson.M{
										"branches": bson.A{
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$movie_id",
																		"$movie_id",
																	},
																},
															},
														},
														"finished",
													},
												},
												"then": 1,
											},
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$movie_id",
																		"$movie_id",
																	},
																},
															},
														},
														"active",
													},
												},
												"then": 0,
											},
										},
										"default": 2,
									},
								},
								"content_status": bson.M{
									"$arrayElemAt": bson.A{
										"$$status",
										bson.M{
											"$indexOfArray": bson.A{
												"$$movie_id",
												"$movie_id",
											},
										},
									},
								},
								"score": bson.M{
									"$arrayElemAt": bson.A{
										"$$score",
										bson.M{
											"$indexOfArray": bson.A{
												"$$movie_id",
												"$movie_id",
											},
										},
									},
								},
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$movie_id",
												"$movie_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "movie_watch_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "tv-series",
					"let": bson.M{
						"tv_id":            "$tv_watch_list.tv_id",
						"tmdb_id":          "$tv_watch_list.tv_tmdb_id",
						"id":               "$tv_watch_list._id",
						"status":           "$tv_watch_list.status",
						"score":            "$tv_watch_list.score",
						"times_finished":   "$tv_watch_list.times_finished",
						"watched_episodes": "$tv_watch_list.watched_episodes",
						"watched_seasons":  "$tv_watch_list.watched_seasons",
					},
					"pipeline": bson.A{
						bson.M{
							"$addFields": bson.M{
								"tv_id": bson.M{
									"$toString": "$_id",
								},
							},
						},
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{
											"$in": bson.A{"$tv_id", "$$tv_id"},
										},
										bson.M{
											"$in": bson.A{"$tmdb_id", "$$tmdb_id"},
										},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title_en":       1,
								"status":         1,
								"total_episodes": 1,
								"total_seasons":  1,
								"tv_id":          1,
								"tmdb_id":        1,
								"_id": bson.M{
									"$arrayElemAt": bson.A{
										"$$id",
										bson.M{
											"$indexOfArray": bson.A{
												"$$tv_id",
												"$tv_id",
											},
										},
									},
								},
								"status_sort": bson.M{
									"$switch": bson.M{
										"branches": bson.A{
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$tv_id",
																		"$tv_id",
																	},
																},
															},
														},
														"finished",
													},
												},
												"then": 1,
											},
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$tv_id",
																		"$tv_id",
																	},
																},
															},
														},
														"active",
													},
												},
												"then": 0,
											},
										},
										"default": 2,
									},
								},
								"content_status": bson.M{
									"$arrayElemAt": bson.A{
										"$$status",
										bson.M{
											"$indexOfArray": bson.A{
												"$$tv_id",
												"$tv_id",
											},
										},
									},
								},
								"score": bson.M{
									"$arrayElemAt": bson.A{
										"$$score",
										bson.M{
											"$indexOfArray": bson.A{
												"$$tv_id",
												"$tv_id",
											},
										},
									},
								},
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$tv_id",
												"$tv_id",
											},
										},
									},
								},
								"watched_episodes": bson.M{
									"$arrayElemAt": bson.A{
										"$$watched_episodes",
										bson.M{
											"$indexOfArray": bson.A{
												"$$tv_id",
												"$tv_id",
											},
										},
									},
								},
								"watched_seasons": bson.M{
									"$arrayElemAt": bson.A{
										"$$watched_seasons",
										bson.M{
											"$indexOfArray": bson.A{
												"$$tv_id",
												"$tv_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "tv_watch_list",
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from": "animes",
					"let": bson.M{
						"anime_id":         "$anime_list.anime_id",
						"mal_id":           "$anime_list.anime_mal_id",
						"id":               "$anime_list._id",
						"status":           "$anime_list.status",
						"score":            "$anime_list.score",
						"times_finished":   "$anime_list.times_finished",
						"watched_episodes": "$anime_list.watched_episodes",
					},
					"pipeline": bson.A{
						bson.M{
							"$addFields": bson.M{
								"anime_id": bson.M{
									"$toString": "$_id",
								},
							},
						},
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{
											"$in": bson.A{"$anime_id", "$$anime_id"},
										},
										bson.M{
											"$in": bson.A{"$mal_id", "$$mal_id"},
										},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title_en":       1,
								"status":         1,
								"type":           1,
								"total_episodes": 1,
								"anime_id":       1,
								"mal_id":         1,
								"is_airing":      1,
								"_id": bson.M{
									"$arrayElemAt": bson.A{
										"$$id",
										bson.M{
											"$indexOfArray": bson.A{
												"$$anime_id",
												"$anime_id",
											},
										},
									},
								},
								"status_sort": bson.M{
									"$switch": bson.M{
										"branches": bson.A{
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$anime_id",
																		"$anime_id",
																	},
																},
															},
														},
														"finished",
													},
												},
												"then": 1,
											},
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$anime_id",
																		"$anime_id",
																	},
																},
															},
														},
														"active",
													},
												},
												"then": 0,
											},
										},
										"default": 2,
									},
								},
								"content_status": bson.M{
									"$arrayElemAt": bson.A{
										"$$status",
										bson.M{
											"$indexOfArray": bson.A{
												"$$anime_id",
												"$anime_id",
											},
										},
									},
								},
								"score": bson.M{
									"$arrayElemAt": bson.A{
										"$$score",
										bson.M{
											"$indexOfArray": bson.A{
												"$$anime_id",
												"$anime_id",
											},
										},
									},
								},
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$anime_id",
												"$anime_id",
											},
										},
									},
								},
								"watched_episodes": bson.M{
									"$arrayElemAt": bson.A{
										"$$watched_episodes",
										bson.M{
											"$indexOfArray": bson.A{
												"$$anime_id",
												"$anime_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "anime_list",
				},
			},

			bson.M{
				"$lookup": bson.M{
					"from": "games",
					"let": bson.M{
						"game_id":            "$game_list.game_id",
						"rawg_id":            "$game_list.game_rawg_id",
						"id":                 "$game_list._id",
						"status":             "$game_list.status",
						"score":              "$game_list.score",
						"times_finished":     "$game_list.times_finished",
						"achievement_status": "$game_list.achievement_status",
						"hours_played":       "$game_list.hours_played",
					},
					"pipeline": bson.A{
						bson.M{
							"$addFields": bson.M{
								"game_id": bson.M{
									"$toString": "$_id",
								},
							},
						},
						bson.M{
							"$match": bson.M{
								"$expr": bson.M{
									"$or": bson.A{
										bson.M{
											"$in": bson.A{"$game_id", "$$game_id"},
										},
										bson.M{
											"$in": bson.A{"$rawg_id", "$$rawg_id"},
										},
									},
								},
							},
						},
						bson.M{
							"$project": bson.M{
								"image_url":      1,
								"title_original": 1,
								"title":          1,
								"tba":            1,
								"game_id":        1,
								"rawg_id":        1,
								"_id": bson.M{
									"$arrayElemAt": bson.A{
										"$$id",
										bson.M{
											"$indexOfArray": bson.A{
												"$$game_id",
												"$game_id",
											},
										},
									},
								},
								"status_sort": bson.M{
									"$switch": bson.M{
										"branches": bson.A{
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$game_id",
																		"$game_id",
																	},
																},
															},
														},
														"finished",
													},
												},
												"then": 1,
											},
											bson.M{
												"case": bson.M{
													"$eq": bson.A{
														bson.M{
															"$arrayElemAt": bson.A{
																"$$status",
																bson.M{
																	"$indexOfArray": bson.A{
																		"$$game_id",
																		"$game_id",
																	},
																},
															},
														},
														"active",
													},
												},
												"then": 0,
											},
										},
										"default": 2,
									},
								},
								"content_status": bson.M{
									"$arrayElemAt": bson.A{
										"$$status",
										bson.M{
											"$indexOfArray": bson.A{
												"$$game_id",
												"$game_id",
											},
										},
									},
								},
								"score": bson.M{
									"$arrayElemAt": bson.A{
										"$$score",
										bson.M{
											"$indexOfArray": bson.A{
												"$$game_id",
												"$game_id",
											},
										},
									},
								},
								"times_finished": bson.M{
									"$arrayElemAt": bson.A{
										"$$times_finished",
										bson.M{
											"$indexOfArray": bson.A{
												"$$game_id",
												"$game_id",
											},
										},
									},
								},
								"achievement_status": bson.M{
									"$arrayElemAt": bson.A{
										"$$achievement_status",
										bson.M{
											"$indexOfArray": bson.A{
												"$$game_id",
												"$game_id",
											},
										},
									},
								},
								"hours_played": bson.M{
									"$arrayElemAt": bson.A{
										"$$hours_played",
										bson.M{
											"$indexOfArray": bson.A{
												"$$game_id",
												"$game_id",
											},
										},
									},
								},
							},
						},
					},
					"as": "game_list",
				},
			},
		},
	}}

	unwind := bson.M{"$unwind": bson.M{
		"path":                       "$lookups",
		"includeArrayIndex":          "index",
		"preserveNullAndEmptyArrays": true,
	}}

	replaceRoot := bson.M{"$replaceRoot": bson.M{
		"newRoot": "$lookups",
	}}

	var (
		sortType  string
		sortOrder int8
	)

	switch data.Sort {
	case "score":
		sortType = "score"
		sortOrder = -1
	case "timeswatched":
		sortType = "times_finished"
		sortOrder = -1
	}

	sort := bson.M{"$set": bson.M{
		"movie_watch_list": bson.M{
			"$sortArray": bson.M{
				"input": "$movie_watch_list",
				"sortBy": bson.M{
					"status_sort": 1,
					sortType:      sortOrder,
				},
			},
		},
		"anime_list": bson.M{
			"$sortArray": bson.M{
				"input": "$anime_list",
				"sortBy": bson.M{
					"status_sort": 1,
					sortType:      sortOrder,
				},
			},
		},
		"game_list": bson.M{
			"$sortArray": bson.M{
				"input": "$game_list",
				"sortBy": bson.M{
					"status_sort": 1,
					sortType:      sortOrder,
				},
			},
		},
		"tv_watch_list": bson.M{
			"$sortArray": bson.M{
				"input": "$tv_watch_list",
				"sortBy": bson.M{
					"status_sort": 1,
					sortType:      sortOrder,
				},
			},
		},
	}}

	cursor, err := userListModel.UserListCollection.Aggregate(context.TODO(), bson.A{
		match, animeListLookup, gameListLookup, movieListLookup,
		tvListLookup, addFields, facet, unwind, replaceRoot, sort,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to aggregate user list: ", err)

		return responses.UserList{}, fmt.Errorf("Failed to aggregate user list.")
	}

	var userList []responses.UserList
	if err = cursor.All(context.TODO(), &userList); err != nil {
		logrus.WithFields(logrus.Fields{
			"uid": uid,
		}).Error("failed to decode user list: ", err)

		return responses.UserList{}, fmt.Errorf("Failed to decode user list.")
	}

	if len(userList) > 0 {
		return userList[0], nil
	}

	return responses.UserList{}, nil
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

//! Others
func getListSortHandler(sort string) (string, int8) {
	switch sort {
	case "popularity":
		return "popularity", -1
	case "new":
		return "created_at", -1
	case "old":
		return "created_at", 1
	case "score":
		return "score", -1
	default:
		return "", -1
	}
}

func statusPriorityAggregation() bson.M {
	return bson.M{
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
	}
}
