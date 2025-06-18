package requests

type TMDBImportRequest struct {
	TMDBUsername string `json:"tmdb_username" binding:"required"`
}
