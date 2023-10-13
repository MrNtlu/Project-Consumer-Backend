package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type UserListController struct {
	Database *db.MongoDB
}

func NewUserListController(mongoDB *db.MongoDB) UserListController {
	return UserListController{
		Database: mongoDB,
	}
}

const errUserListPremium = "Free members can add up to 175 content to their list, you can get premium membership for unlimited access."

// Create Anime List
// @Summary Create Anime List
// @Description Creates Anime List
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param createanimelist body requests.CreateAnimeList true "Create Anime List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.AnimeList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /list/anime [post]
func (u *UserListController) CreateAnimeList(c *gin.Context) {
	var data requests.CreateAnimeList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		createdAnimeList models.AnimeList
		err              error
	)

	animeModel := models.NewAnimeModel(u.Database)
	anime, err := animeModel.GetAnimeDetails(requests.ID{
		ID: data.AnimeID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if anime.TitleOriginal == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)
	userListModel := models.NewUserListModel(u.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count, _ := userListModel.GetUserListCount(uid)

	if !isPremium && count >= models.UserListLimit {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errUserListPremium,
		})

		return
	}

	if createdAnimeList, err = userListModel.CreateAnimeList(uid, data, anime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: createdAnimeList.Status,
		ContentTitle:     anime.TitleOriginal,
		ContentImage:     anime.ImageURL,
		ContentType:      "anime",
		ContentID:        createdAnimeList.AnimeID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdAnimeList})
}

// Create Game List
// @Summary Create Game List
// @Description Creates Game List
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param creategamelist body requests.CreateGameList true "Create Game List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.GameList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /list/game [post]
func (u *UserListController) CreateGameList(c *gin.Context) {
	var data requests.CreateGameList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		createdGameList models.GameList
		err             error
	)

	gameModel := models.NewGameModel(u.Database)
	game, err := gameModel.GetGameDetails(requests.ID{
		ID: data.GameID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if game.TitleOriginal == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)
	userListModel := models.NewUserListModel(u.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count, _ := userListModel.GetUserListCount(uid)

	if !isPremium && count >= models.UserListLimit {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errUserListPremium,
		})

		return
	}

	if createdGameList, err = userListModel.CreateGameList(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: createdGameList.Status,
		ContentTitle:     game.Title,
		ContentImage:     game.ImageUrl,
		ContentType:      "game",
		ContentID:        createdGameList.GameID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdGameList})
}

// Create Movie Watch List
// @Summary Create Movie Watch List
// @Description Creates Movie Watch List
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param createmoviewatchlist body requests.CreateMovieWatchList true "Create Movie Watch List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.MovieWatchList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /list/movie [post]
func (u *UserListController) CreateMovieWatchList(c *gin.Context) {
	var data requests.CreateMovieWatchList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		createdWatchList models.MovieWatchList
		err              error
	)

	movieModel := models.NewMovieModel(u.Database)
	movie, err := movieModel.GetMovieDetails(requests.ID{
		ID: data.MovieID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if movie.TitleOriginal == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)
	userListModel := models.NewUserListModel(u.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count, _ := userListModel.GetUserListCount(uid)

	if !isPremium && count >= models.UserListLimit {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errUserListPremium,
		})

		return
	}

	if createdWatchList, err = userListModel.CreateMovieWatchList(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: createdWatchList.Status,
		ContentTitle:     movie.TitleEn,
		ContentImage:     movie.ImageURL,
		ContentType:      "movie",
		ContentID:        createdWatchList.MovieID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdWatchList})
}

// Create TVSeries Watch List
// @Summary Create TVSeries Watch List
// @Description Creates TVSeries Watch List
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param createtvserieswatchlist body requests.CreateTVSeriesWatchList true "Create TVSeries Watch List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.TVSeriesWatchList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /list/tv [post]
func (u *UserListController) CreateTVSeriesWatchList(c *gin.Context) {
	var data requests.CreateTVSeriesWatchList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		createdTVSeriesWatchList models.TVSeriesWatchList
		err                      error
	)

	tvSeriesModel := models.NewTVModel(u.Database)
	tvSeries, err := tvSeriesModel.GetTVSeriesDetails(requests.ID{
		ID: data.TvID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if tvSeries.TitleOriginal == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	userModel := models.NewUserModel(u.Database)
	userListModel := models.NewUserListModel(u.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count, _ := userListModel.GetUserListCount(uid)

	if !isPremium && count >= models.UserListLimit {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errUserListPremium,
		})

		return
	}

	if createdTVSeriesWatchList, err = userListModel.CreateTVSeriesWatchList(uid, data, tvSeries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: createdTVSeriesWatchList.Status,
		ContentTitle:     tvSeries.TitleEn,
		ContentImage:     tvSeries.ImageURL,
		ContentType:      "tv",
		ContentID:        createdTVSeriesWatchList.TvID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdTVSeriesWatchList})
}

// Update User List Visibility
// @Summary Update User List Visibility
// @Description Updates user list is_public value
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param updateuserlist body requests.UpdateUserList true "Update User List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list [patch]
func (u *UserListController) UpdateUserListPublicVisibility(c *gin.Context) {
	var data requests.UpdateUserList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	userList, err := userListModel.GetBaseUserListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if err := userListModel.UpdateUserListPublicVisibility(userList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User list visibility updated."})
}

// Increment Anime List Episode
// @Summary Increment Anime List
// @Description Increment anime list episode
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.AnimeList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/anime/inc [patch]
func (u *UserListController) IncrementAnimeListEpisodeByID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedAnimeList models.AnimeList
		err              error
	)

	userListModel := models.NewUserListModel(u.Database)
	animeList, err := userListModel.GetBaseAnimeListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if animeList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != animeList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	animeModel := models.NewAnimeModel(u.Database)
	anime, _ := animeModel.GetAnimeDetails(requests.ID{
		ID: animeList.AnimeID,
	})

	if updatedAnimeList, err = userListModel.IncrementAnimeListEpisodeByID(animeList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedAnimeList.Status,
		ContentTitle:     anime.TitleOriginal,
		ContentImage:     anime.ImageURL,
		ContentType:      "anime",
		ContentID:        updatedAnimeList.AnimeID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Anime list updated.", "data": updatedAnimeList})
}

// Update Anime List
// @Summary Update Anime List
// @Description Updates anime list
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param updateanimelist body requests.UpdateAnimeList true "Update Anime List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.AnimeList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/anime [patch]
func (u *UserListController) UpdateAnimeListByID(c *gin.Context) {
	var data requests.UpdateAnimeList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedAnimeList models.AnimeList
		err              error
	)

	userListModel := models.NewUserListModel(u.Database)
	animeList, err := userListModel.GetBaseAnimeListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if animeList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != animeList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	animeModel := models.NewAnimeModel(u.Database)
	anime, _ := animeModel.GetAnimeDetails(requests.ID{
		ID: animeList.AnimeID,
	})

	if data.WatchedEpisodes != nil && (anime.Episodes != nil && *data.WatchedEpisodes > *anime.Episodes) {
		data.WatchedEpisodes = anime.Episodes
	}

	if updatedAnimeList, err = userListModel.UpdateAnimeListByID(animeList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedAnimeList.Status,
		ContentTitle:     anime.TitleOriginal,
		ContentImage:     anime.ImageURL,
		ContentType:      "anime",
		ContentID:        updatedAnimeList.AnimeID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Anime list updated.", "data": updatedAnimeList})
}

// Increment Game List
// @Summary Increment Game List
// @Description Increment game list hours played
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.GameList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/game [patch]
func (u *UserListController) IncrementGameListHourByID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedGameList models.GameList
		err             error
	)

	userListModel := models.NewUserListModel(u.Database)
	gameList, err := userListModel.GetBaseGameListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if gameList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != gameList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	if updatedGameList, err = userListModel.IncrementGameListHourByID(gameList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	gameModel := models.NewGameModel(u.Database)
	game, _ := gameModel.GetGameDetails(requests.ID{
		ID: gameList.GameID,
	})

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedGameList.Status,
		ContentTitle:     game.Title,
		ContentImage:     game.ImageUrl,
		ContentType:      "game",
		ContentID:        updatedGameList.GameID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Game list updated.", "data": updatedGameList})
}

// Update Game List
// @Summary Update Game List
// @Description Updates game list
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param updategamelist body requests.UpdateGameList true "Update Game List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.GameList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/game [patch]
func (u *UserListController) UpdateGameListByID(c *gin.Context) {
	var data requests.UpdateGameList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedGameList models.GameList
		err             error
	)

	userListModel := models.NewUserListModel(u.Database)
	gameList, err := userListModel.GetBaseGameListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if gameList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != gameList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	if updatedGameList, err = userListModel.UpdateGameListByID(gameList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	gameModel := models.NewGameModel(u.Database)
	game, _ := gameModel.GetGameDetails(requests.ID{
		ID: gameList.GameID,
	})

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedGameList.Status,
		ContentTitle:     game.Title,
		ContentImage:     game.ImageUrl,
		ContentType:      "game",
		ContentID:        updatedGameList.GameID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Game list updated.", "data": updatedGameList})
}

// Update Movie List
// @Summary Update Movie List
// @Description Updates movie list
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param updatemovielist body requests.UpdateMovieList true "Update Movie List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.MovieWatchList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/movie [patch]
func (u *UserListController) UpdateMovieListByID(c *gin.Context) {
	var data requests.UpdateMovieList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedWatchList models.MovieWatchList
		err              error
	)

	userListModel := models.NewUserListModel(u.Database)
	movieList, err := userListModel.GetBaseMovieListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if movieList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != movieList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	if updatedWatchList, err = userListModel.UpdateMovieListByID(movieList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	movieModel := models.NewMovieModel(u.Database)
	movie, _ := movieModel.GetMovieDetails(requests.ID{
		ID: movieList.MovieID,
	})

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedWatchList.Status,
		ContentTitle:     movie.TitleEn,
		ContentImage:     movie.ImageURL,
		ContentType:      "movie",
		ContentID:        updatedWatchList.MovieID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Movie list updated.", "data": updatedWatchList})
}

// Increment TV Series List
// @Summary Increment TV Series List
// @Description Increment tv series list episode or season
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param incrementtvserieslist body requests.IncrementTVSeriesList true "Increment TV Series List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.TVSeriesWatchList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/tv [patch]
func (u *UserListController) IncrementTVSeriesListEpisodeSeasonByID(c *gin.Context) {
	var data requests.IncrementTVSeriesList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedTVList models.TVSeriesWatchList
		err           error
	)

	userListModel := models.NewUserListModel(u.Database)
	tvList, err := userListModel.GetBaseTVSeriesListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tvList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != tvList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	tvSeriesModel := models.NewTVModel(u.Database)
	tvSeries, _ := tvSeriesModel.GetTVSeriesDetails(requests.ID{
		ID: tvList.TvID,
	})

	if updatedTVList, err = userListModel.IncrementTVSeriesListEpisodeSeasonByID(tvList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedTVList.Status,
		ContentTitle:     tvSeries.TitleEn,
		ContentImage:     tvSeries.ImageURL,
		ContentType:      "tv",
		ContentID:        updatedTVList.TvID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "TV series watch list updated.", "data": updatedTVList})
}

// Update TV Series List
// @Summary Update TV Series List
// @Description Updates tv series list
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param updatetvserieslist body requests.UpdateTVSeriesList true "Update TV Series List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} models.TVSeriesWatchList
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/tv [patch]
func (u *UserListController) UpdateTVSeriesListByID(c *gin.Context) {
	var data requests.UpdateTVSeriesList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		updatedTVList models.TVSeriesWatchList
		err           error
	)

	userListModel := models.NewUserListModel(u.Database)
	tvList, err := userListModel.GetBaseTVSeriesListByID(data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tvList.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	if uid != tvList.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": ErrUnauthorized})
		return
	}

	tvSeriesModel := models.NewTVModel(u.Database)
	tvSeries, _ := tvSeriesModel.GetTVSeriesDetails(requests.ID{
		ID: tvList.TvID,
	})

	if data.WatchedEpisodes != nil && (tvSeries.TotalEpisodes != 0 && *data.WatchedEpisodes > tvSeries.TotalEpisodes) {
		data.WatchedEpisodes = &tvSeries.TotalEpisodes
	}

	if updatedTVList, err = userListModel.UpdateTVSeriesListByID(tvList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(u.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.UserListLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: updatedTVList.Status,
		ContentTitle:     tvSeries.TitleEn,
		ContentImage:     tvSeries.ImageURL,
		ContentType:      "tv",
		ContentID:        updatedTVList.TvID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "TV series watch list updated.", "data": updatedTVList})
}

// Get User List
// @Summary Get User List by User ID
// @Description Returns user list by user id
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param sortlist query requests.SortList true "Sort List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.UserList
// @Failure 500 {string} string
// @Router /list [get]
func (u *UserListController) GetUserListByUserID(c *gin.Context) {
	var data requests.SortList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	userList, err := userListModel.GetUserListByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userList})
}

// Get Logs
// @Summary Get Logs by User ID and date range
// @Description Returns logs by user id and date range
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param logsbydaterange query requests.LogsByDateRange true "Logs by Date Range"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.LogsByRange
// @Failure 500 {string} string
// @Router /list/logs [get]
func (u *UserListController) GetLogsByDateRange(c *gin.Context) {
	var data requests.LogsByDateRange
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	logsModel := models.NewLogsModel(u.Database)

	logs, err := logsModel.GetLogsByDateRange(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// Delete List by Type
// @Summary Delete List by Type
// @Description Deletes list by type and user id
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param deletelist body requests.DeleteList true "Delete List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /list [delete]
func (u *UserListController) DeleteListByUserIDAndType(c *gin.Context) {
	var data requests.DeleteList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	var (
		contentTitle string
		contentImage string
		contentID    string
	)

	switch data.Type {
	case "anime":
		animeModel := models.NewAnimeModel(u.Database)

		animeList, _ := userListModel.GetBaseAnimeListByID(data.ID)
		anime, _ := animeModel.GetAnimeDetails(requests.ID{
			ID: animeList.AnimeID,
		})

		contentTitle = anime.TitleOriginal
		contentImage = anime.ImageURL
		contentID = animeList.AnimeID
	case "game":
		gameModel := models.NewGameModel(u.Database)

		gameList, _ := userListModel.GetBaseGameListByID(data.ID)
		game, _ := gameModel.GetGameDetails(requests.ID{
			ID: gameList.GameID,
		})

		contentTitle = game.Title
		contentImage = game.ImageUrl
		contentID = gameList.GameID
	case "movie":
		movieModel := models.NewMovieModel(u.Database)

		movieList, _ := userListModel.GetBaseMovieListByID(data.ID)
		movie, _ := movieModel.GetMovieDetails(requests.ID{
			ID: movieList.MovieID,
		})

		contentTitle = movie.TitleEn
		contentImage = movie.ImageURL
		contentID = movieList.MovieID
	case "tv":
		tvSeriesModel := models.NewTVModel(u.Database)

		tvList, _ := userListModel.GetBaseTVSeriesListByID(data.ID)
		tvSeries, _ := tvSeriesModel.GetTVSeriesDetails(requests.ID{
			ID: tvList.TvID,
		})

		contentTitle = tvSeries.TitleEn
		contentImage = tvSeries.ImageURL
		contentID = tvList.TvID
	}

	isDeleted, err := userListModel.DeleteListByUserIDAndType(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		logModel := models.NewLogsModel(u.Database)

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.UserListLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      data.Type,
			ContentID:        contentID,
		})

		c.JSON(http.StatusOK, gin.H{"message": "List deleted successfully."})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}
