package requests

type CreateAnimeList struct {
	AnimeID         string   `json:"anime_id" binding:"required"`
	AnimeMALID      int64    `json:"anime_mal_id" binding:"required"`
	Status          string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	WatchedEpisodes *int64   `json:"watched_episodes" binding:"required,number,min=0"`
	TimesFinished   *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	Score           *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
}

type CreateGameList struct {
	GameID            string   `json:"game_id" binding:"required"`
	GameRAWGID        int64    `json:"game_rawg_id" binding:"required"`
	Status            string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	Score             *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
	TimesFinished     *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	HoursPlayed       *int     `json:"hours_played" binding:"omitempty,number,min=0"`
	AchievementStatus *float32 `json:"achievement_status" binding:"omitempty,number,min=0,max=100"`
}

type CreateMovieWatchList struct {
	MovieID       string   `json:"movie_id" binding:"required"`
	MovieTmdbID   string   `json:"movie_tmdb_id" binding:"required"`
	Status        string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	TimesFinished *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	Score         *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
}

type CreateTVSeriesWatchList struct {
	TvID            string   `json:"tv_id" binding:"required"`
	TvTmdbID        string   `json:"tv_tmdb_id" binding:"required"`
	Status          string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	WatchedEpisodes *int     `json:"watched_episodes" binding:"required,number,min=0"`
	WatchedSeasons  *int     `json:"watched_seasons" binding:"required,number,min=0"`
	TimesFinished   *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	Score           *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
}

type SortList struct {
	Sort string `form:"sort" binding:"required,oneof=popularity new old score"`
}

type UpdateUserList struct {
	IsPublic bool `json:"is_public"`
}

type UpdateAnimeList struct {
	ID              string   `json:"id" binding:"required"`
	IsUpdatingScore bool     `json:"is_updating_score"`
	Score           *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
	TimesFinished   *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	Status          *string  `json:"status" binding:"omitempty,oneof=active finished dropped planto"`
	WatchedEpisodes *int64   `json:"watched_episodes" binding:"omitempty,number,min=0"`
}

type UpdateGameList struct {
	ID                string   `json:"id" binding:"required"`
	IsUpdatingScore   bool     `json:"is_updating_score"`
	Score             *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
	TimesFinished     *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	HoursPlayed       *int     `json:"hours_played" binding:"omitempty,number,min=0"`
	Status            *string  `json:"status" binding:"omitempty,oneof=active finished dropped planto"`
	AchievementStatus *float32 `json:"achievement_status" binding:"omitempty,number,min=0"`
}

type UpdateMovieList struct {
	ID              string   `json:"id" binding:"required"`
	IsUpdatingScore bool     `json:"is_updating_score"`
	Score           *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
	TimesFinished   *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	Status          *string  `json:"status" binding:"omitempty,oneof=active finished dropped planto"`
}

type UpdateTVSeriesList struct {
	ID              string   `json:"id" binding:"required"`
	IsUpdatingScore bool     `json:"is_updating_score"`
	Score           *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
	TimesFinished   *int     `json:"times_finished" binding:"omitempty,number,min=0"`
	Status          *string  `json:"status" binding:"omitempty,oneof=active finished dropped planto"`
	WatchedEpisodes *int     `json:"watched_episodes" binding:"omitempty,number,min=0"`
	WatchedSeasons  *int     `json:"watched_seasons" binding:"omitempty,number,min=0"`
}

type DeleteList struct {
	ID   string `json:"id" binding:"required"`
	Type string `json:"type" binding:"required,oneof=anime game movie tv"`
}
