package routes

import (
	"app/db"

	"github.com/gin-gonic/gin"
)

func tvRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	// tvController := controllers.NewTVController(mongoDB)

	// tv := router.Group("/tv")
	// {
	// 	tv.GET("", tvController.GetMovies)
	// }
}
