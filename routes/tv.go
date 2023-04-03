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
	}
}
