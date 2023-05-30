package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type ConsumeLater struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID            string             `bson:"user_id" json:"user_id"`
	ContentID         string             `bson:"content_id" json:"content_id"`
	ContentExternalID *string            `bson:"content_external_id" json:"content_external_id"`
	ContentType       string             `bson:"content_type" json:"content_type"`
	SelfNote          *string            `bson:"self_note" json:"self_note"`
}
