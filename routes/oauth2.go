package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func oauth2Router(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	OAuth2Controller := controllers.NewOAuth2Controller(mongoDB)

	oauth := router.Group("/oauth")
	{
		oauth.POST("/google", OAuth2Controller.OAuth2GoogleLogin(jwtToken))
		// oauth.POST("/apple", OAuth2Controller.OAuth2AppleLogin(jwtToken))
	}
}
