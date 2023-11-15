package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendRequest struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Sender    Author             `bson:"sender" json:"sender"`
	Receiver  Author             `bson:"receiver" json:"receiver"`
	IsIgnored bool               `bson:"is_ignored" json:"is_ignored"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
