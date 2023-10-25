package requests

type CreateReview struct {
	ContentID            string  `json:"content_id" binding:"required"`
	ContentExternalID    *string `json:"content_external_id"`
	ContentExternalIntID *int64  `json:"content_external_int_id"`
	ContentType          string  `json:"content_type" binding:"required,oneof=anime game movie tv"`
	Star                 int8    `json:"star" binding:"required,number,min=1,max=5"`
	Review               *string `json:"review" binding:"omitempty,min=6,max=1000"`
}

type UpdateReview struct {
	ID     string  `json:"id" binding:"required"`
	Review *string `json:"review" binding:"omitempty,min=6,max=1000"`
	Star   *int8   `json:"star" binding:"required,number,min=1,max=5"`
}

type VoteReview struct {
	ID     string `json:"id" binding:"required"`
	IsLike bool   `json:"is_like" binding:"required"`
}

type SortReviewByContentID struct {
	ContentID            string  `form:"content_id" binding:"required"`
	ContentExternalID    *string `form:"content_external_id"`
	ContentExternalIntID *int64  `form:"content_external_int_id"`
	Sort                 string  `form:"sort" binding:"required,oneof=popularity latest oldest"`
	Page                 int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
