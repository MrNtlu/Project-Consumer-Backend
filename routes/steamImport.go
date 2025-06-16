package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func steamImportRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	steamImportController := controllers.NewSteamImportController(mongoDB)

	importGroup := router.Group("/import")
	importGroup.Use(jwtToken.MiddlewareFunc())
	{
		importGroup.POST("/steam", steamImportController.ImportFromSteam)
	}
}
