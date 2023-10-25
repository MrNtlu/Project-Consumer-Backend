package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DayOfWeekAnime struct {
	Data      []Anime `bson:"data" json:"data"`
	DayOfWeek int16   `bson:"day_of_week" json:"day_of_week"`
}

type Anime struct {
	ID              primitive.ObjectID    `bson:"_id,omitempty" json:"_id"`
	TitleOriginal   string                `bson:"title_original" json:"title_original"`
	TitleEn         string                `bson:"title_en" json:"title_en"`
	TitleJP         string                `bson:"title_jp" json:"title_jp"`
	Description     string                `bson:"description" json:"description"`
	ImageURL        string                `bson:"image_url" json:"image_url"`
	MalID           int64                 `bson:"mal_id" json:"mal_id"`
	Popularity      float64               `bson:"popularity" json:"popularity"`
	MalScore        float64               `bson:"mal_score" json:"mal_score"`
	MalScoredBy     int64                 `bson:"mal_scored_by" json:"mal_scored_by"`
	Trailer         *string               `bson:"trailer" json:"trailer"`
	Type            string                `bson:"type" json:"type"`
	Source          string                `bson:"source" json:"source"`
	Episodes        *int64                `bson:"episodes" json:"episodes"`
	Season          *string               `bson:"season" json:"season"`
	Year            *int                  `bson:"year" json:"year"`
	Status          string                `bson:"status" json:"status"`
	IsAiring        bool                  `bson:"is_airing" json:"is_airing"`
	Recommendations []AnimeRecommendation `bson:"recommendations" json:"recommendations"`
	Streaming       []AnimeNameURL        `bson:"streaming" json:"streaming"`
	Aired           AnimeAirDate          `bson:"aired" json:"aired"`
	AgeRating       *string               `bson:"age_rating" json:"age_rating"`
	Producers       []AnimeNameURL        `bson:"producers" json:"producers"`
	Studios         []AnimeNameURL        `bson:"studios" json:"studios"`
	Genres          []AnimeGenre          `bson:"genres" json:"genres"`
	Themes          []AnimeGenre          `bson:"themes" json:"themes"`
	Demographics    []AnimeGenre          `bson:"demographics" json:"demographics"`
	Relations       []AnimeRelation       `bson:"relations" json:"relations"`
	Characters      []AnimeCharacter      `bson:"characters" json:"characters"`
	Review          ReviewSummary         `bson:"reviews" json:"reviews"`
}

type AnimeDetails struct {
	ID              primitive.ObjectID    `bson:"_id,omitempty" json:"_id"`
	TitleOriginal   string                `bson:"title_original" json:"title_original"`
	TitleEn         string                `bson:"title_en" json:"title_en"`
	TitleJP         string                `bson:"title_jp" json:"title_jp"`
	Description     string                `bson:"description" json:"description"`
	ImageURL        string                `bson:"image_url" json:"image_url"`
	MalID           int64                 `bson:"mal_id" json:"mal_id"`
	MalScore        float64               `bson:"mal_score" json:"mal_score"`
	MalScoredBy     int64                 `bson:"mal_scored_by" json:"mal_scored_by"`
	Trailer         *string               `bson:"trailer" json:"trailer"`
	Type            string                `bson:"type" json:"type"`
	Source          string                `bson:"source" json:"source"`
	Episodes        *int64                `bson:"episodes" json:"episodes"`
	Season          *string               `bson:"season" json:"season"`
	Year            *int                  `bson:"year" json:"year"`
	Status          string                `bson:"status" json:"status"`
	IsAiring        bool                  `bson:"is_airing" json:"is_airing"`
	AgeRating       *string               `bson:"age_rating" json:"age_rating"`
	Aired           AnimeAirDate          `bson:"aired" json:"aired"`
	Recommendations []AnimeRecommendation `bson:"recommendations" json:"recommendations"`
	Streaming       []AnimeNameURL        `bson:"streaming" json:"streaming"`
	Producers       []AnimeNameURL        `bson:"producers" json:"producers"`
	Studios         []AnimeNameURL        `bson:"studios" json:"studios"`
	Genres          []AnimeGenre          `bson:"genres" json:"genres"`
	Themes          []AnimeGenre          `bson:"themes" json:"themes"`
	Demographics    []AnimeGenre          `bson:"demographics" json:"demographics"`
	Relations       []AnimeRelation       `bson:"relations" json:"relations"`
	Characters      []AnimeCharacter      `bson:"characters" json:"characters"`
	AnimeList       *AnimeWatchList       `bson:"anime_list" json:"anime_list"`
	WatchLater      *ConsumeLater         `bson:"watch_later" json:"watch_later"`
	Review          ReviewSummary         `bson:"reviews" json:"reviews"`
}

type AnimeWatchList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	AnimeMALID      int64              `bson:"anime_mal_id" json:"anime_mal_id"`
	Status          string             `bson:"status" json:"status"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	Score           *float32           `bson:"score" json:"score"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

type AnimeNameURL struct {
	Name string `bson:"name" json:"name"`
	Url  string `bson:"url" json:"url"`
}

type AnimeAirDate struct {
	From      string `bson:"from" json:"from"`
	To        string `bson:"to" json:"to"`
	FromDay   int    `bson:"from_day" json:"from_day"`
	FromMonth int    `bson:"from_month" json:"from_month"`
	FromYear  int    `bson:"from_year" json:"from_year"`
	ToDay     int    `bson:"to_day" json:"to_day"`
	ToMonth   int    `bson:"to_month" json:"to_month"`
	ToYear    int    `bson:"to_year" json:"to_year"`
}

type AnimeRecommendation struct {
	MalID    int64  `bson:"mal_id" json:"mal_id"`
	Title    string `bson:"title" json:"title"`
	ImageURL string `bson:"image_url" json:"image_url"`
}

type AnimeCharacter struct {
	Name  string `bson:"name" json:"name"`
	Role  string `bson:"role" json:"role"`
	Image string `bson:"image" json:"image"`
	MalID int64  `bson:"mal_id" json:"mal_id"`
}

type AnimeGenre struct {
	Name string `bson:"name" json:"name"`
	Url  string `bson:"url" json:"url"`
}

type AnimeRelation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	MalID         int64              `bson:"mal_id" json:"mal_id"`
	AnimeID       string             `bson:"anime_id" json:"anime_id"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	Relation      string             `bson:"relation" json:"relation"`
	Type          string             `bson:"type" json:"type"`
}

type OldAnimeRelation struct {
	Relation string                 `bson:"relation" json:"relation"`
	Source   []AnimeRelationDetails `bson:"source" json:"source"`
}

type AnimeRelationDetails struct {
	Name        string `bson:"name" json:"name"`
	Type        string `bson:"type" json:"type"`
	MalID       int64  `bson:"mal_id" json:"mal_id"`
	RedirectURL string `bson:"redirect_url" json:"redirect_url"`
}
