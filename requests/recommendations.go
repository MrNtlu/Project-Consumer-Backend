package requests

type CreateRecommendation struct {
	ContentID        string  `json:"content_id" binding:"required"`
	RecommendationID string  `json:"recommendation_id" binding:"required"`
	ContentType      string  `json:"content_type" binding:"required,oneof=anime game movie tv"`
	Reason           *string `json:"reason" binding:"omitempty,min=6,max=1000"`
}

type LikeRecommendation struct {
	RecommendationID string `json:"id" binding:"required"`
	ContentType      string `json:"type" binding:"required,oneof=anime game movie tv"`
}

type SortRecommendation struct {
	ContentID   string `form:"content_id" binding:"required"`
	ContentType string `form:"content_type" binding:"required,oneof=anime game movie tv"`
	Sort        string `form:"sort" binding:"required,oneof=popularity latest oldest"`
	Page        int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type SortRecommendationByUserID struct {
	UserID string `form:"user_id" binding:"required"`
	Sort   string `form:"sort" binding:"required,oneof=popularity latest oldest"`
	Page   int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type SortRecommendationsForSocial struct {
	Sort        string  `form:"sort" binding:"required,oneof=popularity latest oldest"`
	ContentType *string `form:"content_type" binding:"omitempty,oneof=anime game movie tv"`
	Page        int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
