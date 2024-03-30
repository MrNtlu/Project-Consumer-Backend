package requests

type Login struct {
	EmailAddress string  `json:"email_address" binding:"required,email"`
	Password     string  `json:"password" binding:"required"`
	FCMToken     *string `json:"fcm_token" binding:"omitempty"`
}

type Register struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	Username     string `json:"username" binding:"required,min=3"`
	Password     string `json:"password" binding:"required,min=6"`
	Image        string `json:"image"`
	FCMToken     string `json:"fcm_token"`
}

type SendFriendRequest struct {
	Username string `json:"username" binding:"required"`
}

type ChangeImage struct {
	Image string `json:"image" binding:"required"`
}

type ChangePassword struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangeFCMToken struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}

type ChangeNotification struct {
	Promotions  *bool `json:"promotions" binding:"required"`
	Updates     *bool `json:"updates" binding:"required"`
	Follows     *bool `json:"follows" binding:"required"`
	ReviewLikes *bool `json:"review_likes" binding:"required"`
}

type ChangeMembership struct {
	IsPremium      bool `json:"is_premium"`
	MembershipType int  `json:"membership_type"`
}

type ChangeUsername struct {
	Username string `json:"username" binding:"required,min=3"`
}

type ForgotPassword struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
}

type GetProfile struct {
	Username string `form:"username" binding:"required"`
}
