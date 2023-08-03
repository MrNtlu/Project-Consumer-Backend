package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	"github.com/gin-gonic/gin"
)

func movieRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	movieController := controllers.NewMovieController(mongoDB)

	movie := router.Group("/movie")
	{
		movie.GET("/preview", movieController.GetPreviewMovies)
		movie.GET("/upcoming", movieController.GetUpcomingMoviesBySort)
		movie.GET("/top", movieController.GetTopRatedMoviesBySort)
		movie.GET("", movieController.GetMoviesBySortAndFilter)
		movie.Use(helpers.OptionalTokenCheck).GET("/details", movieController.GetMovieDetails)
		movie.GET("/search", movieController.SearchMovieByTitle)
	}
}
