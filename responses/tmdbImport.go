package responses

type TMDBImportResponse struct {
	ImportedCount  int      `json:"imported_count"`
	SkippedCount   int      `json:"skipped_count"`
	ErrorCount     int      `json:"error_count"`
	Message        string   `json:"message"`
	ImportedTitles []string `json:"imported_titles"`
	SkippedTitles  []string `json:"skipped_titles"`
}

type TMDBEntry struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	MediaType    string  `json:"media_type"` // "movie" or "tv"
	Rating       float32 `json:"rating"`
	WatchlistID  int     `json:"watchlist_id,omitempty"`
	ReleaseDate  string  `json:"release_date,omitempty"`
	Overview     string  `json:"overview,omitempty"`
	PosterPath   string  `json:"poster_path,omitempty"`
	VoteAverage  float32 `json:"vote_average"`
	VoteCount    int     `json:"vote_count"`
	Popularity   float32 `json:"popularity"`
	OriginalLang string  `json:"original_language,omitempty"`
}

type TMDBWatchlistResponse struct {
	Page         int         `json:"page"`
	Results      []TMDBEntry `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

type TMDBRatedResponse struct {
	Page         int         `json:"page"`
	Results      []TMDBEntry `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

type TMDBFavoriteResponse struct {
	Page         int         `json:"page"`
	Results      []TMDBEntry `json:"results"`
	TotalPages   int         `json:"total_pages"`
	TotalResults int         `json:"total_results"`
}

type TMDBAccountResponse struct {
	Avatar struct {
		Gravatar struct {
			Hash string `json:"hash"`
		} `json:"gravatar"`
		TMDB struct {
			AvatarPath string `json:"avatar_path"`
		} `json:"tmdb"`
	} `json:"avatar"`
	ID           int    `json:"id"`
	ISO639_1     string `json:"iso_639_1"`
	ISO3166_1    string `json:"iso_3166_1"`
	Name         string `json:"name"`
	IncludeAdult bool   `json:"include_adult"`
	Username     string `json:"username"`
}
