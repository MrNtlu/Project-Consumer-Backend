package requests

type GoogleLogin struct {
	Token    string `json:"token" binding:"required"`
	FCMToken string `json:"fcm_token" binding:"required"`
}
