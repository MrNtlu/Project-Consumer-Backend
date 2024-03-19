package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecommendationWithContent struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID                string             `bson:"user_id" json:"user_id"`
	Author                Author             `bson:"author" json:"author"`
	IsAuthor              bool               `bson:"is_author" json:"is_author"`
	ContentID             string             `bson:"content_id" json:"content_id"`
	ContentType           string             `bson:"content_type" json:"content_type"` // anime, movie, tv or game
	RecommendationID      string             `bson:"recommendation_id" json:"recommendation_id"`
	Reason                *string            `bson:"reason" json:"reason"`
	Popularity            int64              `bson:"popularity" json:"popularity"`
	Likes                 []string           `bson:"likes" json:"likes"`
	IsLiked               bool               `bson:"is_liked" json:"is_liked"`
	Content               ReviewContent      `bson:"content" json:"content"`
	RecommendationContent ReviewContent      `bson:"recommendation_content" json:"recommendation_content"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
}
