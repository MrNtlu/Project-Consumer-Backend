package requests

type CreateReview struct {
	ContentID            string  `json:"content_id" binding:"required"`
	ContentExternalID    *string `json:"content_external_id"`
	ContentExternalIntID *int64  `json:"content_external_int_id"`
	Star                 int8    `json:"star" binding:"required,number,min=1,max=5"`
	Review               *string `json:"review" binding:"omitempty,len=1000"`
}
