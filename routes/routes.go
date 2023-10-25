package routes

import (
	"app/db"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	apiRouter := router.Group("/api/v1")

	userRouter(apiRouter, jwtToken, mongoDB)
	tvRouter(apiRouter, mongoDB)
	movieRouter(apiRouter, mongoDB)
	animeRouter(apiRouter, mongoDB)
	gameRouter(apiRouter, mongoDB)
	oauth2Router(apiRouter, jwtToken, mongoDB)
	userListRouter(apiRouter, jwtToken, mongoDB)
	userInteractionRouter(apiRouter, jwtToken, mongoDB)
	aiSuggestionsRouter(apiRouter, jwtToken, mongoDB)
	reviewRouter(apiRouter, jwtToken, mongoDB)

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "All routes lead to Rome 🏛️"})
	})
}
