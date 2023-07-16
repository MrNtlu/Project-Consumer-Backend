package responses

import "time"

type LogsByRange struct {
	Date  string `bson:"date" json:"date"`
	Count int    `bson:"count" json:"count"`
	Log   []Log  `bson:"data" json:"data"`
}

type Log struct {
	LogType          string    `bson:"log_type" json:"log_type"`
	LogAction        string    `bson:"log_action" json:"log_action"`
	LogActionDetails string    `bson:"log_action_details" json:"log_action_details"`
	ContentTitle     string    `bson:"content_title" json:"content_title"`
	ContentImage     string    `bson:"content_image" json:"content_image"`
	ContentType      string    `bson:"content_type" json:"content_type"`
	ContentID        string    `bson:"content_id" json:"content_id"`
	CreatedAt        time.Time `bson:"created_at" json:"created_at"`
}
