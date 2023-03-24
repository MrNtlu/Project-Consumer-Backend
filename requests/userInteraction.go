package requests

type CreateConsumeLater struct {
	ContentID   string  `json:"content_id" binding:"required"`
	ContentType string  `json:"content_type" binding:"required,oneof=anime game movie tvseries"`
	SelfNote    *string `json:"self_note"`
}

type UpdateConsumeLater struct {
	ID       string  `json:"id" binding:"required"`
	SelfNote *string `json:"self_note"`
}
