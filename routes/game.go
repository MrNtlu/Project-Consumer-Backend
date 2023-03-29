package routes

import (
	"app/controllers"
	"app/db"

	"github.com/gin-gonic/gin"
)

func gameRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	gameController := controllers.NewGameController(mongoDB)

	game := router.Group("/game")
	{
		game.GET("/upcoming", gameController.GetUpcomingGamesBySort)
		game.GET("", gameController.GetGamesByFilterAndSort)
	}
}
