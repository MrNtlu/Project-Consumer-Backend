package routes

import (
	"app/controllers"
	"app/db"
	"app/helpers"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func recommendationRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	recommendationController := controllers.NewRecommendationController(mongoDB)

	recommendation := router.Group("/recommendation").Use(jwtToken.MiddlewareFunc())
	{
		recommendation.POST("", recommendationController.CreateRecommendation)
		recommendation.DELETE("", recommendationController.DeleteRecommendationByID)
		// recommendation.GET("/liked", recommendationController.GetLikedReviews)
		recommendation.GET("/profile", recommendationController.GetRecommendationsByUserID)
		// recommendation.PATCH("", recommendationController.UpdateReview)
		recommendation.PATCH("/like", recommendationController.LikeRecommendation)
	}

	recommendationOptional := router.Group("/recommendation").Use(helpers.OptionalTokenCheck)
	{
		recommendationOptional.GET("", recommendationController.GetRecommendationsByContentID)
		recommendationOptional.GET("/social", recommendationController.GetRecommendationsForSocial)
	}
}
