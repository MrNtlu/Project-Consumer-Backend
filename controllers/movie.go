package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MovieController struct {
	Database *db.MongoDB
}

func NewMovieController(mongoDB *db.MongoDB) MovieController {
	return MovieController{
		Database: mongoDB,
	}
}

// Get Upcoming Movies
// @Summary Get Upcoming Movies by Sort
// @Description Returns upcoming movies by sort with pagination
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param sortupcoming body requests.SortUpcoming true "Sort Upcoming"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/upcoming [get]
func (m *MovieController) GetUpcomingMoviesBySort(c *gin.Context) {
	var data requests.SortUpcoming
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	upcomingMovies, pagination, err := movieModel.GetUpcomingMoviesBySort(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingMovies})
}
