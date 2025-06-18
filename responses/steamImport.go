package responses

type SteamImportResponse struct {
	ImportedCount  int      `json:"imported_count"`
	SkippedCount   int      `json:"skipped_count"`
	ErrorCount     int      `json:"error_count"`
	Message        string   `json:"message"`
	ImportedTitles []string `json:"imported_titles"`
	SkippedTitles  []string `json:"skipped_titles"`
}

type SteamGameEntry struct {
	AppID                    int    `json:"appid"`
	Name                     string `json:"name"`
	PlaytimeForever          int    `json:"playtime_forever"`         // Total playtime in minutes
	PlaytimeWindowsForever   int    `json:"playtime_windows_forever"` // Windows playtime in minutes
	PlaytimeMacForever       int    `json:"playtime_mac_forever"`     // Mac playtime in minutes
	PlaytimeLinuxForever     int    `json:"playtime_linux_forever"`   // Linux playtime in minutes
	PlaytimeDisconnected     int    `json:"playtime_disconnected"`    // Offline playtime in minutes
	RtimeLastPlayed          int64  `json:"rtime_last_played"`        // Unix timestamp of last play
	PlaytimeDeckForever      int    `json:"playtime_deck_forever"`    // Steam Deck playtime in minutes
	HasCommunityVisibleStats bool   `json:"has_community_visible_stats"`
	HasLeaderboards          bool   `json:"has_leaderboards"`
	HasWorkshop              bool   `json:"has_workshop"`
	HasMarket                bool   `json:"has_market"`
	ContentDescriptorIDs     []int  `json:"content_descriptorids"`
	ImgIconURL               string `json:"img_icon_url"`
	ImgLogoURL               string `json:"img_logo_url"`
}

type SteamOwnedGamesResponse struct {
	Response struct {
		GameCount int              `json:"game_count"`
		Games     []SteamGameEntry `json:"games"`
	} `json:"response"`
}

type SteamPlayerSummaryResponse struct {
	Response struct {
		Players []struct {
			SteamID                  string `json:"steamid"`
			CommunityVisibilityState int    `json:"communityvisibilitystate"`
			ProfileState             int    `json:"profilestate"`
			PersonaName              string `json:"personaname"`
			ProfileURL               string `json:"profileurl"`
			Avatar                   string `json:"avatar"`
			AvatarMedium             string `json:"avatarmedium"`
			AvatarFull               string `json:"avatarfull"`
			AvatarHash               string `json:"avatarhash"`
			PersonaState             int    `json:"personastate"`
			RealName                 string `json:"realname"`
			PrimaryClanID            string `json:"primaryclanid"`
			TimeCreated              int64  `json:"timecreated"`
			PersonaStateFlags        int    `json:"personastateflags"`
			LocCountryCode           string `json:"loccountrycode"`
			LocStateCode             string `json:"locstatecode"`
			LocCityID                int    `json:"loccityid"`
		} `json:"players"`
	} `json:"response"`
}
