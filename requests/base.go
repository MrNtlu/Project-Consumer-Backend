package requests

type Pagination struct {
	Page int64 `form:"page" binding:"required,number,min=1"`
}

type ID struct {
	ID string `json:"id" form:"id" binding:"required"`
}

type SortUpcoming struct {
	Sort string `form:"sort" binding:"required,oneof=popularity soon later"`
	Page int64  `form:"page" json:"page" binding:"required,number,min=1"`
}
