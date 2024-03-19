package requests

type CreateRecommendation struct {
	ContentID        string  `json:"content_id" binding:"required"`
	RecommendationID string  `json:"recommendation_id" binding:"required"`
	ContentType      string  `json:"content_type" binding:"required,oneof=anime game movie tv"`
	Reason           *string `json:"reason" binding:"omitempty,min=6,max=1000"`
}

type SortRecommendation struct {
	Sort string `form:"sort" binding:"required,oneof=popularity latest oldest"`
	Page int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type SortRecommendationByUserID struct {
	UserID string `form:"user_id" binding:"required"`
	Sort   string `form:"sort" binding:"required,oneof=popularity latest oldest"`
	Page   int64  `form:"page" json:"page" binding:"required,number,min=1"`
}
