package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func socialRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	socialController := controllers.NewSocialController(mongoDB)

	preview := router.Group("/social").Use(helpers.OptionalTokenCheck)
	{
		preview.GET("", socialController.GetSocials)
	}
}
