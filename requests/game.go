package requests

type SortGame struct {
	Sort string `form:"sort" binding:"required,oneof=popularity new old"`
	Page int64  `form:"page" json:"page" binding:"required,number,min=1"`
}

type SortFilterGame struct {
	Genres   *string `form:"genres"`
	Platform *string `form:"platform"`
	TBA      *bool   `form:"tba"`
	Sort     string  `form:"sort" binding:"required,oneof=popularity new old"`
	Page     int64   `form:"page" json:"page" binding:"required,number,min=1"`
}