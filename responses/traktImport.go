package responses

type TraktImportResponse struct {
	ImportedCount  int      `json:"imported_count"`
	SkippedCount   int      `json:"skipped_count"`
	ErrorCount     int      `json:"error_count"`
	Message        string   `json:"message"`
	ImportedTitles []string `json:"imported_titles"`
	SkippedTitles  []string `json:"skipped_titles"`
}

type TraktEntry struct {
	ID        int     `json:"id"`
	Title     string  `json:"title"`
	MediaType string  `json:"media_type"` // "movie" or "show"
	Rating    *int    `json:"rating"`
	WatchedAt *string `json:"watched_at"`
	Plays     int     `json:"plays"`
	Status    string  `json:"status"`
}

type TraktMovie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		IMDB  string `json:"imdb"`
		TMDB  int    `json:"tmdb"`
	} `json:"ids"`
}

type TraktShow struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
	IDs   struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
		IMDB  string `json:"imdb"`
		TMDB  int    `json:"tmdb"`
		TVDB  int    `json:"tvdb"`
	} `json:"ids"`
}

type TraktWatchedMovie struct {
	Plays         int        `json:"plays"`
	LastWatchedAt string     `json:"last_watched_at"`
	LastUpdatedAt string     `json:"last_updated_at"`
	Movie         TraktMovie `json:"movie"`
}

type TraktWatchedShow struct {
	Plays         int       `json:"plays"`
	LastWatchedAt string    `json:"last_watched_at"`
	LastUpdatedAt string    `json:"last_updated_at"`
	Show          TraktShow `json:"show"`
	Seasons       []struct {
		Number   int `json:"number"`
		Episodes []struct {
			Number        int    `json:"number"`
			Plays         int    `json:"plays"`
			LastWatchedAt string `json:"last_watched_at"`
		} `json:"episodes"`
	} `json:"seasons"`
}

type TraktWatchlistItem struct {
	Rank     int         `json:"rank"`
	ListedAt string      `json:"listed_at"`
	Type     string      `json:"type"`
	Movie    *TraktMovie `json:"movie,omitempty"`
	Show     *TraktShow  `json:"show,omitempty"`
}

type TraktCollectionMovie struct {
	CollectedAt string     `json:"collected_at"`
	UpdatedAt   string     `json:"updated_at"`
	Movie       TraktMovie `json:"movie"`
	Metadata    struct {
		MediaType     string `json:"media_type"`
		Resolution    string `json:"resolution"`
		Audio         string `json:"audio"`
		AudioChannels string `json:"audio_channels"`
		HDR           string `json:"hdr"`
	} `json:"metadata"`
}

type TraktCollectionShow struct {
	CollectedAt string    `json:"collected_at"`
	UpdatedAt   string    `json:"updated_at"`
	Show        TraktShow `json:"show"`
	Seasons     []struct {
		Number   int `json:"number"`
		Episodes []struct {
			Number      int    `json:"number"`
			CollectedAt string `json:"collected_at"`
		} `json:"episodes"`
	} `json:"seasons"`
}

type TraktRatingItem struct {
	RatedAt string      `json:"rated_at"`
	Rating  int         `json:"rating"`
	Type    string      `json:"type"`
	Movie   *TraktMovie `json:"movie,omitempty"`
	Show    *TraktShow  `json:"show,omitempty"`
}

type TraktUserStats struct {
	Movies struct {
		Plays     int `json:"plays"`
		Watched   int `json:"watched"`
		Minutes   int `json:"minutes"`
		Collected int `json:"collected"`
		Ratings   int `json:"ratings"`
		Comments  int `json:"comments"`
	} `json:"movies"`
	Shows struct {
		Watched   int `json:"watched"`
		Collected int `json:"collected"`
		Ratings   int `json:"ratings"`
		Comments  int `json:"comments"`
	} `json:"shows"`
	Seasons struct {
		Ratings  int `json:"ratings"`
		Comments int `json:"comments"`
	} `json:"seasons"`
	Episodes struct {
		Plays     int `json:"plays"`
		Watched   int `json:"watched"`
		Minutes   int `json:"minutes"`
		Collected int `json:"collected"`
		Ratings   int `json:"ratings"`
		Comments  int `json:"comments"`
	} `json:"episodes"`
}
