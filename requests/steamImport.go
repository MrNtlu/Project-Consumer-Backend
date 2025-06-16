package requests

type SteamImportRequest struct {
	SteamID       string `json:"steam_id,omitempty"`
	SteamUsername string `json:"steam_username,omitempty"`
}
