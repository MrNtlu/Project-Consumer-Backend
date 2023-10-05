package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func userListRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	userListController := controllers.NewUserListController(mongoDB)

	baseRoute := router.Group("/list")
	{
		userList := baseRoute.Use(jwtToken.MiddlewareFunc())
		{
			userList.DELETE("", userListController.DeleteListByUserIDAndType)
			userList.GET("", userListController.GetUserListByUserID)
			userList.GET("/logs", userListController.GetLogsByDateRange)
			userList.PATCH("", userListController.UpdateUserListPublicVisibility)
		}

		anime := baseRoute.Group("/anime").Use(jwtToken.MiddlewareFunc())
		{
			anime.POST("", userListController.CreateAnimeList)
			anime.PATCH("", userListController.UpdateAnimeListByID)
			anime.PATCH("/inc", userListController.IncrementAnimeListEpisodeByID)
		}

		game := baseRoute.Group("/game").Use(jwtToken.MiddlewareFunc())
		{
			game.POST("", userListController.CreateGameList)
			game.PATCH("", userListController.UpdateGameListByID)
			game.PATCH("/inc", userListController.IncrementGameListHourByID)
		}

		movie := baseRoute.Group("/movie").Use(jwtToken.MiddlewareFunc())
		{
			movie.POST("", userListController.CreateMovieWatchList)
			movie.PATCH("", userListController.UpdateMovieListByID)
		}

		tv := baseRoute.Group("/tv").Use(jwtToken.MiddlewareFunc())
		{
			tv.POST("", userListController.CreateTVSeriesWatchList)
			tv.PATCH("", userListController.UpdateTVSeriesListByID)
			tv.PATCH("/inc", userListController.IncrementTVSeriesListEpisodeSeasonByID)
		}
	}
}
