package requests

type IMDBImportRequest struct {
	IMDBUserID string `json:"imdb_user_id" binding:"omitempty"`
	IMDBListID string `json:"imdb_list_id" binding:"omitempty"`
}
