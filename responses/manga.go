package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PreviewManga struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleJP       string             `bson:"title_jp" json:"title_jp"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
	MalID         int64              `bson:"mal_id" json:"mal_id"`
}

type Manga struct {
	ID              primitive.ObjectID    `bson:"_id,omitempty" json:"_id"`
	TitleOriginal   string                `bson:"title_original" json:"title_original"`
	TitleEn         string                `bson:"title_en" json:"title_en"`
	TitleJP         string                `bson:"title_jp" json:"title_jp"`
	Description     string                `bson:"description" json:"description"`
	ImageURL        string                `bson:"image_url" json:"image_url"`
	MalID           int64                 `bson:"mal_id" json:"mal_id"`
	MalScore        float64               `bson:"mal_score" json:"mal_score"`
	MalScoredBy     int64                 `bson:"mal_scored_by" json:"mal_scored_by"`
	Type            string                `bson:"type" json:"type"`
	Chapters        *int64                `bson:"chapters" json:"chapters"`
	Volumes         *int64                `bson:"volumes" json:"volumes"`
	Status          string                `bson:"status" json:"status"`
	IsPublishing    bool                  `bson:"is_publishing" json:"is_publishing"`
	Published       AnimeAirDate          `bson:"published" json:"published"`
	Recommendations []AnimeRecommendation `bson:"recommendations" json:"recommendations"`
	Serializations  []AnimeNameURL        `bson:"serializations" json:"serializations"`
	Genres          []AnimeGenre          `bson:"genres" json:"genres"`
	Themes          []AnimeGenre          `bson:"themes" json:"themes"`
	Demographics    []AnimeGenre          `bson:"demographics" json:"demographics"`
	Relations       []AnimeRelation       `bson:"relations" json:"relations"`
	Characters      []AnimeCharacter      `bson:"characters" json:"characters"`
	Review          ReviewSummary         `bson:"reviews" json:"reviews"`
}

type MangaDetails struct {
	ID               primitive.ObjectID    `bson:"_id,omitempty" json:"_id"`
	TitleOriginal    string                `bson:"title_original" json:"title_original"`
	TitleEn          string                `bson:"title_en" json:"title_en"`
	TitleJP          string                `bson:"title_jp" json:"title_jp"`
	Description      string                `bson:"description" json:"description"`
	DescriptionExtra string                `bson:"description_extra" json:"description_extra"`
	ImageURL         string                `bson:"image_url" json:"image_url"`
	MalID            int64                 `bson:"mal_id" json:"mal_id"`
	MalScore         float64               `bson:"mal_score" json:"mal_score"`
	MalScoredBy      int64                 `bson:"mal_scored_by" json:"mal_scored_by"`
	Type             string                `bson:"type" json:"type"`
	Chapters         *int64                `bson:"chapters" json:"chapters"`
	Volumes          *int64                `bson:"volumes" json:"volumes"`
	Status           string                `bson:"status" json:"status"`
	IsPublishing     bool                  `bson:"is_publishing" json:"is_publishing"`
	Published        AnimeAirDate          `bson:"published" json:"published"`
	Recommendations  []AnimeRecommendation `bson:"recommendations" json:"recommendations"`
	Serializations   []AnimeNameURL        `bson:"serializations" json:"serializations"`
	Genres           []AnimeGenre          `bson:"genres" json:"genres"`
	Themes           []AnimeGenre          `bson:"themes" json:"themes"`
	Demographics     []AnimeGenre          `bson:"demographics" json:"demographics"`
	Relations        []AnimeRelation       `bson:"relations" json:"relations"`
	Characters       []AnimeCharacter      `bson:"characters" json:"characters"`
	MangaReadList    *MangaReadList        `bson:"manga_list" json:"manga_list"`
	WatchLater       *ConsumeLater         `bson:"watch_later" json:"watch_later"`
	Review           ReviewSummary         `bson:"reviews" json:"reviews"`
}

type MangaReadList struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	MangaID       string             `bson:"manga_id" json:"manga_id"`
	MangaMALID    int64              `bson:"manga_mal_id" json:"manga_mal_id"`
	Status        string             `bson:"status" json:"status"`
	ReadChapters  int                `bson:"read_chapters" json:"read_chapters"`
	ReadVolumes   int                `bson:"read_volumes" json:"read_volumes"`
	Score         *float32           `bson:"score" json:"score"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}
