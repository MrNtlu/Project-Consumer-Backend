package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func customListRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	customListController := controllers.NewCustomListController(mongoDB)

	customList := router.Group("/custom-list").Use(jwtToken.MiddlewareFunc())
	{
		customList.POST("", customListController.CreateCustomList)
		customList.PATCH("", customListController.UpdateCustomList)
		customList.PATCH("/add", customListController.UpdateAddContentToCustomList)
		customList.DELETE("/content", customListController.DeleteBulkContentFromCustomListByID)
		customList.DELETE("", customListController.DeleteCustomListByID)
	}

	customListOptional := router.Group("/custom-list").Use(helpers.OptionalTokenCheck)
	{
		customListOptional.GET("", customListController.GetCustomListsByUserID)
		customListOptional.GET("/details", customListController.GetCustomListDetails)
	}
}
