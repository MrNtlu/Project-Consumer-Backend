package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func malImportRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	malImportController := controllers.NewMALImportController(mongoDB)

	importGroup := router.Group("/import")
	importGroup.Use(jwtToken.MiddlewareFunc())
	{
		importGroup.POST("/mal", malImportController.ImportFromMAL)
	}
}
