package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func recommendationRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	reviewController := controllers.NewRecommendationController(mongoDB)

	recommendation := router.Group("/recommendation").Use(jwtToken.MiddlewareFunc())
	{
		recommendation.POST("", reviewController.CreateRecommendation)
		// recommendation.GET("/liked", reviewController.GetLikedReviews)
		// recommendation.GET("/profile", reviewController.GetReviewsByUID)
		// recommendation.PATCH("", reviewController.UpdateReview)
		// recommendation.DELETE("", reviewController.DeleteReviewByID)
		// recommendation.PATCH("/like", reviewController.VoteReview)
	}

	// recommendationOptional := router.Group("/recommendation").Use(helpers.OptionalTokenCheck)
	// {
	// 	recommendationOptional.GET("", recommendationController.)
	// }
}
