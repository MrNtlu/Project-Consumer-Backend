package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID               string             `bson:"user_id" json:"user_id"`
	Author               ReviewUser         `bson:"author" json:"author"`
	IsAuthor             bool               `bson:"is_author" json:"is_author"`
	ContentID            string             `bson:"content_id" json:"content_id"`
	ContentExternalID    *string            `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64             `bson:"content_external_int_id" json:"content_external_int_id"`
	Star                 int8               `bson:"star" json:"star"`
	Review               *string            `bson:"review" json:"review"`
	Likes                []ReviewUser       `bson:"likes" json:"likes"`
	Dislikes             []ReviewUser       `bson:"dislikes" json:"dislikes"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

type ReviewUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Image        string             `bson:"image" json:"image"`
	Username     string             `bson:"username" json:"username"`
	EmailAddress string             `bson:"email" json:"email"`
	IsPremium    bool               `bson:"is_premium" json:"is_premium"`
}
