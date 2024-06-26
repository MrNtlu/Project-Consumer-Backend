package requests

// We have 2 external id's because Movie and TVSeries external ids are string but
// game and anime external ids are integers.
type CreateConsumeLater struct {
	ContentID            string  `json:"content_id" binding:"required"`
	ContentExternalID    *string `json:"content_external_id"`
	ContentExternalIntID *int64  `json:"content_external_int_id"`
	ContentType          string  `json:"content_type" binding:"required,oneof=anime game movie tv"`
	SelfNote             *string `json:"self_note"`
}

type SortFilterConsumeLater struct {
	ContentType       *string `form:"type" binding:"omitempty,oneof=anime game movie tv"`
	Genre             *string `form:"genre"`
	StreamingPlatform *string `form:"streaming_platform"`
	Sort              string  `form:"sort" binding:"required,oneof=new old alphabetical unalphabetical soon later"`
}

type UpdateConsumeLater struct {
	ID       string  `json:"id" binding:"required"`
	SelfNote *string `json:"self_note"`
}

type MarkConsumeLater struct {
	ID    string   `json:"id" binding:"required"`
	Score *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
}
