package requests

type NotInterestedRequest struct {
	ContentID   string `json:"content_id"`
	ContentType string `json:"content_type"`
	IsDelete    bool   `json:"is_delete"`
}
