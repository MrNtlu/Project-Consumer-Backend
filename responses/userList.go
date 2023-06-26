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
	AnimeList                 []AnimeList        `bson:"anime_list" json:"anime_list"`
	MovieList                 []MovieList        `bson:"movie_watch_list" json:"movie_watch_list"`
	TVList                    []TVSeriesList     `bson:"tv_watch_list" json:"tv_watch_list"`
	GameList                  []GameList         `bson:"game_list" json:"game_list"`
}

type AnimeList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ContentStatus   string             `bson:"content_status" json:"content_status"`
	Score           *float32           `bson:"score" json:"score"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	AnimeMALID      int64              `bson:"mal_id" json:"mal_id"`
	Status          string             `bson:"status" json:"status"`
	TitleEn         string             `bson:"title_en" json:"title_en"`
	TitleOriginal   string             `bson:"title_original" json:"title_original"`
	ImageURL        *string            `bson:"image_url" json:"image_url"`
	TotalEpisodes   *int64             `bson:"total_episodes" json:"total_episodes"`
	Type            string             `bson:"type" json:"type"`
	IsAiring        bool               `bson:"is_airing" json:"is_airing"`
}

type GameList struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ContentStatus     string             `bson:"content_status" json:"content_status"`
	Score             *float32           `bson:"score" json:"score"`
	AchievementStatus *float32           `bson:"achievement_status" json:"achievement_status"`
	HoursPlayed       *int               `bson:"hours_played" json:"hours_played"`
	TimesFinished     int                `bson:"times_finished" json:"times_finished"`
	GameID            string             `bson:"game_id" json:"game_id"`
	GameRAWGID        int64              `bson:"rawg_id" json:"rawg_id"`
	Status            string             `bson:"status" json:"status"`
	Title             string             `bson:"title" json:"title"`
	TitleOriginal     string             `bson:"title_original" json:"title_original"`
	ImageURL          *string            `bson:"image_url" json:"image_url"`
	TBA               bool               `bson:"tba" json:"tba"`
}

type MovieList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ContentStatus string             `bson:"content_status" json:"content_status"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	MovieID       string             `bson:"movie_id" json:"movie_id"`
	MovieTmdbID   string             `bson:"tmdb_id" json:"tmdb_id"`
	Status        string             `bson:"status" json:"status"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ImageURL      *string            `bson:"image_url" json:"image_url"`
}

type TVSeriesList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ContentStatus   string             `bson:"content_status" json:"content_status"`
	Score           *float32           `bson:"score" json:"score"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	WatchedSeasons  int                `bson:"watched_seasons" json:"watched_seasons"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	TvID            string             `bson:"tv_id" json:"tv_id"`
	TvTmdbID        string             `bson:"tmdb_id" json:"tmdb_id"`
	Status          string             `bson:"status" json:"status"`
	TitleEn         string             `bson:"title_en" json:"title_en"`
	TitleOriginal   string             `bson:"title_original" json:"title_original"`
	ImageURL        *string            `bson:"image_url" json:"image_url"`
	TotalEpisodes   *int64             `bson:"total_episodes" json:"total_episodes"`
	TotalSeasons    *int64             `bson:"total_seasons" json:"total_seasons"`
}
