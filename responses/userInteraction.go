package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConsumeLater struct {
	ID                   primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	UserID               string              `bson:"user_id" json:"user_id"`
	ContentID            string              `bson:"content_id" json:"content_id"`
	ContentExternalID    *string             `bson:"content_external_id" json:"content_external_id"`
	ContentExternalIntID *int64              `bson:"content_external_int_id" json:"content_external_int_id"`
	ContentType          string              `bson:"content_type" json:"content_type"`
	SelfNote             *string             `bson:"self_note" json:"self_note"`
	CreatedAt            time.Time           `bson:"created_at" json:"created_at"`
	Content              ConsumeLaterContent `bson:"content" json:"content"`
}

type ConsumeLaterContent struct {
	TitleOriginal string  `bson:"title_original" json:"title_original"`
	ImageURL      *string `bson:"image_url" json:"image_url"`
	Description   string  `bson:"description" json:"description"`
}
