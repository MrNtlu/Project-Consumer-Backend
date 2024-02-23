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
		movie.GET("/actor", movieController.GetMoviesByActor)
		movie.GET("/streaming-platforms", movieController.GetMoviesByStreamingPlatform)
		movie.GET("/popular-actors", movieController.GetPopularActors)
		movie.GET("/popular-streaming-platforms", movieController.GetPopularStreamingPlatforms)
		movie.Use(helpers.OptionalTokenCheck).GET("/details", movieController.GetMovieDetails)
		movie.GET("/search", movieController.SearchMovieByTitle)
		movie.GET("/theaters", movieController.GetMoviesInTheater)
	}
}
