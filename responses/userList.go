package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserList struct {
	ID                        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID                    string             `bson:"user_id" json:"user_id"`
	Slug                      string             `bson:"slug" json:"slug"`
	IsPublic                  bool               `bson:"is_public" json:"is_public"`
	AnimeCount                int32              `bson:"anime_count" json:"anime_count"`
	GameCount                 int32              `bson:"game_count" json:"game_count"`
	AnimeTotalWatchedEpisodes int64              `bson:"anime_total_watched_episodes" json:"anime_total_watched_episodes"`
	GameTotalFinished         int64              `bson:"game_total_finished" json:"game_total_finished"`
	AnimeAvgScore             float64            `bson:"anime_avg_score" json:"anime_avg_score"`
	GameAvgScore              float64            `bson:"game_avg_score" json:"game_avg_score"`
}

type AnimeList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	Status          string             `bson:"status" json:"status"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	Score           *float32           `bson:"score" json:"score"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	Anime           Anime              `bson:"anime" json:"anime"`
}

type GameList struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	GameID            string             `bson:"game_id" json:"game_id"`
	Status            string             `bson:"status" json:"status"`
	Score             *float32           `bson:"score" json:"score"`
	AchievementStatus *float32           `bson:"achievement_status" json:"achievement_status"`
	TimesFinished     int                `bson:"times_finished" json:"times_finished"`
	Game              Game               `bson:"game" json:"game"`
}

type MovieList struct {
	UserID        string   `bson:"user_id" json:"user_id"`
	MovieID       string   `bson:"movie_id" json:"movie_id"`
	Status        string   `bson:"status" json:"status"`
	Score         *float32 `bson:"score" json:"score"`
	TimesFinished int      `bson:"times_finished" json:"times_finished"`
	Movie         Movie    `bson:"movie" json:"movie"`
}
