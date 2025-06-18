package responses

type AniListImportResponse struct {
	ImportedCount  int      `json:"imported_count"`
	SkippedCount   int      `json:"skipped_count"`
	ErrorCount     int      `json:"error_count"`
	Message        string   `json:"message"`
	ImportedTitles []string `json:"imported_titles"`
	SkippedTitles  []string `json:"skipped_titles"`
}

type AniListEntry struct {
	ID              int     `json:"id"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	Score           *int    `json:"score"`
	Progress        int     `json:"progress"`
	ProgressVolumes *int    `json:"progress_volumes,omitempty"`
	Repeat          int     `json:"repeat"`
	MediaType       string  `json:"media_type"` // "ANIME" or "MANGA"
	StartDate       *string `json:"start_date"`
	CompletedDate   *string `json:"completed_date"`
	UpdatedAt       *string `json:"updated_at"`
}

type AniListMedia struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Episodes *int   `json:"episodes"`
	Chapters *int   `json:"chapters"`
	Volumes  *int   `json:"volumes"`
	Status   string `json:"status"`
	Type     string `json:"type"`
}

type AniListMediaListEntry struct {
	ID                    int          `json:"id"`
	Status                string       `json:"status"`
	Score                 int          `json:"score"`
	Progress              int          `json:"progress"`
	ProgressVolumes       *int         `json:"progressVolumes"`
	Repeat                int          `json:"repeat"`
	Priority              int          `json:"priority"`
	Private               bool         `json:"private"`
	Notes                 string       `json:"notes"`
	HiddenFromStatusLists bool         `json:"hiddenFromStatusLists"`
	CustomLists           []string     `json:"customLists"`
	StartedAt             *AniListDate `json:"startedAt"`
	CompletedAt           *AniListDate `json:"completedAt"`
	UpdatedAt             int64        `json:"updatedAt"`
	CreatedAt             int64        `json:"createdAt"`
	Media                 AniListMedia `json:"media"`
}

type AniListDate struct {
	Year  *int `json:"year"`
	Month *int `json:"month"`
	Day   *int `json:"day"`
}

type AniListMediaListCollection struct {
	Lists []struct {
		Name                 string                  `json:"name"`
		IsCustomList         bool                    `json:"isCustomList"`
		IsSplitCompletedList bool                    `json:"isSplitCompletedList"`
		Status               string                  `json:"status"`
		Entries              []AniListMediaListEntry `json:"entries"`
	} `json:"lists"`
	User struct {
		ID               int    `json:"id"`
		Name             string `json:"name"`
		MediaListOptions struct {
			ScoreFormat string `json:"scoreFormat"`
			AnimeList   struct {
				SectionOrder                  []string `json:"sectionOrder"`
				SplitCompletedSectionByFormat bool     `json:"splitCompletedSectionByFormat"`
				CustomLists                   []string `json:"customLists"`
			} `json:"animeList"`
			MangaList struct {
				SectionOrder                  []string `json:"sectionOrder"`
				SplitCompletedSectionByFormat bool     `json:"splitCompletedSectionByFormat"`
				CustomLists                   []string `json:"customLists"`
			} `json:"mangaList"`
		} `json:"mediaListOptions"`
	} `json:"user"`
}

type AniListGraphQLResponse struct {
	Data struct {
		MediaListCollection AniListMediaListCollection `json:"MediaListCollection"`
	} `json:"data"`
	Errors []struct {
		Message   string `json:"message"`
		Status    int    `json:"status"`
		Locations []struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"locations"`
	} `json:"errors,omitempty"`
}
