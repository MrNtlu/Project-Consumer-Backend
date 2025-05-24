package routes

import (
	"app/db"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func SetupRoutes(
	router *gin.Engine,
	jwtToken *jwt.GinJWTMiddleware,
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
) {
	apiRouter := router.Group("/api/v1")

	previewRouter(apiRouter, mongoDB)
	socialRouter(apiRouter, jwtToken, mongoDB)
	userRouter(apiRouter, jwtToken, mongoDB)
	tvRouter(apiRouter, mongoDB)
	movieRouter(apiRouter, mongoDB)
	animeRouter(apiRouter, mongoDB)
	mangaRouter(apiRouter, mongoDB)
	gameRouter(apiRouter, mongoDB)
	oauth2Router(apiRouter, jwtToken, mongoDB)
	userListRouter(apiRouter, jwtToken, mongoDB)
	userInteractionRouter(apiRouter, jwtToken, mongoDB)
	aiSuggestionsRouter(apiRouter, jwtToken, mongoDB, pinecone, pineconeIndex)
	reviewRouter(apiRouter, jwtToken, mongoDB)
	recommendationRouter(apiRouter, jwtToken, mongoDB)
	customListRouter(apiRouter, jwtToken, mongoDB)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "All routes lead to Rome 🏛️"})
	})
}
