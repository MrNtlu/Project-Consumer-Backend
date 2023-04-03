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
// @Tags movie
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

	movieModel := models.NewTVModel(tv.Database)

	upcomingTVSeries, pagination, err := movieModel.GetUpcomingTVSeries(data)
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
// @Tags movie
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

	movieModel := models.NewTVModel(tv.Database)

	upcomingTVSeries, pagination, err := movieModel.GetUpcomingSeasonTVSeries(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingTVSeries})
}
