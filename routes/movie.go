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
		movie.GET("/upcoming", movieController.GetUpcomingMoviesBySort)
		movie.GET("", movieController.GetMoviesBySortAndFilter)
		movie.GET("/decade", movieController.GetPopularMoviesByDecade)
		movie.GET("/genre", movieController.GetPopularMoviesByGenre)
		movie.Use(helpers.OptionalTokenCheck).GET("/details", movieController.GetMovieDetails)
		movie.GET("/search", movieController.SearchMovieByTitle)
	}
}
