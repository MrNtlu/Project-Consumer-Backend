package requests

type SortUpcomingAnime struct {
	Sort      string `form:"sort" binding:"required,oneof=popularity date"`
	SortOrder int8   `form:"type" json:"type" binding:"required,oneof=1 -1"`
	Page      int64  `form:"page" json:"page" binding:"required,number,min=1"`
}
