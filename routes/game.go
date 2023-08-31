package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	"github.com/gin-gonic/gin"
)

func gameRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	gameController := controllers.NewGameController(mongoDB)

	game := router.Group("/game")
	{
		game.GET("/preview", gameController.GetPreviewGames)
		game.GET("/upcoming", gameController.GetUpcomingGamesBySort)
		game.GET("", gameController.GetGamesByFilterAndSort)
		game.Use(helpers.OptionalTokenCheck).GET("/details", gameController.GetGameDetails)
	}
}
