package routes

import (
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func userListRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	// userListController := controllers.NewUserListController(mongoDB)

	// userList := router.Group("/list").Use(jwtToken.MiddlewareFunc())
	// {

	// }
}
