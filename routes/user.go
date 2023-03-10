package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func userRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	userController := controllers.NewUserController(mongoDB)

	// router.GET("/confirm-password-reset", userController.ConfirmPasswordReset)

	auth := router.Group("/auth")
	{
		auth.POST("/login", jwtToken.LoginHandler)
		auth.POST("/register", userController.Register)
		auth.POST("/logout", jwtToken.LogoutHandler)
		auth.GET("/refresh", jwtToken.RefreshHandler)

		//TODO: Implement confirm password
		//https://github.com/MrNtlu/Asset-Manager/blob/master/controllers/user.go
		// 	auth.GET("/confirm-password-reset", userController.ConfirmPasswordReset)
	}

	user := router.Group("/user")
	{
		user.POST("/forgot-password", userController.ForgotPassword)

		user.Use(jwtToken.MiddlewareFunc())
		{
			user.GET("/info", userController.GetUserInfo)
			user.DELETE("/delete", userController.DeleteUser)
			user.PATCH("/password", userController.ChangePassword)
			user.PATCH("/notification", userController.ChangeNotificationPreference)
			user.PATCH("/token", userController.UpdateFCMToken)
			user.PATCH("/membership", userController.ChangeUserMembership)
		}
	}
}
