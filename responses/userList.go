package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type AnimeList struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	AnimeID         string             `bson:"anime_id" json:"anime_id"`
	Status          string             `bson:"status" json:"status"`
	WatchedEpisodes int                `bson:"watched_episodes" json:"watched_episodes"`
	Score           *float32           `bson:"score" json:"score"`
	TimesFinished   int                `bson:"times_finished" json:"times_finished"`
	Anime           Anime              `bson:"anime" json:"anime"`
}
