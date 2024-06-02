package requests

type SortFilterGame struct {
	Genres    *string `form:"genres"`
	Platform  *string `form:"platform"`
	TBA       *bool   `form:"tba"`
	Publisher *string `form:"publisher"`
	Rating    *int    `form:"rating" binding:"omitempty,number,min=1,max=5"`
	Sort      string  `form:"sort" binding:"required,oneof=top popularity new old"`
	Page      int64   `form:"page" json:"page" binding:"required,number,min=1"`
}
