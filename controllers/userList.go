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

// Create Anime List
// @Summary Create Anime List
// @Description Creates Anime List
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param createanimelist body requests.CreateAnimeList true "Create Anime List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /list/anime [post]
func (u *UserListController) CreateAnimeList(c *gin.Context) {
	var data requests.CreateAnimeList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

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

	userListModel := models.NewUserListModel(u.Database)

	if err := userListModel.CreateAnimeList(uid, data, anime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
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
// @Success 201 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /list/game [post]
func (u *UserListController) CreateGameList(c *gin.Context) {
	var data requests.CreateGameList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

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

	userListModel := models.NewUserListModel(u.Database)

	if err := userListModel.CreateGameList(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
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
// @Success 201 {string} string
// @Failure 500 {string} string
// @Router /list/movie [post]
func (u *UserListController) CreateMovieWatchList(c *gin.Context) {
	var data requests.CreateMovieWatchList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	if err := userListModel.CreateMovieWatchList(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
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
// @Success 201 {string} string
// @Failure 500 {string} string
// @Router /list/tv [post]
func (u *UserListController) CreateTVSeriesWatchList(c *gin.Context) {
	var data requests.CreateTVSeriesWatchList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	if err := userListModel.CreateTVSeriesWatchList(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created."})
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

// Update Anime List
// @Summary Update Anime List
// @Description Updates anime list
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param updateanimelist body requests.UpdateAnimeList true "Update Anime List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/anime [patch]
func (u *UserListController) UpdateAnimeListByID(c *gin.Context) {
	var data requests.UpdateAnimeList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

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

	if data.WatchedEpisodes != nil {
		animeModel := models.NewAnimeModel(u.Database)
		anime, _ := animeModel.GetAnimeDetails(requests.ID{
			ID: animeList.AnimeID,
		})

		if anime.Episodes != nil && *data.WatchedEpisodes > *anime.Episodes {
			data.WatchedEpisodes = anime.Episodes
		}
	}

	if err := userListModel.UpdateAnimeListByID(animeList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Anime list updated."})
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
// @Success 200 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/game [patch]
func (u *UserListController) UpdateGameListByID(c *gin.Context) {
	var data requests.UpdateGameList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

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

	if err := userListModel.UpdateGameListByID(gameList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game list updated."})
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
// @Success 200 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/movie [patch]
func (u *UserListController) UpdateMovieListByID(c *gin.Context) {
	var data requests.UpdateMovieList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

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

	if err := userListModel.UpdateMovieListByID(movieList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie list updated."})
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
// @Success 200 {string} string
// @Failure 403 {string} string "Unauthorized update"
// @Failure 404 {string} string "Could not found"
// @Failure 500 {string} string
// @Router /list/tv [patch]
func (u *UserListController) UpdateTVSeriesListByID(c *gin.Context) {
	var data requests.UpdateTVSeriesList
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

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

	//TODO Implement
	// if data.WatchedEpisodes != nil {
	// 	tvSeriesModel := models.NewTVSeriesModel(u.Database)
	// 	tvSeries, _ := tvSeriesModel.GetTVSeriesDetails(requests.ID{
	// 		ID: tvList.TvID,
	// 	})

	// 	if tvSeries.Episodes != nil && *data.WatchedEpisodes > *tvSeries.Episodes {
	// 		data.WatchedEpisodes = tvSeries.Episodes
	// 	}
	// }

	if err := userListModel.UpdateTVSeriesListByID(tvList, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "TV series watch list updated."})
}

// Get User List
// @Summary Get User List by User ID
// @Description Returns user list by user id
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.UserList
// @Failure 500 {string} string
// @Router /list [get]
func (u *UserListController) GetUserListByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	userList, err := userListModel.GetUserListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": userList})
}

// Get Anime List
// @Summary Get Anime List by User ID
// @Description Returns anime list by user id
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param sortlist query requests.SortList true "Sort List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.AnimeList
// @Failure 500 {string} string
// @Router /list/anime [get]
func (u *UserListController) GetAnimeListByUserID(c *gin.Context) {
	var data requests.SortList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	animeList, err := userListModel.GetAnimeListByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": animeList})
}

// Get Game List
// @Summary Get Game List by User ID
// @Description Returns game list by user id
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param sortlist query requests.SortList true "Sort List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.GameList
// @Failure 500 {string} string
// @Router /list/game [get]
func (u *UserListController) GetGameListByUserID(c *gin.Context) {
	var data requests.SortList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	gameList, err := userListModel.GetGameListByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gameList})
}

// Get Movie Watch List
// @Summary Get Movie Watch List by User ID
// @Description Returns movie watch list by user id
// @Tags user_list
// @Accept application/json
// @Produce application/json
// @Param sortlist query requests.SortList true "Sort List"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.MovieList
// @Failure 500 {string} string
// @Router /list/movie [get]
func (u *UserListController) GetMovieListByUserID(c *gin.Context) {
	var data requests.SortList
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userListModel := models.NewUserListModel(u.Database)

	movieList, err := userListModel.GetMovieListByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movieList})
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

	isDeleted, err := userListModel.DeleteListByUserIDAndType(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "List deleted successfully."})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}
