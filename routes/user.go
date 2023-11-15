package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func userRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	userController := controllers.NewUserController(mongoDB)

	router.GET("/confirm-password-reset", userController.ConfirmPasswordReset)

	auth := router.Group("/auth")
	{
		auth.POST("/login", jwtToken.LoginHandler)
		auth.POST("/register", userController.Register)
		auth.POST("/logout", jwtToken.LogoutHandler)
		auth.GET("/refresh", jwtToken.RefreshHandler)

		auth.GET("/confirm-password-reset", userController.ConfirmPasswordReset)
	}

	user := router.Group("/user")
	{
		user.POST("/forgot-password", userController.ForgotPassword)

		user.Use(jwtToken.MiddlewareFunc())
		{
			user.GET("/basic", userController.GetBasicUserInfo)
			user.GET("/info", userController.GetUserInfo)
			user.GET("/profile", userController.GetUserInfoFromUsername)
			user.GET("/requests", userController.GetFriendRequests)
			user.DELETE("/delete", userController.DeleteUser)
			user.PATCH("/password", userController.ChangePassword)
			user.PATCH("/image", userController.UpdateUser)
			user.PATCH("/notification", userController.ChangeNotificationPreference)
			user.PATCH("/token", userController.UpdateFCMToken)
			user.PATCH("/membership", userController.ChangeUserMembership)
			user.PATCH("/username", userController.ChangeUsername)
			user.POST("/friend", userController.SendFriendRequest)
		}
	}
}
