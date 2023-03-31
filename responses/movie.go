package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type Movie struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	TitleEn             string                 `bson:"title_en" json:"title_en"`
	TitleOriginal       string                 `bson:"title_original" json:"title_original"`
	Description         string                 `bson:"description" json:"description"`
	ImageURL            string                 `bson:"image_url" json:"image_url"`
	SmallImageURL       string                 `bson:"small_image_url" json:"small_image_url"`
	Status              string                 `bson:"status" json:"status"`
	Length              int                    `bson:"length" json:"length"`
	ImdbID              string                 `bson:"imdb_id" json:"imdb_id"`
	TmdbID              string                 `bson:"tmdb_id" json:"tmdb_id"`
	TmdbPopularity      float64                `bson:"tmdb_popularity" json:"tmdb_popularity"`
	TmdbVote            float64                `bson:"tmdb_vote" json:"tmdb_vote"`
	TmdbVoteCount       int64                  `bson:"tmdb_vote_count" json:"tmdb_vote_count"`
	ReleaseDate         string                 `bson:"release_date" json:"release_date"`
	ProductionCompanies []ProductionAndCompany `bson:"production_companies" json:"production_companies"`
	Genres              []Genre                `bson:"genres" json:"genres"`
	Streaming           []Streaming            `bson:"streaming" json:"streaming"`
}

type ProductionAndCompany struct {
	Logo          *string `bson:"logo" json:"logo"`
	Name          string  `bson:"name" json:"name"`
	OriginCountry string  `bson:"origin_country" json:"origin_country"`
}

type Genre struct {
	TmdbID string `bson:"tmdb_id" json:"tmdb_id"`
	Name   string `bson:"name" json:"name"`
}

type Streaming struct {
	CountryCode        string              `bson:"country_code" json:"country_code"`
	StreamingPlatforms []StreamingPlatform `bson:"streaming_platforms" json:"streaming_platforms"`
	BuyOptions         []StreamingPlatform `bson:"buy_options" json:"buy_options"`
	RentOptions        []StreamingPlatform `bson:"rent_options" json:"rent_options"`
}

type StreamingPlatform struct {
	Logo string `bson:"logo" json:"logo"`
	Name string `bson:"name" json:"name"`
}