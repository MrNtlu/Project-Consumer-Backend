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
// @Tags lists
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
// @Tags lists
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

	// gameModel := models.NewGameModel(u.Database)
	// game, err := gameModel.GetGameDetails()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"error": err.Error(),
	// 	})

	// 	return
	// }
	//TODO Check if game found

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
// @Tags lists
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
// @Tags lists
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

// Get User List
// @Summary Get User List by User ID
// @Description Returns user list by user id
// @Tags lists
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
// @Tags lists
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
// @Tags lists
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

// Delete List by Type
// @Summary Delete List by Type
// @Description Deletes list by type and user id
// @Tags lists
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
