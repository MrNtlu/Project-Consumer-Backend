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

type LogsByDateRange struct {
	From string  `form:"from" binding:"required" time_format:"2006-01-02"`
	To   string  `form:"to" binding:"required" time_format:"2006-01-02"`
	Sort *string `form:"sort" binding:"omitempty,oneof=asc desc"`
}

type LogStatInterval struct {
	Interval string `form:"interval" binding:"required,oneof=weekly monthly 3months"`
}
