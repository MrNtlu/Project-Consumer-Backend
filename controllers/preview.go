package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
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

// Get Previews
// @Summary Get Previews
// @Description Returns previews
// @Tags preview
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.PreviewMovie
// @Success 200 {array} responses.PreviewAnime
// @Success 200 {array} responses.PreviewTVSeries
// @Success 200 {array} responses.PreviewGame
// @Success 200 {array} responses.PreviewManga
// @Success 200 {array} responses.ActorDetails
// @Failure 500 {string} string
// @Router /preview [get]
func (pr *PreviewController) GetHomePreview(c *gin.Context) {
	movieModel := models.NewMovieModel(pr.Database)
	tvModel := models.NewTVModel(pr.Database)
	animeModel := models.NewAnimeModel(pr.Database)
	gameModel := models.NewGameModel(pr.Database)
	mangaModel := models.NewMangaModel(pr.Database)

	upcomingMovies, err := movieModel.GetUpcomingPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularMovies, err := movieModel.GetPopularPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topMovies, err := movieModel.GetTopPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularActors, err := movieModel.GetPopularActors(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	moviesInTheater, err := movieModel.GetInTheaterPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// TV Series

	upcomingTVSeries, err := tvModel.GetUpcomingPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularTVSeries, err := tvModel.GetPopularPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topTVSeries, err := tvModel.GetTopPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularActorsTVSeries, err := tvModel.GetPopularActors(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeek := int16(time.Now().UTC().Weekday()) + 1
	dayOfWeekTVSeries, err := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek(dayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Anime

	upcomingAnimes, err := animeModel.GetPreviewUpcomingAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedAnimes, err := animeModel.GetPreviewTopAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularAnimes, err := animeModel.GetPreviewPopularAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeekAnime, err := animeModel.GetCurrentlyAiringAnimesByDayOfWeek(dayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Game

	upcomingGames, err := gameModel.GetPreviewUpcomingGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedGames, err := gameModel.GetPreviewTopGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularGames, err := gameModel.GetPreviewPopularGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Manga

	publishingManga, err := mangaModel.GetPreviewCurrentlyPublishingManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedManga, err := mangaModel.GetPreviewTopManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularManga, err := mangaModel.GetPreviewPopularManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movie": gin.H{"upcoming": upcomingMovies, "popular": popularMovies, "top": topMovies, "extra": moviesInTheater, "actors": popularActors},
		"tv":    gin.H{"upcoming": upcomingTVSeries, "popular": popularTVSeries, "top": topTVSeries, "extra": dayOfWeekTVSeries, "actors": popularActorsTVSeries},
		"anime": gin.H{"upcoming": upcomingAnimes, "top": topRatedAnimes, "popular": popularAnimes, "extra": dayOfWeekAnime},
		"game":  gin.H{"upcoming": upcomingGames, "top": topRatedGames, "popular": popularGames, "extra": nil},
		"manga": gin.H{"upcoming": publishingManga, "top": topRatedManga, "popular": popularManga, "extra": nil},
	})
}

// Get Previews
// @Summary Get Previews
// @Description Returns previews
// @Tags preview
// @Accept application/json
// @Produce application/json
// @Param regionfilters body requests.RegionFilters true "Region Filters"
// @Success 200 {array} responses.PreviewMovie
// @Success 200 {array} responses.PreviewAnime
// @Success 200 {array} responses.PreviewTVSeries
// @Success 200 {array} responses.PreviewGame
// @Success 200 {array} responses.PreviewManga
// @Success 200 {array} responses.ActorDetails
// @Success 200 {array} responses.StreamingPlatform
// @Success 200 {array} responses.AnimeNameURL
// @Failure 500 {string} string
// @Router /preview/v2 [get]
func (pr *PreviewController) GetHomePreviewV2(c *gin.Context) {
	var data requests.RegionFilters
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(pr.Database)
	tvModel := models.NewTVModel(pr.Database)
	animeModel := models.NewAnimeModel(pr.Database)
	gameModel := models.NewGameModel(pr.Database)
	mangaModel := models.NewMangaModel(pr.Database)

	upcomingMovies, err := movieModel.GetUpcomingPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularMovies, err := movieModel.GetPopularPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topMovies, err := movieModel.GetTopPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularActors, err := movieModel.GetPopularActors(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	moviesInTheater, err := movieModel.GetInTheaterPreviewMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	moviePopularSP, _ := movieModel.GetPopularStreamingPlatforms(data.Region)

	// TV Series

	upcomingTVSeries, err := tvModel.GetUpcomingPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularTVSeries, err := tvModel.GetPopularPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topTVSeries, err := tvModel.GetTopPreviewTVSeries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularActorsTVSeries, err := tvModel.GetPopularActors(requests.Pagination{Page: 1})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeek := int16(time.Now().UTC().Weekday()) + 1
	dayOfWeekTVSeries, err := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek(dayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	tvSeriesPopularSP, _ := tvModel.GetPopularStreamingPlatforms(data.Region)

	// Anime

	upcomingAnimes, err := animeModel.GetPreviewUpcomingAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedAnimes, err := animeModel.GetPreviewTopAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularAnimes, err := animeModel.GetPreviewPopularAnimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	dayOfWeekAnime, err := animeModel.GetCurrentlyAiringAnimesByDayOfWeek(dayOfWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	animePopularSP, _ := animeModel.GetPopularStreamingPlatforms()

	animePopularStudios, _ := animeModel.GetPopularStudios()

	// Game

	upcomingGames, err := gameModel.GetPreviewUpcomingGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedGames, err := gameModel.GetPreviewTopGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularGames, err := gameModel.GetPreviewPopularGames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Manga

	publishingManga, err := mangaModel.GetPreviewCurrentlyPublishingManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topRatedManga, err := mangaModel.GetPreviewTopManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularManga, err := mangaModel.GetPreviewPopularManga()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movie": gin.H{
			"upcoming": upcomingMovies, "popular": popularMovies, "top": topMovies,
			"extra": moviesInTheater, "actors": popularActors, "streaming_platforms": moviePopularSP,
		},
		"tv": gin.H{
			"upcoming": upcomingTVSeries, "popular": popularTVSeries, "top": topTVSeries,
			"extra": dayOfWeekTVSeries, "actors": popularActorsTVSeries, "streaming_platforms": tvSeriesPopularSP,
		},
		"anime": gin.H{
			"upcoming": upcomingAnimes, "top": topRatedAnimes, "popular": popularAnimes,
			"extra": dayOfWeekAnime, "anime_streaming_platforms": animePopularSP, "studios": animePopularStudios,
		},
		"game":  gin.H{"upcoming": upcomingGames, "top": topRatedGames, "popular": popularGames, "extra": nil},
		"manga": gin.H{"upcoming": publishingManga, "top": topRatedManga, "popular": popularManga, "extra": nil},
	})
}
