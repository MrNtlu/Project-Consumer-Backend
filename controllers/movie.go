package controllers

import (
	"app/db"
	"app/models"
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

//TODO For testing only, remove later.
func (m *MovieController) GetMovies(c *gin.Context) {
	movieModel := models.NewMovieModel(m.Database)

	movies := movieModel.GetMovies()

	c.JSON(http.StatusOK, gin.H{"data": movies})
}
