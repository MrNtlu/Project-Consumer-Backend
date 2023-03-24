package requests

type CreateAnimeList struct {
	AnimeID         string   `json:"anime_id" binding:"required"`
	Status          int      `json:"status" binding:"required,number,oneof=0 1 2 3"`
	WatchedEpisodes int      `json:"watched_episodes" binding:"required,number"`
	Score           *float32 `json:"score"`
}

type CreateGameList struct {
	GameID            string   `json:"game_id" binding:"required"`
	Status            int      `json:"status" binding:"required,number,oneof=0 1 2 3"`
	Score             *float32 `json:"score"`
	AchievementStatus *float32 `json:"achievement_status"`
}

type CreateMovieWatchList struct {
	MovieID string   `json:"movie_id" binding:"required"`
	Status  int      `json:"status" binding:"required,number,oneof=0 1 2 3"`
	Score   *float32 `json:"score"`
}

type CreateTVSeriesWatchList struct {
	TvID            string   `json:"tv_id" binding:"required"`
	Status          int      `json:"status" binding:"required,number,oneof=0 1 2 3"`
	WatchedEpisodes int      `json:"watched_episodes" binding:"required,number"`
	WatchedSeasons  int      `json:"watched_seasons" binding:"required,number"`
	Score           *float32 `json:"score"`
}
