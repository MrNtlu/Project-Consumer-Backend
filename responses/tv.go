package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PreviewTVSeries struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TmdbID        string             `bson:"tmdb_id" json:"tmdb_id"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
}

type DayOfWeekTVSeries struct {
	Data      []TVSeries `bson:"data" json:"data"`
	DayOfWeek int16      `bson:"day_of_week" json:"day_of_week"`
}

type TVSeries struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	TitleEn             string                 `bson:"title_en" json:"title_en"`
	TitleOriginal       string                 `bson:"title_original" json:"title_original"`
	Description         string                 `bson:"description" json:"description"`
	ImageURL            string                 `bson:"image_url" json:"image_url"`
	Status              string                 `bson:"status" json:"status"`
	TmdbID              string                 `bson:"tmdb_id" json:"tmdb_id"`
	TmdbPopularity      float64                `bson:"tmdb_popularity" json:"tmdb_popularity"`
	TopRated            float64                `bson:"top_rated" json:"top_rated"`
	TmdbVote            float64                `bson:"tmdb_vote" json:"tmdb_vote"`
	TmdbVoteCount       int64                  `bson:"tmdb_vote_count" json:"tmdb_vote_count"`
	ProductionCompanies []ProductionAndCompany `bson:"production_companies" json:"production_companies"`
	TotalSeasons        int                    `bson:"total_seasons" json:"total_seasons"`
	TotalEpisodes       int                    `bson:"total_episodes" json:"total_episodes"`
	FirstAirDate        string                 `bson:"first_air_date" json:"first_air_date"`
	Backdrop            *string                `bson:"backdrop" json:"backdrop"`
	Recommendations     []Recommendation       `bson:"recommendations" json:"recommendations"`
	Genres              []string               `bson:"genres" json:"genres"`
	Images              []string               `bson:"images" json:"images"`
	Videos              []Trailer              `bson:"videos" json:"videos"`
	Streaming           []Streaming            `bson:"streaming" json:"streaming"`
	Seasons             []Season               `bson:"seasons" json:"seasons"`
	Networks            []ProductionAndCompany `bson:"networks" json:"networks"`
	Actors              []Actor                `bson:"actors" json:"actors"`
	Review              ReviewSummary          `bson:"reviews" json:"reviews"`
}

type TVSeriesDetails struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	TitleEn             string                 `bson:"title_en" json:"title_en"`
	TitleOriginal       string                 `bson:"title_original" json:"title_original"`
	Description         string                 `bson:"description" json:"description"`
	ImageURL            string                 `bson:"image_url" json:"image_url"`
	Status              string                 `bson:"status" json:"status"`
	TmdbID              string                 `bson:"tmdb_id" json:"tmdb_id"`
	TmdbPopularity      float64                `bson:"tmdb_popularity" json:"tmdb_popularity"`
	TmdbVote            float64                `bson:"tmdb_vote" json:"tmdb_vote"`
	TmdbVoteCount       int64                  `bson:"tmdb_vote_count" json:"tmdb_vote_count"`
	TotalSeasons        int                    `bson:"total_seasons" json:"total_seasons"`
	TotalEpisodes       int                    `bson:"total_episodes" json:"total_episodes"`
	FirstAirDate        string                 `bson:"first_air_date" json:"first_air_date"`
	Backdrop            *string                `bson:"backdrop" json:"backdrop"`
	Genres              []string               `bson:"genres" json:"genres"`
	Images              []string               `bson:"images" json:"images"`
	Videos              []Trailer              `bson:"videos" json:"videos"`
	Recommendations     []Recommendation       `bson:"recommendations" json:"recommendations"`
	Streaming           []Streaming            `bson:"streaming" json:"streaming"`
	Seasons             []Season               `bson:"seasons" json:"seasons"`
	Networks            []ProductionAndCompany `bson:"networks" json:"networks"`
	ProductionCompanies []ProductionAndCompany `bson:"production_companies" json:"production_companies"`
	Actors              []Actor                `bson:"actors" json:"actors"`
	TVList              *TVDetailsList         `bson:"tv_list" json:"tv_list"`
	WatchLater          *ConsumeLater          `bson:"watch_later" json:"watch_later"`
	Review              ReviewSummary          `bson:"reviews" json:"reviews"`
}

type TVDetailsList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	TvID            string             `bson:"tv_id" json:"tv_id"`
	TvTmdbID        string             `bson:"tv_tmdb_id" json:"tv_tmdb_id"`
	Status          string             `bson:"status" json:"status"`
	Score           *float32           `bson:"score" json:"score"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	WatchedSeasons  int                `bson:"watched_seasons" json:"watched_seasons"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

type Season struct {
	AirDate      string `bson:"air_date" json:"air_date"`
	EpisodeCount int    `bson:"episode_count" json:"episode_count"`
	Name         string `bson:"name" json:"name"`
	SeasonNum    int    `bson:"season_num" json:"season_num"`
	ImageURL     string `bson:"image_url" json:"image_url"`
}
