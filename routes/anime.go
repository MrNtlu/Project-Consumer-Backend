package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	"github.com/gin-gonic/gin"
)

func animeRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	animeController := controllers.NewAnimeController(mongoDB)

	anime := router.Group("/anime")
	{
		anime.GET("/upcoming", animeController.GetUpcomingAnimesBySort)
		anime.GET("", animeController.GetAnimesBySortAndFilter)
		anime.GET("/streaming-platforms", animeController.GetAnimesByStreamingPlatform)
		anime.GET("/studios", animeController.GetAnimesByStudios)
		anime.GET("/popular", animeController.GetPopularAnimes)
		anime.GET("/popular-streaming-platforms", animeController.GetPopularStreamingPlatforms)
		anime.GET("/popular-studios", animeController.GetPopularStudios)
		anime.Use(helpers.OptionalTokenCheck).GET("/details", animeController.GetAnimeDetails)
		anime.GET("/search", animeController.SearchAnimeByTitle)
	}
}
