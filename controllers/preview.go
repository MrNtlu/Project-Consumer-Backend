package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PreviewController struct {
	Database *db.MongoDB
}

func NewPreviewController(mongoDB *db.MongoDB) PreviewController {
	return PreviewController{
		Database: mongoDB,
	}
}

func (pr *PreviewController) GetHomePreview(c *gin.Context) {
	movieModel := models.NewMovieModel(pr.Database)
	tvModel := models.NewTVModel(pr.Database)
	animeModel := models.NewAnimeModel(pr.Database)
	gameModel := models.NewGameModel(pr.Database)

	upcomingMovies, _, err := movieModel.GetUpcomingMoviesBySort(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularMovies, _, err := movieModel.GetMoviesBySortAndFilter(requests.SortFilterMovie{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topMovies, _, err := movieModel.GetMoviesBySortAndFilter(requests.SortFilterMovie{
		Sort: "top",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	moviesInTheater, _, err := movieModel.GetMoviesInTheater(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// TV Series

	upcomingTVSeries, _, err := tvModel.GetUpcomingTVSeries(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularTVSeries, _, err := tvModel.GetTVSeriesBySortAndFilter(requests.SortFilterTVSeries{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topTVSeries, _, err := tvModel.GetTVSeriesBySortAndFilter(requests.SortFilterTVSeries{
		Sort: "top",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeekTVSeries, err := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeek := int16(time.Now().UTC().Weekday()) + 1

	var dayOfWeekTVSeriesList responses.DayOfWeekTVSeries
	for _, item := range dayOfWeekTVSeries {
		if item.DayOfWeek == dayOfWeek {
			dayOfWeekTVSeriesList = item
		}
	}

	// Anime

	upcomingAnimes, _, err := animeModel.GetUpcomingAnimesBySort(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedAnimes, _, err := animeModel.GetAnimesBySortAndFilter(requests.SortFilterAnime{
		Sort: "top",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularAnimes, _, err := animeModel.GetAnimesBySortAndFilter(requests.SortFilterAnime{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeekAnime, err := animeModel.GetCurrentlyAiringAnimesByDayOfWeek()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	var dayOfWeekAnimeList responses.DayOfWeekAnime
	for _, item := range dayOfWeekAnime {
		if item.DayOfWeek == dayOfWeek {
			dayOfWeekAnimeList = item
		}
	}

	// Game

	upcomingGames, _, err := gameModel.GetUpcomingGamesBySort(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedGames, _, err := gameModel.GetGamesByFilterAndSort(requests.SortFilterGame{
		Sort: "top",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularGames, _, err := gameModel.GetGamesByFilterAndSort(requests.SortFilterGame{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movie": gin.H{"upcoming": upcomingMovies, "popular": popularMovies, "top": topMovies, "extra": moviesInTheater},
		"tv":    gin.H{"upcoming": upcomingTVSeries, "popular": popularTVSeries, "top": topTVSeries, "extra": dayOfWeekTVSeriesList.Data},
		"anime": gin.H{"upcoming": upcomingAnimes, "top": topRatedAnimes, "popular": popularAnimes, "extra": dayOfWeekAnimeList.Data},
		"game":  gin.H{"upcoming": upcomingGames, "top": topRatedGames, "popular": popularGames, "extra": nil},
	})

}
