package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	Database *db.MongoDB
}

func NewGameController(mongoDB *db.MongoDB) GameController {
	return GameController{
		Database: mongoDB,
	}
}

// Get Preview Games
// @Summary Get Preview Games
// @Description Returns preview games
// @Tags game
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.Game
// @Failure 500 {string} string
// @Router /game/preview [get]
func (g *GameController) GetPreviewGames(c *gin.Context) {
	gameModel := models.NewGameModel(g.Database)

	upcomingGames, _, err := gameModel.GetUpcomingGamesBySort(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedGames, _, err := gameModel.GetGamesByFilterAndSort(requests.SortFilterGame{
		Sort: "top",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularGames, _, err := gameModel.GetGamesByFilterAndSort(requests.SortFilterGame{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"upcoming": upcomingGames, "top": topRatedGames, "popular": popularGames})
}

// Get Upcoming Games
// @Summary Get Upcoming Games by Sort
// @Description Returns upcoming games by sort with pagination
// @Tags game
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.Game
// @Failure 500 {string} string
// @Router /game/upcoming [get]
func (g *GameController) GetUpcomingGamesBySort(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	gameModel := models.NewGameModel(g.Database)

	upcomingGames, pagination, err := gameModel.GetUpcomingGamesBySort(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingGames})
}

// Get Games
// @Summary Get Games by Sort and Filter
// @Description Returns games by sort and filter
// @Tags game
// @Accept application/json
// @Produce application/json
// @Param sortfiltergame body requests.SortFilterGame true "Sort and Filter Game"
// @Success 200 {array} responses.Game
// @Failure 500 {string} string
// @Router /game [get]
func (g *GameController) GetGamesByFilterAndSort(c *gin.Context) {
	var data requests.SortFilterGame
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	gameModel := models.NewGameModel(g.Database)

	games, pagination, err := gameModel.GetGamesByFilterAndSort(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": games})
}

// Get Game Details
// @Summary Get Game Details
// @Description Returns game details with optional authentication
// @Tags game
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Success 200 {array} responses.GameDetails
// @Failure 500 {string} string
// @Router /game/details [get]
func (g *GameController) GetGameDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	gameModel := models.NewGameModel(g.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		gameDetailsWithPlayList, err := gameModel.GetGameDetailsWithPlayList(data, uid.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gameDetailsWithPlayList,
		})
	} else {
		gameDetails, err := gameModel.GetGameDetails(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gameDetails,
		})
	}
}
