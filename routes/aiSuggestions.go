package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/redis/go-redis/v9"
)

func aiSuggestionsRouter(
	router *gin.RouterGroup,
	jwtToken *jwt.GinJWTMiddleware,
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
	redisClient *redis.Client,
) {
	aiSuggestionsController := controllers.NewAISuggestionsController(mongoDB, pinecone, pineconeIndex, redisClient)

	suggestions := router.Group("/suggestions").Use(jwtToken.MiddlewareFunc())
	{
		suggestions.GET("", aiSuggestionsController.GenerateAISuggestions)
		suggestions.POST("/not-interested", aiSuggestionsController.NotInterested)
	}
}
