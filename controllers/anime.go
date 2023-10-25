package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"
	"time"

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

// Get Preview Animes
// @Summary Get Preview Animes
// @Description Returns preview animes
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime/preview [get]
func (a *AnimeController) GetPreviewAnimes(c *gin.Context) {
	animeModel := models.NewAnimeModel(a.Database)

	upcomingAnimes, _, err := animeModel.GetUpcomingAnimesBySort(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedAnimes, _, err := animeModel.GetAnimesBySortAndFilter(requests.SortFilterAnime{
		Sort: "top",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularAnimes, _, err := animeModel.GetAnimesBySortAndFilter(requests.SortFilterAnime{
		Sort: "popularity",
		Page: 1,
	})
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

	dayOfWeek := int16(time.Now().UTC().Weekday()) + 1

	var dayOfWeekAnimeList responses.DayOfWeekAnime
	for _, item := range dayOfWeekAnime {
		if item.DayOfWeek == dayOfWeek {
			dayOfWeekAnimeList = item
		}
	}

	c.JSON(http.StatusOK, gin.H{"upcoming": upcomingAnimes, "top": topRatedAnimes, "popular": popularAnimes, "extra": dayOfWeekAnimeList.Data})
}

// Get Upcoming Animes
// @Summary Get Upcoming Animes by Sort
// @Description Returns upcoming animes by sort with pagination
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime/upcoming [get]
func (a *AnimeController) GetUpcomingAnimesBySort(c *gin.Context) {
	var data requests.Pagination
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

	animes, pagination, err := animeModel.GetAnimesByYearAndSeason(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": animes})
}

// Get Upcoming Animes
// @Summary Get Upcoming Animes by Day of Week
// @Description Returns upcoming animes by day of week
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.DayOfWeekAnime
// @Failure 500 {string} string
// @Router /anime/airing [get]
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

	animes, pagination, err := animeModel.GetAnimesBySortAndFilter(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": animes})
}

// Get Anime Details
// @Summary Get Anime Details
// @Description Returns anime details with optional authentication
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Success 200 {array} responses.Anime
// @Success 200 {array} responses.AnimeDetails
// @Failure 500 {string} string
// @Router /anime/details [get]
func (a *AnimeController) GetAnimeDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	animeModel := models.NewAnimeModel(a.Database)
	reviewModel := models.NewReviewModel(a.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		animeDetailsWithWatchList, err := animeModel.GetAnimeDetailsWithWatchList(data, uid.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if animeDetailsWithWatchList.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, nil, &animeDetailsWithWatchList.MalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		animeDetailsWithWatchList.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": animeDetailsWithWatchList,
		})
	} else {
		animeDetails, err := animeModel.GetAnimeDetails(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if animeDetails.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, nil, &animeDetails.MalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		animeDetails.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": animeDetails,
		})
	}
}

// Search Anime
// @Summary Search Anime
// @Description Search animes
// @Tags anime
// @Accept application/json
// @Produce application/json
// @Param search body requests.Search true "Search"
// @Success 200 {array} responses.Anime
// @Failure 500 {string} string
// @Router /anime/search [get]
func (a *AnimeController) SearchAnimeByTitle(c *gin.Context) {
	var data requests.Search
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	animeModel := models.NewAnimeModel(a.Database)

	animes, pagination, err := animeModel.SearchAnimeByTitle(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": animes})
}
