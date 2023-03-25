package responses

type CurrentlyAiringAnimeResponse struct {
	ID   int     `bson:"_id" json:"_id"`
	Data []Anime `bson:"data" json:"data"`
}

type Anime struct {
	TitleOriginal string          `bson:"title_original" json:"title_original"`
	TitleEn       string          `bson:"title_en" json:"title_en"`
	TitleJP       string          `bson:"title_jp" json:"title_jp"`
	Description   string          `bson:"description" json:"description"`
	ImageURL      string          `bson:"image_url" json:"image_url"`
	SmallImageURL string          `bson:"small_image_url" json:"small_image_url"`
	MalID         int64           `bson:"mal_id" json:"mal_id"`
	MalScore      float64         `bson:"mal_score" json:"mal_score"`
	MalScoredBy   int64           `bson:"mal_scored_by" json:"mal_scored_by"`
	Trailer       *string         `bson:"trailer" json:"trailer"`
	Type          string          `bson:"type" json:"type"`
	Source        string          `bson:"source" json:"source"`
	Episodes      *int64          `bson:"episodes" json:"episodes"`
	Season        *string         `bson:"season" json:"season"`
	Year          *int            `bson:"year" json:"year"`
	Status        string          `bson:"status" json:"status"`
	IsAiring      bool            `bson:"is_airing" json:"is_airing"`
	Streaming     []AnimeNameURL  `bson:"streaming" json:"streaming"`
	Aired         AnimeAirDate    `bson:"aired" json:"aired"`
	AgeRating     *string         `bson:"age_rating" json:"age_rating"`
	Producers     []AnimeNameURL  `bson:"producers" json:"producers"`
	Studios       []AnimeNameURL  `bson:"studios" json:"studios"`
	Genres        []AnimeGenre    `bson:"genres" json:"genres"`
	Themes        []AnimeGenre    `bson:"themes" json:"themes"`
	Demographics  []AnimeGenre    `bson:"demographics" json:"demographics"`
	Relations     []AnimeRelation `bson:"relations" json:"relations"`
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

type AnimeGenre struct {
	Name  string `bson:"name" json:"name"`
	Url   string `bson:"url" json:"url"`
	MalID int64  `bson:"mal_id" json:"mal_id"`
}

type AnimeRelation struct {
	Relation string                 `bson:"relation" json:"relation"`
	Source   []AnimeRelationDetails `bson:"source" json:"source"`
}

type AnimeRelationDetails struct {
	Name        string `bson:"name" json:"name"`
	Type        string `bson:"type" json:"type"`
	MalID       int64  `bson:"mal_id" json:"mal_id"`
	RedirectURL string `bson:"redirect_url" json:"redirect_url"`
}
