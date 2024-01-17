package controllers

import (
	"app/db"
	"app/models"
	"app/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PreviewController struct {
	Database *db.MongoDB
}

func NewPreviewController(mongoDB *db.MongoDB) PreviewController {
	return PreviewController{
		Database: mongoDB,
	}
}

// Get Previews
// @Summary Get Previews
// @Description Returns previews
// @Tags preview
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.PreviewMovie
// @Success 200 {array} responses.PreviewAnime
// @Success 200 {array} responses.PreviewTVSeries
// @Success 200 {array} responses.PreviewGame
// @Failure 500 {string} string
// @Router /preview [get]
func (pr *PreviewController) GetHomePreview(c *gin.Context) {
	movieModel := models.NewMovieModel(pr.Database)
	tvModel := models.NewTVModel(pr.Database)
	animeModel := models.NewAnimeModel(pr.Database)
	gameModel := models.NewGameModel(pr.Database)

	upcomingMovies, err := movieModel.GetUpcomingPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularMovies, err := movieModel.GetPopularPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topMovies, err := movieModel.GetTopPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	moviesInTheater, err := movieModel.GetInTheaterPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// TV Series

	upcomingTVSeries, err := tvModel.GetUpcomingPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularTVSeries, err := tvModel.GetPopularPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topTVSeries, err := tvModel.GetTopPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeekTVSeries, err := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeek := int16(time.Now().UTC().Weekday()) + 1

	var dayOfWeekTVSeriesList responses.DayOfWeekTVSeries
	for _, item := range dayOfWeekTVSeries {
		if item.DayOfWeek == dayOfWeek {
			dayOfWeekTVSeriesList = item
		}
	}

	// Anime

	upcomingAnimes, err := animeModel.GetPreviewUpcomingAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedAnimes, err := animeModel.GetPreviewTopAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularAnimes, err := animeModel.GetPreviewPopularAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeekAnime, err := animeModel.GetCurrentlyAiringAnimesByDayOfWeek()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	var dayOfWeekAnimeList responses.DayOfWeekAnime
	for _, item := range dayOfWeekAnime {
		if item.DayOfWeek == dayOfWeek {
			dayOfWeekAnimeList = item
		}
	}

	// Game

	upcomingGames, err := gameModel.GetPreviewUpcomingGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedGames, err := gameModel.GetPreviewTopGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularGames, err := gameModel.GetPreviewPopularGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movie": gin.H{"upcoming": upcomingMovies, "popular": popularMovies, "top": topMovies, "extra": moviesInTheater},
		"tv":    gin.H{"upcoming": upcomingTVSeries, "popular": popularTVSeries, "top": topTVSeries, "extra": dayOfWeekTVSeriesList.Data},
		"anime": gin.H{"upcoming": upcomingAnimes, "top": topRatedAnimes, "popular": popularAnimes, "extra": dayOfWeekAnimeList.Data},
		"game":  gin.H{"upcoming": upcomingGames, "top": topRatedGames, "popular": popularGames, "extra": nil},
	})

}
