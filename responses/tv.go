package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type TVSeries struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	TitleEn             string                 `bson:"title_en" json:"title_en"`
	TitleOriginal       string                 `bson:"title_original" json:"title_original"`
	Description         string                 `bson:"description" json:"description"`
	ImageURL            string                 `bson:"image_url" json:"image_url"`
	SmallImageURL       string                 `bson:"small_image_url" json:"small_image_url"`
	Status              string                 `bson:"status" json:"status"`
	TmdbID              int                    `bson:"tmdb_id" json:"tmdb_id"`
	TmdbPopularity      float64                `bson:"tmdb_popularity" json:"tmdb_popularity"`
	TmdbVote            float64                `bson:"tmdb_vote" json:"tmdb_vote"`
	TmdbVoteCount       int64                  `bson:"tmdb_vote_count" json:"tmdb_vote_count"`
	ProductionCompanies []ProductionAndCompany `bson:"production_companies" json:"production_companies"`
	TotalSeasons        int                    `bson:"total_seasons" json:"total_seasons"`
	TotalEpisodes       int                    `bson:"total_episodes" json:"total_episodes"`
	FirstAirDate        string                 `bson:"first_air_date" json:"first_air_date"`
	Genres              []Genre                `bson:"genres" json:"genres"`
	Streaming           []Streaming            `bson:"streaming" json:"streaming"`
	Seasons             []Season               `bson:"seasons" json:"seasons"`
	Networks            []ProductionAndCompany `bson:"networks" json:"networks"`
}

type Season struct {
	AirDate      string `bson:"air_date" json:"air_date"`
	EpisodeCount int    `bson:"episode_count" json:"episode_count"`
	Name         string `bson:"name" json:"name"`
	Description  string `bson:"description" json:"description"`
	SeasonNum    int    `bson:"season_num" json:"season_num"`
	ImageURL     string `bson:"image_url" json:"image_url"`
}
