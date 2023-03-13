package requests

type Pagination struct {
	Page int64 `form:"page" binding:"required,number,min=1"`
}
