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

// Get Upcoming Animes
// @Summary Get Upcoming Animes by Sort
// @Description Returns upcoming animes by sort with pagination
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Param sortanime body requests.SortAnime true "Sort Anime"
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime/upcoming [get]
func (a *AnimeController) GetUpcomingAnimesBySort(c *gin.Context) {
	var data requests.SortAnime
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

// Get Animes By Year and Season
// @Summary Get Animes by Year and Season
// @Description Returns animes by year and season
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Param sortbyyearseasonanime body requests.SortByYearSeasonAnime true "Sort Anime By Year and Season"
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime/season [get]
func (a *AnimeController) GetAnimesByYearAndSeason(c *gin.Context) {
	var data requests.SortByYearSeasonAnime
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	animeModel := models.NewAnimeModel(a.Database)

	upcomingAnimes, pagination, err := animeModel.GetAnimesByYearAndSeason(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingAnimes})
}

// Get Upcoming Animes
// @Summary Get Upcoming Animes by Day of Week
// @Description Returns upcoming animes by day of week
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime/upcoming [get]
func (a *AnimeController) GetCurrentlyAiringAnimesByDayOfWeek(c *gin.Context) {
	animeModel := models.NewAnimeModel(a.Database)

	currentlyAiringAnimeResponse, err := animeModel.GetCurrentlyAiringAnimesByDayOfWeek()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": currentlyAiringAnimeResponse})
}

// Get Animes
// @Summary Get Animes by Sort and Filter
// @Description Returns animes by sort and filter
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Param sortfilteranime body requests.SortFilterAnime true "Sort and Filter Anime"
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime [get]
func (a *AnimeController) GetAnimesBySortAndFilter(c *gin.Context) {
	var data requests.SortFilterAnime
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	animeModel := models.NewAnimeModel(a.Database)

	animeList, pagination, err := animeModel.GetAnimesBySortAndFilter(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": animeList})
}
