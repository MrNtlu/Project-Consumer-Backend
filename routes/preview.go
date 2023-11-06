package routes

import (
	"app/controllers"
	"app/db"

	"github.com/gin-gonic/gin"
)

func previewRouter(router *gin.RouterGroup, mongoDB *db.MongoDB) {
	previewController := controllers.NewPreviewController(mongoDB)

	preview := router.Group("/preview")
	{
		preview.GET("", previewController.GetHomePreview)
	}
}
