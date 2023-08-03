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
		anime.GET("/preview", animeController.GetPreviewAnimes)
		anime.GET("/upcoming", animeController.GetUpcomingAnimesBySort)
		anime.GET("/popular", animeController.GetPopularAnimesBySort)
		anime.GET("/season", animeController.GetAnimesByYearAndSeason)            //TODO Check or delete
		anime.GET("/airing", animeController.GetCurrentlyAiringAnimesByDayOfWeek) //TODO Check or delete
		anime.GET("", animeController.GetAnimesBySortAndFilter)
		anime.Use(helpers.OptionalTokenCheck).GET("/details", animeController.GetAnimeDetails)
		anime.GET("/search", animeController.SearchAnimeByTitle)
	}
}
