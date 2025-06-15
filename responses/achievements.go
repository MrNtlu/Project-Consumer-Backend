package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Achievement struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title       string             `bson:"title" json:"title"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type UserAchievementResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title       string             `bson:"title" json:"title"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	Description string             `bson:"description" json:"description"`
	Unlocked    bool               `bson:"unlocked" json:"unlocked"`
	UnlockedAt  *time.Time         `bson:"unlocked_at,omitempty" json:"unlocked_at,omitempty"`
}
