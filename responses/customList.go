package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomList struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	UserID      string              `bson:"user_id" json:"user_id"`
	Name        string              `bson:"name" json:"name"`
	Description *string             `bson:"description" json:"description"`
	Likes       []string            `bson:"likes" json:"likes"`
	IsPrivate   bool                `bson:"is_private" json:"is_private"`
	Content     []CustomListContent `bson:"content" json:"content"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
}

type CustomListContent struct {
	Order                int      `bson:"order" json:"order"`
	ContentID            string   `bson:"content_id" json:"content_id"`
	ContentExternalID    *string  `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64   `bson:"content_external_int_id" json:"content_external_int_id"`
	ContentType          string   `bson:"content_type" json:"content_type"`
	TitleEn              string   `bson:"title_en" json:"title_en"`
	TitleOriginal        string   `bson:"title_original" json:"title_original"`
	ImageURL             *string  `bson:"image_url" json:"image_url"`
	Score                *float64 `bson:"score" json:"score"`
	Description          string   `bson:"description" json:"description"`
}
