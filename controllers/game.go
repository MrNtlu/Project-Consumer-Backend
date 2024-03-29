package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
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
	reviewModel := models.NewReviewModel(g.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userID := uid.(string)

		gameDetailsWithPlayList, err := gameModel.GetGameDetailsWithPlayList(data, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if gameDetailsWithPlayList.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, userID, nil, &gameDetailsWithPlayList.RawgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		var review *responses.Review

		if reviewSummary.IsReviewed {
			reviewResponse, _ := reviewModel.GetBaseReviewResponseByUserIDAndContentID(data.ID, userID)
			review = &reviewResponse
		} else {
			review = nil
		}

		reviewSummary.Review = review
		gameDetailsWithPlayList.Review = reviewSummary

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

		if gameDetails.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, "-1", nil, &gameDetails.RawgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		gameDetails.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": gameDetails,
		})
	}
}

// Search Game
// @Summary Search Game
// @Description Search Game
// @Tags game
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Param search body requests.Search true "Search"
// @Success 200 {array} responses.Game
// @Failure 500 {string} string
// @Router /game/search [get]
func (g *GameController) SearchGameByTitle(c *gin.Context) {
	var data requests.Search
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	gameModel := models.NewGameModel(g.Database)

	animes, pagination, err := gameModel.SearchGameByTitle(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": animes})
}
