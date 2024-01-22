package requests

type AssistantRequest struct {
	ContentName string `form:"content_name" binding:"required"`
	ContentType string `form:"content_type" binding:"required,oneof=anime game movie tv"`
}
