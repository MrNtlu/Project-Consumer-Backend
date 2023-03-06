package requests

type Login struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
}

type Register struct {
	EmailAddress string  `json:"email_address" binding:"required,email"`
	Username     string  `json:"username" binding:"required"`
	Password     string  `json:"password" binding:"required,min=6"`
	Image        *string `json:"image"`
	FCMToken     string  `bson:"fcm_token" json:"fcm_token"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangeFCMToken struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}

type ChangeNotification struct {
	AppNotification  *bool `json:"app_notification" binding:"required"`
	MailNotification *bool `json:"mail_notification" binding:"required"`
}

type ChangeMembership struct {
	IsPremium bool `json:"is_premium"`
}

type ForgotPassword struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
}
