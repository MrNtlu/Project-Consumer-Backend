package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PreviewMovie struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TmdbID        string             `bson:"tmdb_id" json:"tmdb_id"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
}

type Movie struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	TitleEn             string                 `bson:"title_en" json:"title_en"`
	TitleOriginal       string                 `bson:"title_original" json:"title_original"`
	Description         string                 `bson:"description" json:"description"`
	ImageURL            string                 `bson:"image_url" json:"image_url"`
	Status              string                 `bson:"status" json:"status"`
	Length              int                    `bson:"length" json:"length"`
	ImdbID              string                 `bson:"imdb_id" json:"imdb_id"`
	TmdbID              string                 `bson:"tmdb_id" json:"tmdb_id"`
	TmdbPopularity      float64                `bson:"tmdb_popularity" json:"tmdb_popularity"`
	TopRated            float64                `bson:"top_rated" json:"top_rated"`
	TmdbVote            float64                `bson:"tmdb_vote" json:"tmdb_vote"`
	TmdbVoteCount       int64                  `bson:"tmdb_vote_count" json:"tmdb_vote_count"`
	ReleaseDate         string                 `bson:"release_date" json:"release_date"`
	Backdrop            *string                `bson:"backdrop" json:"backdrop"`
	Recommendations     []Recommendation       `bson:"recommendations" json:"recommendations"`
	ProductionCompanies []ProductionAndCompany `bson:"production_companies" json:"production_companies"`
	Genres              []string               `bson:"genres" json:"genres"`
	Images              []string               `bson:"images" json:"images"`
	Videos              []Trailer              `bson:"videos" json:"videos"`
	Streaming           []Streaming            `bson:"streaming" json:"streaming"`
	Actors              []Actor                `bson:"actors" json:"actors"`
	Translations        []Translation          `bson:"translations" json:"translations"`
	Review              ReviewSummary          `bson:"reviews" json:"reviews"`
}

type MovieDetails struct {
	ID                  primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	TitleEn             string                 `bson:"title_en" json:"title_en"`
	TitleOriginal       string                 `bson:"title_original" json:"title_original"`
	Description         string                 `bson:"description" json:"description"`
	ImageURL            string                 `bson:"image_url" json:"image_url"`
	Status              string                 `bson:"status" json:"status"`
	Length              int                    `bson:"length" json:"length"`
	ImdbID              string                 `bson:"imdb_id" json:"imdb_id"`
	TmdbID              string                 `bson:"tmdb_id" json:"tmdb_id"`
	TmdbPopularity      float64                `bson:"tmdb_popularity" json:"tmdb_popularity"`
	TmdbVote            float64                `bson:"tmdb_vote" json:"tmdb_vote"`
	TmdbVoteCount       int64                  `bson:"tmdb_vote_count" json:"tmdb_vote_count"`
	ReleaseDate         string                 `bson:"release_date" json:"release_date"`
	Backdrop            *string                `bson:"backdrop" json:"backdrop"`
	Recommendations     []Recommendation       `bson:"recommendations" json:"recommendations"`
	ProductionCompanies []ProductionAndCompany `bson:"production_companies" json:"production_companies"`
	Genres              []string               `bson:"genres" json:"genres"`
	Images              []string               `bson:"images" json:"images"`
	Videos              []Trailer              `bson:"videos" json:"videos"`
	Streaming           []Streaming            `bson:"streaming" json:"streaming"`
	Actors              []Actor                `bson:"actors" json:"actors"`
	Translations        []Translation          `bson:"translations" json:"translations"`
	WatchList           *MovieDetailsWatchList `bson:"watch_list" json:"watch_list"`
	WatchLater          *ConsumeLater          `bson:"watch_later" json:"watch_later"`
	Review              ReviewSummary          `bson:"reviews" json:"reviews"`
}

type MovieDetailsWatchList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	MovieID       string             `bson:"movie_id" json:"movie_id"`
	MovieTmdbID   string             `bson:"movie_tmdb_id" json:"movie_tmdb_id"`
	Status        string             `bson:"status" json:"status"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

type ProductionAndCompany struct {
	Logo          *string `bson:"logo" json:"logo"`
	Name          string  `bson:"name" json:"name"`
	OriginCountry string  `bson:"origin_country" json:"origin_country"`
}

type Recommendation struct {
	TmdbID        string `bson:"tmdb_id" json:"tmdb_id"`
	TitleEn       string `bson:"title_en" json:"title_en"`
	TitleOriginal string `bson:"title_original" json:"title_original"`
	ReleaseDate   string `bson:"release_date" json:"release_date"`
	Description   string `bson:"description" json:"description"`
	ImageURL      string `bson:"image_url" json:"image_url"`
}

type Actor struct {
	TmdbID    string  `bson:"tmdb_id" json:"tmdb_id"`
	Image     *string `bson:"image" json:"image"`
	Name      string  `bson:"name" json:"name"`
	Character string  `bson:"character" json:"character"`
}

type Trailer struct {
	Name string `bson:"name" json:"name"`
	Key  string `bson:"key" json:"key"`
	Type string `bson:"type" json:"title_original"`
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

type Translation struct {
	LanCode     string `bson:"lan_code" json:"lan_code"`
	LanName     string `bson:"lan_name" json:"lan_name"`
	LanNameEn   string `bson:"lan_name_en" json:"lan_name_en"`
	Title       string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
}
