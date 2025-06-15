package requests

type MALImportRequest struct {
	Username string `json:"username" binding:"required" validate:"required"`
}
