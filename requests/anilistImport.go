package requests

type AniListImportRequest struct {
	AniListUsername string `json:"anilist_username" binding:"required"`
}
