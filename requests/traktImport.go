package requests

type TraktImportRequest struct {
	TraktUsername string `json:"trakt_username" binding:"required"`
}
