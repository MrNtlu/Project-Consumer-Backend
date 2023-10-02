package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Game struct {
	ID                        primitive.ObjectID            `bson:"_id,omitempty" json:"_id"`
	Title                     string                        `bson:"title" json:"title"`
	TitleOriginal             string                        `bson:"title_original" json:"title_original"`
	Description               string                        `bson:"description" json:"description"`
	TBA                       bool                          `bson:"tba" json:"tba"`
	RawgID                    int64                         `bson:"rawg_id" json:"rawg_id"`
	Popularity                float64                       `bson:"popularity" json:"popularity"`
	RawgRating                float64                       `bson:"rawg_rating" json:"rawg_rating"`
	RawgRatingCount           int64                         `bson:"rawg_rating_count" json:"rawg_rating_count"`
	MetacriticScore           int64                         `bson:"metacritic_score" json:"metacritic_score"`
	MetacriticScoreByPlatform []GameMetacriticScorePlatform `bson:"metacritic_score_by_platform" json:"metacritic_score_by_platform"`
	ReleaseDate               string                        `bson:"release_date" json:"release_date"`
	ImageUrl                  string                        `bson:"image_url" json:"image_url"`
	AgeRating                 *string                       `bson:"age_rating" json:"age_rating"`
	RelatedGames              []GameRelation                `bson:"related_games" json:"related_games"`
	Genres                    []string                      `bson:"genres" json:"genres"`
	Screenshots               []string                      `bson:"screenshots" json:"screenshots"`
	Tags                      []string                      `bson:"tags" json:"tags"`
	Platforms                 []string                      `bson:"platforms" json:"platforms"`
	Developers                []string                      `bson:"developers" json:"developers"`
	Publishers                []string                      `bson:"publishers" json:"publishers"`
	Stores                    []GameStore                   `bson:"stores" json:"stores"`
}

type GameDetails struct {
	ID                        primitive.ObjectID            `bson:"_id,omitempty" json:"_id"`
	Title                     string                        `bson:"title" json:"title"`
	TitleOriginal             string                        `bson:"title_original" json:"title_original"`
	Description               string                        `bson:"description" json:"description"`
	TBA                       bool                          `bson:"tba" json:"tba"`
	RawgID                    int64                         `bson:"rawg_id" json:"rawg_id"`
	RawgRating                float64                       `bson:"rawg_rating" json:"rawg_rating"`
	RawgRatingCount           int64                         `bson:"rawg_rating_count" json:"rawg_rating_count"`
	MetacriticScore           int64                         `bson:"metacritic_score" json:"metacritic_score"`
	MetacriticScoreByPlatform []GameMetacriticScorePlatform `bson:"metacritic_score_by_platform" json:"metacritic_score_by_platform"`
	ReleaseDate               string                        `bson:"release_date" json:"release_date"`
	ImageUrl                  string                        `bson:"image_url" json:"image_url"`
	AgeRating                 *string                       `bson:"age_rating" json:"age_rating"`
	RelatedGames              []GameDetailsRelation         `bson:"related_games" json:"related_games"`
	Genres                    []string                      `bson:"genres" json:"genres"`
	Screenshots               []string                      `bson:"screenshots" json:"screenshots"`
	Tags                      []string                      `bson:"tags" json:"tags"`
	Platforms                 []string                      `bson:"platforms" json:"platforms"`
	Developers                []string                      `bson:"developers" json:"developers"`
	Publishers                []string                      `bson:"publishers" json:"publishers"`
	Stores                    []GameStore                   `bson:"stores" json:"stores"`
	GameList                  *GamePlayList                 `bson:"game_list" json:"game_list"`
	WatchLater                *ConsumeLater                 `bson:"watch_later" json:"watch_later"`
}

type GamePlayList struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	GameID            string             `bson:"game_id" json:"game_id"`
	GameRAWGID        int64              `bson:"game_rawg_id" json:"game_rawg_id"`
	Status            string             `bson:"status" json:"status"`
	HoursPlayed       *int               `bson:"hours_played" json:"hours_played"`
	Score             *float32           `bson:"score" json:"score"`
	AchievementStatus *float32           `bson:"achievement_status" json:"achievement_status"`
	TimesFinished     int                `bson:"times_finished" json:"times_finished"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
}

type GameMetacriticScorePlatform struct {
	Score    float64 `bson:"score" json:"score"`
	Platform string  `bson:"platform" json:"platform"`
}

type GameRelation struct {
	Name        string `bson:"name" json:"name"`
	ReleaseDate string `bson:"release_date" json:"release_date"`
	RawgID      int64  `bson:"rawg_id" json:"rawg_id"`
}

type GameDetailsRelation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	RawgID        int64              `bson:"rawg_id" json:"rawg_id"`
	Title         string             `bson:"title" json:"title"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ReleaseDate   string             `bson:"release_date" json:"release_date"`
	ImageUrl      string             `bson:"image_url" json:"image_url"`
	Platforms     []string           `bson:"platforms" json:"platforms"`
}

type GameStore struct {
	Url     string `bson:"url" json:"url"`
	StoreID int    `bson:"store_id" json:"store_id"`
}
