package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func aiSuggestionsRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	aiSuggestionsController := controllers.NewAISuggestionsController(mongoDB)

	suggestions := router.Group("/suggestions").Use(jwtToken.MiddlewareFunc())
	{
		suggestions.GET("", aiSuggestionsController.GetAISuggestions)
	}
}
