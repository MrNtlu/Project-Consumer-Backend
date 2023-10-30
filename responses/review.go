package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID               string             `bson:"user_id" json:"user_id"`
	Author               Author             `bson:"author" json:"author"`
	IsAuthor             bool               `bson:"is_author" json:"is_author"`
	ContentID            string             `bson:"content_id" json:"content_id"`
	ContentExternalID    *string            `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64             `bson:"content_external_int_id" json:"content_external_int_id"`
	Star                 int8               `bson:"star" json:"star"`
	Review               *string            `bson:"review" json:"review"`
	Popularity           int64              `bson:"popularity" json:"popularity"`
	IsLiked              bool               `bson:"is_liked" json:"is_liked"`
	Likes                []string           `bson:"likes" json:"likes"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

type Author struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Image        string             `bson:"image" json:"image"`
	Username     string             `bson:"username" json:"username"`
	EmailAddress string             `bson:"email" json:"email"`
	IsPremium    bool               `bson:"is_premium" json:"is_premium"`
}

type ReviewSummary struct {
	AverageStar float64    `bson:"avg_star" json:"avg_star"`
	TotalVotes  int32      `bson:"total_votes" json:"total_votes"`
	IsReviewed  bool       `bson:"is_reviewed" json:"is_reviewed"`
	StarCounts  StarCounts `bson:"star_counts" json:"star_counts"`
}

type StarCounts struct {
	OneStar   int16 `bson:"one_star" json:"one_star"`
	TwoStar   int16 `bson:"two_star" json:"two_star"`
	ThreeStar int16 `bson:"three_star" json:"three_star"`
	FourStar  int16 `bson:"four_star" json:"four_star"`
	FiveStar  int16 `bson:"five_star" json:"five_star"`
}
