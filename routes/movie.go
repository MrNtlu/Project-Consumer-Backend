package routes

import (
	"app/controllers"
	"app/db"

	"github.com/gin-gonic/gin"
)

func movieRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	movieController := controllers.NewMovieController(mongoDB)

	movie := router.Group("/movie")
	{
		movie.GET("", movieController.GetMovies)
	}
}
