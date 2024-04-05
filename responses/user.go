package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IsUserPremium struct {
	IsPremium         bool `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
	MembershipType    int  `bson:"membership_type" json:"membership_type"`
}

type Leaderboard struct {
	Username  string `bson:"username" json:"username"`
	UserID    string `bson:"user_id" json:"user_id"`
	Level     int    `bson:"level" json:"level"`
	IsPremium bool   `bson:"is_premium" json:"is_premium"`
	Image     string `bson:"image" json:"image"`
}

type UserLevel struct {
	TotalScore int `bson:"total_score" json:"total_score"`
}

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username          string             `bson:"username" json:"username"`
	EmailAddress      string             `bson:"email" json:"email"`
	Image             string             `bson:"image" json:"image"`
	IsPremium         bool               `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool               `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
	MembershipType    int                `bson:"membership_type" json:"membership_type"` //0 Basic, 1 Premium 2 Premium Supporter
	IsOAuthUser       bool               `bson:"is_oauth" json:"is_oauth"`
	OAuthType         *int               `bson:"oauth_type" json:"oauth_type"` //0 google, 1 apple
	RefreshToken      *string            `bson:"refresh_token" json:"-"`
	FCMToken          string             `bson:"fcm_token" json:"fcm_token"`
	CanChangeUsername bool               `bson:"can_change_username" json:"can_change_username"`
	AppNotification   Notification       `bson:"app_notification" json:"app_notification"`
	Streak            int                `bson:"streak" json:"streak"`
	UserListCount     int64              `bson:"user_list_count" json:"user_list_count"`
	ConsumeLaterCount int64              `bson:"consume_later_count" json:"consume_later_count"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
}

type Notification struct {
	FriendRequest bool `bson:"friend_request" json:"friend_request"`
	ReviewLikes   bool `bson:"review_likes" json:"review_likes"`
}

type UserStats struct {
	AnimeCount           int   `bson:"anime_count" json:"anime_count"`
	GameCount            int   `bson:"game_count" json:"game_count"`
	MovieCount           int   `bson:"movie_count" json:"movie_count"`
	TVCount              int   `bson:"tv_count" json:"tv_count"`
	MovieWatchedTime     int64 `bson:"movie_watched_time" json:"movie_watched_time"`
	TVWatchedEpisodes    int64 `bson:"tv_watched_episodes" json:"tv_watched_episodes"`
	AnimeWatchedEpisodes int64 `bson:"anime_watched_episodes" json:"anime_watched_episodes"`
	GameTotalHoursPlayed int64 `bson:"game_total_hours_played" json:"game_total_hours_played"`
	MovieTotalScore      int64 `bson:"movie_total_score" json:"movie_total_score"`
	TVTotalScore         int64 `bson:"tv_total_score" json:"tv_total_score"`
	AnimeTotalScore      int64 `bson:"anime_total_score" json:"anime_total_score"`
	GameTotalScore       int64 `bson:"game_total_score" json:"game_total_score"`
}

type UserInfo struct {
	ID                      primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
	Username                string              `bson:"username" json:"username"`
	EmailAddress            string              `bson:"email" json:"email"`
	IsPremium               bool                `bson:"is_premium" json:"is_premium"`
	IsFriendRequestSent     bool                `bson:"is_friend_request_sent" json:"is_friend_request_sent"`
	IsFriendRequestReceived bool                `bson:"is_friend_request_received" json:"is_friend_request_received"`
	IsFriendsWith           bool                `bson:"is_friends_with" json:"is_friends_with"`
	FriendRequestCount      int64               `bson:"friend_request_count" json:"friend_request_count"`
	MembershipType          int                 `bson:"membership_type" json:"membership_type"`
	Image                   string              `bson:"image" json:"image"`
	FCMToken                string              `bson:"fcm_token" json:"fcm_token"`
	Level                   int                 `bson:"level" json:"level"`
	Streak                  int                 `bson:"streak" json:"streak"`
	MaxStreak               int                 `bson:"max_streak" json:"max_streak"`
	AnimeCount              int                 `bson:"anime_count" json:"anime_count"`
	GameCount               int                 `bson:"game_count" json:"game_count"`
	MovieCount              int                 `bson:"movie_count" json:"movie_count"`
	TVCount                 int                 `bson:"tv_count" json:"tv_count"`
	MovieWatchedTime        int64               `bson:"movie_watched_time" json:"movie_watched_time"`
	AnimeWatchedEpisodes    int64               `bson:"anime_watched_episodes" json:"anime_watched_episodes"`
	TVWatchedEpisodes       int64               `bson:"tv_watched_episodes" json:"tv_watched_episodes"`
	GameTotalHoursPlayed    int64               `bson:"game_total_hours_played" json:"game_total_hours_played"`
	MovieTotalScore         float64             `bson:"movie_total_score" json:"movie_total_score"`
	TVTotalScore            float64             `bson:"tv_total_score" json:"tv_total_score"`
	AnimeTotalScore         float64             `bson:"anime_total_score" json:"anime_total_score"`
	GameTotalScore          float64             `bson:"game_total_score" json:"game_total_score"`
	CreatedAt               time.Time           `bson:"created_at" json:"created_at"`
	LegendContent           []UserInfoContent   `bson:"legend_content" json:"legend_content"`
	ConsumeLater            []ConsumeLater      `bson:"consume_later" json:"consume_later"`
	Reviews                 []ReviewWithContent `bson:"reviews" json:"reviews"`
	CustomLists             []CustomList        `bson:"custom_lists" json:"custom_lists"`
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
