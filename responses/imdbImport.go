package responses

type IMDBImportResponse struct {
	ImportedCount int    `json:"imported_count"`
	SkippedCount  int    `json:"skipped_count"`
	ErrorCount    int    `json:"error_count"`
	Message       string `json:"message"`
}

type IMDBEntry struct {
	IMDBID     string  `json:"imdb_id"`
	Title      string  `json:"title"`
	Year       int     `json:"year"`
	Type       string  `json:"type"`
	UserRating float32 `json:"user_rating"`
}
