package requests

type CreateAnimeList struct {
	AnimeID         string   `json:"anime_id" binding:"required"`
	Status          string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	WatchedEpisodes *int64   `json:"watched_episodes" binding:"required,number,min=0"`
	Score           *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
}

type CreateGameList struct {
	GameID            string   `json:"game_id" binding:"required"`
	Status            string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	Score             *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
	AchievementStatus *float32 `json:"achievement_status"`
}

type CreateMovieWatchList struct {
	MovieID string   `json:"movie_id" binding:"required"`
	Status  string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	Score   *float32 `json:"score" binding:"omitempty,number,min=0,max=10"`
}

type CreateTVSeriesWatchList struct {
	TvID            string   `json:"tv_id" binding:"required"`
	Status          string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	WatchedEpisodes *int     `json:"watched_episodes" binding:"required,number"`
	WatchedSeasons  *int     `json:"watched_seasons" binding:"required,number"`
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
	Status            *string  `json:"status" binding:"omitempty,oneof=active finished dropped planto"`
	AchievementStatus *float32 `json:"achievement_status" binding:"omitempty,number,min=0"`
}

type DeleteList struct {
	ID   string `json:"id" binding:"required"`
	Type string `json:"type" binding:"required,oneof=anime game movie tv"`
}
