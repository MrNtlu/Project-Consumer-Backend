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
		anime.GET("/season", animeController.GetAnimesByYearAndSeason) //TODO Check or delete
		anime.GET("/airing", animeController.GetCurrentlyAiringAnimesByDayOfWeek)
		anime.GET("", animeController.GetAnimesBySortAndFilter)
		anime.GET("/popular", animeController.GetPopularAnimes)
		anime.GET("/popular-streaming-services", animeController.GetPopularStreamingServices)
		anime.GET("/popular-studios", animeController.GetPopularStudios)
		anime.Use(helpers.OptionalTokenCheck).GET("/details", animeController.GetAnimeDetails)
		anime.GET("/search", animeController.SearchAnimeByTitle)
	}
}
