package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

type IsUserPremium struct {
	IsPremium         bool `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
}

type UserLevel struct {
	TotalScore int `bson:"total_score" json:"total_score"`
}

type UserInfo struct {
	Username       string            `bson:"username" json:"username"`
	EmailAddress   string            `bson:"email" json:"email"`
	IsPremium      bool              `bson:"is_premium" json:"is_premium"`
	MembershipType int               `bson:"membership_type" json:"membership_type"`
	Image          string            `bson:"image" json:"image"`
	AnimeCount     int               `bson:"anime_count" json:"anime_count"`
	GameCount      int               `bson:"game_count" json:"game_count"`
	MovieCount     int               `bson:"movie_count" json:"movie_count"`
	TVCount        int               `bson:"tv_count" json:"tv_count"`
	FCMToken       string            `bson:"fcm_token" json:"fcm_token"`
	Level          int               `bson:"level" json:"level"`
	LegendContent  []UserInfoContent `bson:"legend_content" json:"legend_content"`
}

type UserInfoContent struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TitleEn       string             `bson:"title_en" json:"title_en"`
	TitleOriginal string             `bson:"title_original" json:"title_original"`
	ImageURL      string             `bson:"image_url" json:"image_url"`
	TimesFinished int                `bson:"times_finished" json:"times_finished"`
	ContentType   string             `bson:"content_type" json:"content_type"`
}
