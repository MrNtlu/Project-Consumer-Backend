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

	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "All routes lead to rome"})
	})
}
