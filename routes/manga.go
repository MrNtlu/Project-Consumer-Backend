package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	"github.com/gin-gonic/gin"
)

func mangaRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	mangaController := controllers.NewMangaController(mongoDB)

	manga := router.Group("/manga")
	{
		manga.GET("", mangaController.GetMangaBySortAndFilter)
		manga.Use(helpers.OptionalTokenCheck).GET("/details", mangaController.GetMangaDetails)
	}
}
