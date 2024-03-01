package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"

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

		var review *responses.Review

		if reviewSummary.IsReviewed {
			reviewResponse, _ := reviewModel.GetBaseReviewResponseByUserIDAndContentID(data.ID, userID)
			review = &reviewResponse
		} else {
			review = nil
		}

		reviewSummary.Review = review
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

// Get Popular Actors
// @Summary Get Popular Actors
// @Description Returns Popular Actors
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.ActorDetails
// @Failure 500 {string} string
// @Router /tv/popular-actors [get]
func (tv *TVController) GetPopularActors(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	actors, err := tvModel.GetPopularActors(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actors})
}

// Get Popular Streaming Platforms
// @Summary Get Popular Streaming Platforms
// @Description Returns Popular Streaming Platforms
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param regionfilters body requests.RegionFilters true "Region Filters"
// @Success 200 {array} responses.StreamingPlatform
// @Failure 500 {string} string
// @Router /tv/popular-streaming-services [get]
func (tv *TVController) GetPopularStreamingPlatforms(c *gin.Context) {
	var data requests.RegionFilters
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	actors, err := tvModel.GetPopularStreamingPlatforms(data.Region)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actors})
}

// Get TV Series by Actors
// @Summary Get TV Series by Actors
// @Description Returns TV Series by Actors
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param idpagination body requests.IDPagination true "ID Pagination"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/actor [get]
func (tv *TVController) GetTVSeriesByActor(c *gin.Context) {
	var data requests.IDPagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	tvSeries, pagination, err := tvModel.GetTVSeriesByActor(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": tvSeries})
}

// Get TV Series by Actors
// @Summary Get TV Series by Actors
// @Description Returns TV Series by Actors
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param filterbystreamingplatformandregion body requests.FilterByStreamingPlatformAndRegion true "Filter By Streaming Platform And Region"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/streaming-services [get]
func (tv *TVController) GetTVSeriesByStreamingPlatform(c *gin.Context) {
	var data requests.FilterByStreamingPlatformAndRegion
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	tvSeries, pagination, err := tvModel.GetTVSeriesByStreamingPlatform(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": tvSeries})
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
