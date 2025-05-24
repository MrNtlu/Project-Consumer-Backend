package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func aiSuggestionsRouter(
	router *gin.RouterGroup,
	jwtToken *jwt.GinJWTMiddleware,
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
) {
	aiSuggestionsController := controllers.NewAISuggestionsController(mongoDB, pinecone, pineconeIndex)

	suggestions := router.Group("/suggestions").Use(jwtToken.MiddlewareFunc())
	{
		suggestions.GET("", aiSuggestionsController.GenerateAISuggestions)
	}
}
