package requests

type CreateLog struct {
	LogType          string `json:"log_type" binding:"required,oneof=userlist later"`
	LogAction        string `json:"log_action" binding:"required,oneof=add update delete"`
	LogActionDetails string `json:"log_action_details" binding:"required"`
	ContentTitle     string `json:"content_title" binding:"required"`
	ContentImage     string `json:"content_image"`
	ContentType      string `json:"content_type" binding:"required,oneof=anime game movie tv"`
	ContentID        string `json:"content_id" binding:"required"`
}
