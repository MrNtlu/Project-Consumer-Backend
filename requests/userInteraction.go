package requests

type CreateConsumeLater struct {
	UserID      string  `json:"user_id" binding:"required"`
	ContentID   string  `json:"content_id" binding:"required"`
	ContentType string  `json:"content_type" binding:"required,oneof=anime game movie tvseries"`
	SelfNote    *string `json:"self_note"`
}
