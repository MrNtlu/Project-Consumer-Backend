package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
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
// @Param sortupcoming body requests.SortUpcoming true "Sort Upcoming"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/upcoming [get]
func (tv *TVController) GetUpcomingTVSeries(c *gin.Context) {
	var data requests.SortUpcoming
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

// Get Upcoming Seasons for TV Series
// @Summary Get Upcoming Seasons for TV Series by Sort
// @Description Returns upcoming tv series by sort with pagination
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param sortupcoming body requests.SortUpcoming true "Sort Upcoming"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/upcoming/season [get]
func (tv *TVController) GetUpcomingSeasonTVSeries(c *gin.Context) {
	var data requests.SortUpcoming
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	upcomingTVSeries, pagination, err := tvModel.GetUpcomingSeasonTVSeries(data)
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

// Get Popular TV Series by Decade
// @Summary Get Popular TV Series by decade
// @Description Returns popular tv series by decade with pagination
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param filterbydecade body requests.FilterByDecade true "Filter by Decade"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/decade [get]
func (tv *TVController) GetPopularTVSeriesByDecade(c *gin.Context) {
	var data requests.FilterByDecade
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	var dateTo = data.Decade + 10

	tvSeries, pagination, err := tvModel.GetTVSeriesBySortAndFilter(requests.SortFilterTVSeries{
		FirstAirDateFrom: &data.Decade,
		FirstAirDateTo:   &dateTo,
		Sort:             "popularity",
		Page:             data.Page,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": tvSeries})
}

// Get Popular TV Series by Genre
// @Summary Get Popular TV Series by genre
// @Description Returns popular tv series by genre with pagination
// @Tags tv
// @Accept application/json
// @Produce application/json
// @Param filterbygenre body requests.FilterByGenre true "Filter by Genre"
// @Success 200 {array} responses.TVSeries
// @Failure 500 {string} string
// @Router /tv/decade [get]
func (tv *TVController) GetPopularTVSeriesByGenre(c *gin.Context) {
	var data requests.FilterByGenre
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	tvModel := models.NewTVModel(tv.Database)

	tvSeries, pagination, err := tvModel.GetTVSeriesBySortAndFilter(requests.SortFilterTVSeries{
		Genres: &data.Genre,
		Sort:   "popularity",
		Page:   data.Page,
	})
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

	tvSeriesDetails, err := tvModel.GetTVSeriesDetails(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	uid, OK := c.Get("uuid")
	if OK && uid != nil && tvSeriesDetails.TitleOriginal != "" {
		tvSeriesDetailsWithWatchList, err := tvModel.GetTVSeriesDetailsWithWatchList(data, uid.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": tvSeriesDetailsWithWatchList,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": tvSeriesDetails,
		})
	}
}
