package requests

type SortFilterManga struct {
	Status       *string `form:"status" binding:"omitempty,oneof=publishing discontinued finished"`
	Genres       *string `form:"genres"`
	Demographics *string `form:"demographics"`
	Themes       *string `form:"themes"`
	Sort         string  `form:"sort" binding:"required,oneof=top popularity new old"`
	Page         int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
