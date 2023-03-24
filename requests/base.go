package requests

type Pagination struct {
	Page int64 `form:"page" binding:"required,number,min=1"`
}

type ID struct {
	ID string `json:"id" form:"id" binding:"required"`
}
