package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AnimeController struct {
	Database *db.MongoDB
}

func NewAnimeController(mongoDB *db.MongoDB) AnimeController {
	return AnimeController{
		Database: mongoDB,
	}
}

func (a *AnimeController) GetUpcomingAnimesBySort(c *gin.Context) {
	var data requests.SortUpcomingAnime
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	animeModel := models.NewAnimeModel(a.Database)

	upcomingAnimes, pagination, err := animeModel.GetUpcomingAnimesBySort(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingAnimes})
}
