package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func achievementRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	achievementController := controllers.NewAchievementController(mongoDB)

	// Public achievements endpoint
	router.GET("/achievements/all", achievementController.GetAllAchievements)

	// Authenticated achievements endpoint
	achievement := router.Group("/achievements")
	achievement.Use(jwtToken.MiddlewareFunc())
	{
		achievement.GET("", achievementController.GetUserAchievements)
	}
}
