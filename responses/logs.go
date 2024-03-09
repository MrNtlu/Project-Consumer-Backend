package responses

import "time"

type LogDates struct {
	Dates []time.Time `bson:"dates" json:"dates"`
}

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

type FinishedLogStats struct {
	ContentType     string `bson:"content_type" json:"content_type"`
	Length          int64  `bson:"length" json:"length"`
	TotalEpisodes   int64  `bson:"total_episodes" json:"total_episodes"`
	TotalSeasons    int64  `bson:"total_seasons" json:"total_seasons"`
	MetacriticScore int64  `bson:"metacritic_score" json:"metacritic_score"`
	Count           int64  `bson:"count" json:"count"`
}

type MostLikedGenres struct {
	Type     string `bson:"type" json:"type"`
	Genre    string `bson:"genre" json:"genre"`
	MaxCount int64  `bson:"max_count" json:"max_count"`
}

type ChartLogs struct {
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Count     int64     `bson:"count" json:"count"`
	DayOfWeek int64     `bson:"day_of_week" json:"day_of_week"`
}

type ExtraStatistics struct {
	FinishedLogStats []FinishedLogStats `bson:"stats" json:"stats"`
	MostLikedGenres  []MostLikedGenres  `bson:"genres" json:"genres"`
	ChartLogs        []ChartLogs        `bson:"logs" json:"logs"`
}
