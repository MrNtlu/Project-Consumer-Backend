package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func userInteractionRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	userInteractionController := controllers.NewUserInteractionController(mongoDB)

	consume := router.Group("/consume").Use(jwtToken.MiddlewareFunc())
	{
		consume.POST("/move", userInteractionController.MarkConsumeLaterAsUserList)
		consume.POST("", userInteractionController.CreateConsumeLater)
		consume.GET("", userInteractionController.GetConsumeLater)
		consume.DELETE("", userInteractionController.DeleteConsumeLaterById)
		consume.DELETE("/all", userInteractionController.DeleteAllConsumeLaterByUserID)
	}
}
