package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	"github.com/gin-gonic/gin"
)

func tvRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	tvController := controllers.NewTVController(mongoDB)

	tv := router.Group("/tv")
	{
		tv.GET("/upcoming", tvController.GetUpcomingTVSeries)
		tv.GET("/airing", tvController.GetCurrentlyAiringTVSeriesByDayOfWeek)
		tv.GET("", tvController.GetTVSeriesBySortAndFilter)
		tv.Use(helpers.OptionalTokenCheck).GET("/details", tvController.GetTVSeriesDetails)
		tv.GET("/search", tvController.SearchTVSeriesByTitle)
	}
}
