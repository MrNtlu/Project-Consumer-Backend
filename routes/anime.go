package routes

import (
	"app/controllers"
	"app/db"

	"github.com/gin-gonic/gin"
)

func animeRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	animeController := controllers.NewAnimeController(mongoDB)

	anime := router.Group("/anime")
	{
		anime.GET("/upcoming", animeController.GetUpcomingAnimesBySort)
	}
}
