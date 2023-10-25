package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func reviewRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	reviewController := controllers.NewReviewController(mongoDB)

	review := router.Group("/review").Use(jwtToken.MiddlewareFunc())
	{
		review.POST("", reviewController.CreateReview)
		review.PATCH("", reviewController.UpdateReview)
		review.Use(helpers.OptionalTokenCheck).GET("", reviewController.GetReviewsByContentID)
		review.DELETE("", reviewController.DeleteReviewByID)
		review.PATCH("/vote", reviewController.VoteReview)
	}
}
