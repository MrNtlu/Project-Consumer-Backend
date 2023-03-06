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
		// 	auth.GET("/confirm-password-reset", userController.ConfirmPasswordReset)
	}

	// user := router.Group("/user")
	// {
	// 	user.POST("/forgot-password", userController.ForgotPassword)

	// 	user.Use(jwtToken.MiddlewareFunc())
	// 	{
	// 		user.GET("/info", userController.GetUserInfo)
	// 		user.DELETE("", userController.DeleteUser)
	// 		user.PUT("/change-password", userController.ChangePassword)
	// 		user.PUT("/change-currency", userController.ChangeCurrency)
	// 		user.PUT("/change-notification", userController.ChangeNotificationPreference)
	// 		user.PUT("/update-token", userController.UpdateFCMToken)
	// 		user.PUT("/membership", userController.ChangeUserMembership)
	// 	}
	// }
}
