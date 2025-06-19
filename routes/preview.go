package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	"github.com/gin-gonic/gin"
)

func previewRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	previewController := controllers.NewPreviewController(mongoDB)

	preview := router.Group("/preview")
	{
		preview.GET("", helpers.OptionalTokenCheck, previewController.GetHomePreview)
		preview.GET("/v2", previewController.GetHomePreviewV2)
	}
}
