package routes

import (
	"app/db"

	"github.com/gin-gonic/gin"
)

func comicRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	// comicController := controllers.NewComicController(mongoDB)

	// comic := router.Group("/comic")
	// {
	// 	comic.GET("/upcoming", comicController.GetUpcomingMoviesBySort)
	// 	comic.GET("", comicController.GetMoviesBySortAndFilter)
	// 	comic.Use(helpers.OptionalTokenCheck).GET("/details", comicController.GetMovieDetails)
	// }
}
