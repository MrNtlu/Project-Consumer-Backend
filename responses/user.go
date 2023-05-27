package responses

type IsUserPremium struct {
	IsPremium         bool `bson:"is_premium" json:"is_premium"`
	IsLifetimePremium bool `bson:"is_lifetime_premium" json:"is_lifetime_premium"`
}

type UserInfo struct {
	IsPremium        bool   `bson:"is_premium" json:"is_premium"`
	IsOAuth          bool   `bson:"is_oauth" json:"is_oauth"`
	AppNotification  bool   `bson:"app_notification" json:"app_notification"`
	MailNotification bool   `bson:"mail_notification" json:"mail_notification"`
	EmailAddress     string `bson:"email" json:"email"`
	Image            string `bson:"image" json:"image"`
	Username         string `bson:"username" json:"username"`
	FCMToken         string `bson:"fcm_token" json:"fcm_token"`
}
