package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserList struct {
	ID                        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID                    string             `bson:"user_id" json:"user_id"`
	Slug                      string             `bson:"slug" json:"slug"`
	IsPublic                  bool               `bson:"is_public" json:"is_public"`
	AnimeCount                int32              `bson:"anime_count" json:"anime_count"`
	GameCount                 int32              `bson:"game_count" json:"game_count"`
	MovieCount                int32              `bson:"movie_count" json:"movie_count"`
	TVCount                   int32              `bson:"tv_count" json:"tv_count"`
	AnimeTotalWatchedEpisodes int64              `bson:"anime_total_watched_episodes" json:"anime_total_watched_episodes"`
	TVTotalWatchedEpisodes    int64              `bson:"tv_total_watched_episodes" json:"tv_total_watched_episodes"`
	AnimeTotalFinished        int64              `bson:"anime_total_finished" json:"anime_total_finished"`
	GameTotalFinished         int64              `bson:"game_total_finished" json:"game_total_finished"`
	MovieTotalFinished        int64              `bson:"movie_total_finished" json:"movie_total_finished"`
	TVTotalFinished           int64              `bson:"tv_total_finished" json:"tv_total_finished"`
	AnimeAvgScore             float64            `bson:"anime_avg_score" json:"anime_avg_score"`
	GameAvgScore              float64            `bson:"game_avg_score" json:"game_avg_score"`
	MovieAvgScore             float64            `bson:"movie_avg_score" json:"movie_avg_score"`
	TVAvgScore                float64            `bson:"tv_avg_score" json:"tv_avg_score"`
}

type AnimeList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	AnimeMALID      string             `bson:"anime_mal_id" json:"anime_mal_id"`
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
	GameRAWGID        int64              `bson:"game_rawg_id" json:"game_rawg_id"`
	Status            string             `bson:"status" json:"status"`
	Score             *float32           `bson:"score" json:"score"`
	AchievementStatus *float32           `bson:"achievement_status" json:"achievement_status"`
	TimesFinished     int                `bson:"times_finished" json:"times_finished"`
	Game              Game               `bson:"game" json:"game"`
}

type MovieList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	MovieID       string             `bson:"movie_id" json:"movie_id"`
	MovieTmdbID   string             `bson:"movie_tmdb_id" json:"movie_tmdb_id"`
	Status        string             `bson:"status" json:"status"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	Movie         Movie              `bson:"movie" json:"movie"`
}

type TVSeriesList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	TvID            string             `bson:"tv_id" json:"tv_id"`
	TvTmdbID        string             `bson:"tv_tmdb_id" json:"tv_tmdb_id"`
	Status          string             `bson:"status" json:"status"`
	Score           *float32           `bson:"score" json:"score"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	WatchedSeasons  int                `bson:"watched_seasons" json:"watched_seasons"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	TVSeries        TVSeries           `bson:"tv_series" json:"tv_series"`
}
