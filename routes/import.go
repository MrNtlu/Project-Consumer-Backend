package routes

import (
	"app/controllers"
	"app/db"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func importRouter(router *gin.RouterGroup, jwtToken *jwt.GinJWTMiddleware, mongoDB *db.MongoDB) {
	malImportController := controllers.NewMALImportController(mongoDB)
	steamImportController := controllers.NewSteamImportController(mongoDB)
	imdbImportController := controllers.NewIMDBImportController(mongoDB)
	tmdbImportController := controllers.NewTMDBImportController(mongoDB)
	anilistImportController := controllers.NewAniListImportController(mongoDB)
	traktImportController := controllers.NewTraktImportController(mongoDB)

	importGroup := router.Group("/import")
	importGroup.Use(jwtToken.MiddlewareFunc())
	{
		importGroup.POST("/mal", malImportController.ImportFromMAL)
		importGroup.POST("/steam", steamImportController.ImportFromSteam)
		importGroup.POST("/imdb", imdbImportController.ImportWatchlist)
		importGroup.POST("/tmdb", tmdbImportController.ImportUserData)
		importGroup.POST("/anilist", anilistImportController.ImportUserLists)
		importGroup.POST("/trakt", traktImportController.ImportUserData)
	}
}
