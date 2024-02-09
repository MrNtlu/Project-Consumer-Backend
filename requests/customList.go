package requests

type CreateCustomList struct {
	Name        string              `json:"name" binding:"required,min=1,max=250"`
	Description *string             `json:"description" binding:"omitempty"`
	IsPrivate   *bool               `json:"is_private" binding:"required"`
	Content     []CustomListContent `json:"content" binding:"required"`
}

type CustomListContent struct {
	Order                int     `json:"order" binding:"required"`
	ContentID            string  `json:"content_id" binding:"required"`
	ContentExternalID    *string `json:"content_external_id"`
	ContentExternalIntID *int64  `json:"content_external_int_id"`
	ContentType          string  `json:"content_type" binding:"required,oneof=anime game movie tv"`
}

type UpdateCustomList struct {
	ID          string              `json:"id" binding:"required"`
	Name        string              `json:"name" binding:"required,min=1,max=250"`
	Description *string             `json:"description" binding:"omitempty"`
	IsPrivate   *bool               `json:"is_private" binding:"required"`
	Content     []CustomListContent `json:"content" binding:"required"`
}

type AddToCustomList struct {
	ID      string                 `json:"id" binding:"required"`
	Content AddToCustomListContent `json:"content" binding:"required"`
}

type AddToCustomListContent struct {
	ContentID            string  `json:"content_id" binding:"required"`
	ContentExternalID    *string `json:"content_external_id"`
	ContentExternalIntID *int64  `json:"content_external_int_id"`
	ContentType          string  `json:"content_type" binding:"required,oneof=anime game movie tv"`
}

type ReorderCustomList struct {
	ID      string              `json:"id" binding:"required"`
	Content []CustomListContent `json:"content" binding:"required"`
}

type BulkDeleteCustomList struct {
	ID      string   `json:"id" binding:"required"`
	Content []string `json:"content" binding:"required"`
}

type SortCustomList struct {
	UserID string `form:"user_id" binding:"omitempty"`
	Sort   string `form:"sort" binding:"required,oneof=popularity latest oldest alphabetical unalphabetical"`
}

type SortLikeBookmarkCustomList struct {
	Sort string `form:"sort" binding:"required,oneof=popularity latest oldest alphabetical unalphabetical"`
}
