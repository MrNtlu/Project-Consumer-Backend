package requests

type SortFilterGame struct {
	Genres   *string `form:"genres"`
	Platform *string `form:"platform"`
	TBA      *bool   `form:"tba"`
	Sort     string  `form:"sort" binding:"required,oneof=popularity metacritic new old"`
	Page     int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
