package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func imdbImportRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	imdbImportController := controllers.NewIMDBImportController(mongoDB)

	imdbImportGroup := router.Group("/import")
	imdbImportGroup.Use(jwtToken.MiddlewareFunc())
	{
		imdbImportGroup.POST("/imdb", imdbImportController.ImportWatchlist)
	}
}
