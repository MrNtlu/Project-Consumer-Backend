package routes

import (
	"app/controllers"
	"app/db"

	"github.com/gin-gonic/gin"
)

func tvRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	tvController := controllers.NewTVController(mongoDB)

	tv := router.Group("/tv")
	{
		tv.GET("/upcoming", tvController.GetUpcomingTVSeries)
		tv.GET("/upcoming/season", tvController.GetUpcomingSeasonTVSeries)
		tv.GET("", tvController.GetTVSeriesBySortAndFilter)
		tv.GET("/decade", tvController.GetPopularTVSeriesByDecade)
		tv.GET("/genre", tvController.GetPopularTVSeriesByGenre)
	}
}
