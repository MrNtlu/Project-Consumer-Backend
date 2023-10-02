package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserList struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Slug      string             `bson:"slug" json:"slug"`
	IsPublic  bool               `bson:"is_public" json:"is_public"`
	AnimeList []AnimeList        `bson:"anime_list" json:"anime_list"`
	MovieList []MovieList        `bson:"movie_watch_list" json:"movie_watch_list"`
	TVList    []TVSeriesList     `bson:"tv_watch_list" json:"tv_watch_list"`
	GameList  []GameList         `bson:"game_list" json:"game_list"`
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
	StatusSort      int                `bson:"status_sort" json:"status_sort"`
	TitleEn         string             `bson:"title_en" json:"title_en"`
	TitleOriginal   string             `bson:"title_original" json:"title_original"`
	ImageURL        *string            `bson:"image_url" json:"image_url"`
	TotalEpisodes   *int64             `bson:"total_episodes" json:"total_episodes"`
	Type            string             `bson:"type" json:"type"`
	IsAiring        bool               `bson:"is_airing" json:"is_airing"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
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
	StatusSort        int                `bson:"status_sort" json:"status_sort"`
	Title             string             `bson:"title" json:"title"`
	TitleOriginal     string             `bson:"title_original" json:"title_original"`
	ImageURL          *string            `bson:"image_url" json:"image_url"`
	TBA               bool               `bson:"tba" json:"tba"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
}

type MovieList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	ContentStatus string             `bson:"content_status" json:"content_status"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	MovieID       string             `bson:"movie_id" json:"movie_id"`
	MovieTmdbID   string             `bson:"tmdb_id" json:"tmdb_id"`
	Status        string             `bson:"status" json:"status"`
	StatusSort    int                `bson:"status_sort" json:"status_sort"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ImageURL      *string            `bson:"image_url" json:"image_url"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
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
	StatusSort      int                `bson:"status_sort" json:"status_sort"`
	TitleEn         string             `bson:"title_en" json:"title_en"`
	TitleOriginal   string             `bson:"title_original" json:"title_original"`
	ImageURL        *string            `bson:"image_url" json:"image_url"`
	TotalEpisodes   *int64             `bson:"total_episodes" json:"total_episodes"`
	TotalSeasons    *int64             `bson:"total_seasons" json:"total_seasons"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}
