package responses

type MALImportResponse struct {
	ImportedCount int    `json:"imported_count"`
	SkippedCount  int    `json:"skipped_count"`
	ErrorCount    int    `json:"error_count"`
	Message       string `json:"message"`
}

type MALAnimeEntry struct {
	ID              int     `json:"id"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	Score           *int    `json:"score"`
	WatchedEpisodes int     `json:"watched_episodes"`
	TimesFinished   int     `json:"times_finished"`
	StartDate       *string `json:"start_date"`
	FinishDate      *string `json:"finish_date"`
}
