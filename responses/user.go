package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type IsUserPremium struct {
	IsPremium      bool `bson:"is_premium" json:"is_premium"`
	MembershipType int  `bson:"membership_type" json:"membership_type"`
}

type UserLevel struct {
	TotalScore int `bson:"total_score" json:"total_score"`
}

type UserStats struct {
	AnimeCount           int   `bson:"anime_count" json:"anime_count"`
	GameCount            int   `bson:"game_count" json:"game_count"`
	MovieCount           int   `bson:"movie_count" json:"movie_count"`
	TVCount              int   `bson:"tv_count" json:"tv_count"`
	MovieWatchedTime     int64 `bson:"movie_watched_time" json:"movie_watched_time"`
	AnimeWatchedEpisodes int64 `bson:"anime_watched_episodes" json:"anime_watched_episodes"`
	TVWatchedEpisodes    int64 `bson:"tv_watched_episodes" json:"tv_watched_episodes"`
	GameTotalHoursPlayed int64 `bson:"game_total_hours_played" json:"game_total_hours_played"`
}

type UserInfo struct {
	ID                   primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	Username             string              `bson:"username" json:"username"`
	EmailAddress         string              `bson:"email" json:"email"`
	IsPremium            bool                `bson:"is_premium" json:"is_premium"`
	MembershipType       int                 `bson:"membership_type" json:"membership_type"`
	Image                string              `bson:"image" json:"image"`
	FCMToken             string              `bson:"fcm_token" json:"fcm_token"`
	Level                int                 `bson:"level" json:"level"`
	AnimeCount           int                 `bson:"anime_count" json:"anime_count"`
	GameCount            int                 `bson:"game_count" json:"game_count"`
	MovieCount           int                 `bson:"movie_count" json:"movie_count"`
	TVCount              int                 `bson:"tv_count" json:"tv_count"`
	MovieWatchedTime     int64               `bson:"movie_watched_time" json:"movie_watched_time"`
	AnimeWatchedEpisodes int64               `bson:"anime_watched_episodes" json:"anime_watched_episodes"`
	TVWatchedEpisodes    int64               `bson:"tv_watched_episodes" json:"tv_watched_episodes"`
	GameTotalHoursPlayed int64               `bson:"game_total_hours_played" json:"game_total_hours_played"`
	LegendContent        []UserInfoContent   `bson:"legend_content" json:"legend_content"`
	ConsumeLater         []ConsumeLater      `bson:"consume_later" json:"consume_later"`
	Reviews              []ReviewWithContent `bson:"reviews" json:"reviews"`
}

type UserInfoContent struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	HoursPlayed   *int64             `bson:"hours_played" json:"hours_played"`
	ContentType   string             `bson:"content_type" json:"content_type"`
}
