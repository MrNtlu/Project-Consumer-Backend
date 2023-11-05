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

type TVController struct {
	Database *db.MongoDB
}

func NewTVController(mongoDB *db.MongoDB) TVController {
	return TVController{
		Database: mongoDB,
	}
}

// Get Preview TV Series
// @Summary Get Preview TV Series
// @Description Returns preview tv series
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/preview [get]
func (tv *TVController) GetPreviewTVSeries(c *gin.Context) {
	tvModel := models.NewTVModel(tv.Database)

	upcomingTVSeries, _, err := tvModel.GetUpcomingTVSeries(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularTVSeries, _, err := tvModel.GetTVSeriesBySortAndFilter(requests.SortFilterTVSeries{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topTVSeries, _, err := tvModel.GetTVSeriesBySortAndFilter(requests.SortFilterTVSeries{
		Sort: "top",
		Page: 1,
	})
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

	c.JSON(http.StatusOK, gin.H{"upcoming": upcomingTVSeries, "popular": popularTVSeries, "top": topTVSeries, "extra": dayOfWeekTVSeriesList.Data})
}

// Get Upcoming TV Series
// @Summary Get Upcoming TV Series by Sort
// @Description Returns upcoming tv series by sort with pagination
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/upcoming [get]
func (tv *TVController) GetUpcomingTVSeries(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	upcomingTVSeries, pagination, err := tvModel.GetUpcomingTVSeries(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingTVSeries})
}

// Get Currently Airing TV Series
// @Summary Get Currently Airing TV Series by day of week and country code
// @Description Returns list of tv series by day of week
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.DayOfWeekTVSeries
// @Failure 500 {string} string
// @Router /tv/airing [get]
func (tv *TVController) GetCurrentlyAiringTVSeriesByDayOfWeek(c *gin.Context) {
	tvModel := models.NewTVModel(tv.Database)

	tvSeriesList, err := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tvSeriesList})
}

// Get TV Series
// @Summary Get TV Series by Sort and Filter
// @Description Returns tv series by sort and filter with pagination
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param sortfiltertvseries body requests.SortFilterTVSeries true "Sort and Filter TV Series"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv [get]
func (tv *TVController) GetTVSeriesBySortAndFilter(c *gin.Context) {
	var data requests.SortFilterTVSeries
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	tvSeries, pagination, err := tvModel.GetTVSeriesBySortAndFilter(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": tvSeries})
}

// Get TV Series Details
// @Summary Get TV Series Details
// @Description Returns tv series details with optional authentication
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Success 200 {array} responses.TVSeries
// @Success 200 {array} responses.TVSeriesDetails
// @Failure 500 {string} string
// @Router /tv/details [get]
func (tv *TVController) GetTVSeriesDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)
	reviewModel := models.NewReviewModel(tv.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userID := uid.(string)

		tvSeriesDetailsWithWatchList, err := tvModel.GetTVSeriesDetailsWithWatchListAndWatchLater(data, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if tvSeriesDetailsWithWatchList.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, userID, &tvSeriesDetailsWithWatchList.TmdbID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		review, _ := reviewModel.GetBaseReviewResponseByUserIDAndContentID(data.ID, userID)

		reviewSummary.Review = &review
		tvSeriesDetailsWithWatchList.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": tvSeriesDetailsWithWatchList,
		})
	} else {
		tvSeriesDetails, err := tvModel.GetTVSeriesDetails(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if tvSeriesDetails.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, "-1", &tvSeriesDetails.TmdbID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		tvSeriesDetails.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": tvSeriesDetails,
		})
	}
}

// Search TV Series
// @Summary Search TV Series
// @Description Search tv series
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param search body requests.Search true "Search"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/search [get]
func (tv *TVController) SearchTVSeriesByTitle(c *gin.Context) {
	var data requests.Search
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	tvSeries, pagination, err := tvModel.SearchTVSeriesByTitle(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": tvSeries})
}
