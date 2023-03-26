package requests

type CreateAnimeList struct {
	AnimeID         string   `json:"anime_id" binding:"required"`
	Status          string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	WatchedEpisodes *int     `json:"watched_episodes" binding:"required,number,min=0"`
	Score           *float32 `json:"score"`
}

type CreateGameList struct {
	GameID            string   `json:"game_id" binding:"required"`
	Status            string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	Score             *float32 `json:"score"`
	AchievementStatus *float32 `json:"achievement_status"`
}

type CreateMovieWatchList struct {
	MovieID string   `json:"movie_id" binding:"required"`
	Status  string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	Score   *float32 `json:"score"`
}

type CreateTVSeriesWatchList struct {
	TvID            string   `json:"tv_id" binding:"required"`
	Status          string   `json:"status" binding:"required,oneof=active finished dropped planto"`
	WatchedEpisodes *int     `json:"watched_episodes" binding:"required,number"`
	WatchedSeasons  *int     `json:"watched_seasons" binding:"required,number"`
	Score           *float32 `json:"score"`
}

type SortList struct {
	Sort string `form:"sort" binding:"required,oneof=popularity new old score"`
}

type DeleteList struct {
	ID   string `json:"id" binding:"required"`
	Type string `json:"type" binding:"required,oneof=anime game movie tv"`
}
