package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"
	"sync"
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
		upcomingMovies = []responses.PreviewMovie{}
	}

	popularMovies, err := movieModel.GetPopularPreviewMovies()
	if err != nil {
		popularMovies = []responses.PreviewMovie{}
	}

	topMovies, err := movieModel.GetTopPreviewMovies()
	if err != nil {
		topMovies = []responses.PreviewMovie{}
	}

	popularActors, err := movieModel.GetPopularActors(requests.Pagination{Page: 1})
	if err != nil {
		popularActors = []responses.ActorDetails{}
	}

	moviesInTheater, err := movieModel.GetInTheaterPreviewMovies()
	if err != nil {
		moviesInTheater = []responses.PreviewMovie{}
	}

	// TV Series

	upcomingTVSeries, err := tvModel.GetUpcomingPreviewTVSeries()
	if err != nil {
		upcomingTVSeries = []responses.PreviewTVSeries{}
	}

	popularTVSeries, err := tvModel.GetPopularPreviewTVSeries()
	if err != nil {
		popularTVSeries = []responses.PreviewTVSeries{}
	}

	topTVSeries, err := tvModel.GetTopPreviewTVSeries()
	if err != nil {
		topTVSeries = []responses.PreviewTVSeries{}
	}

	popularActorsTVSeries, err := tvModel.GetPopularActors(requests.Pagination{Page: 1})
	if err != nil {
		popularActorsTVSeries = []responses.ActorDetails{}
	}

	dayOfWeek := int16(time.Now().UTC().Weekday()) + 1
	dayOfWeekTVSeries, err := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek(dayOfWeek)
	if err != nil {
		dayOfWeekTVSeries = []responses.PreviewTVSeries{}
	}

	// Anime

	upcomingAnimes, err := animeModel.GetPreviewUpcomingAnimes()
	if err != nil {
		upcomingAnimes = []responses.PreviewAnime{}
	}

	topRatedAnimes, err := animeModel.GetPreviewTopAnimes()
	if err != nil {
		topRatedAnimes = []responses.PreviewAnime{}
	}

	popularAnimes, err := animeModel.GetPreviewPopularAnimes()
	if err != nil {
		popularAnimes = []responses.PreviewAnime{}
	}

	dayOfWeekAnime, err := animeModel.GetCurrentlyAiringAnimesByDayOfWeek(dayOfWeek)
	if err != nil {
		dayOfWeekAnime = []responses.PreviewAnime{}
	}

	// Game

	upcomingGames, err := gameModel.GetPreviewUpcomingGames()
	if err != nil {
		upcomingGames = []responses.PreviewGame{}
	}

	topRatedGames, err := gameModel.GetPreviewTopGames()
	if err != nil {
		topRatedGames = []responses.PreviewGame{}
	}

	popularGames, err := gameModel.GetPreviewPopularGames()
	if err != nil {
		popularGames = []responses.PreviewGame{}
	}

	// Manga

	publishingManga, err := mangaModel.GetPreviewCurrentlyPublishingManga()
	if err != nil {
		publishingManga = []responses.PreviewManga{}
	}

	topRatedManga, err := mangaModel.GetPreviewTopManga()
	if err != nil {
		topRatedManga = []responses.PreviewManga{}
	}

	popularManga, err := mangaModel.GetPreviewPopularManga()
	if err != nil {
		popularManga = []responses.PreviewManga{}
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
	// mangaModel := models.NewMangaModel(pr.Database)

	// Channels for receiving results
	movieCh := make(chan MovieResult)
	tvCh := make(chan TVResult)
	animeCh := make(chan AnimeResult)
	gameCh := make(chan GameResult)
	// mangaCh := make(chan MangaResult)

	// WaitGroup to wait for all Goroutines to finish
	var wg sync.WaitGroup
	wg.Add(4) // Make 5 when Manga added.

	// Movie
	go func() {
		defer wg.Done()
		movieDataCh := make(chan MovieResult)

		go func() {
			upcomingMovies, _ := movieModel.GetUpcomingPreviewMovies()
			movieDataCh <- MovieResult{UpcomingMovies: upcomingMovies}
		}()

		go func() {
			popularMovies, _ := movieModel.GetPopularPreviewMovies()
			movieDataCh <- MovieResult{PopularMovies: popularMovies}
		}()

		go func() {
			topMovies, _ := movieModel.GetTopPreviewMovies()
			movieDataCh <- MovieResult{TopMovies: topMovies}
		}()

		go func() {
			popularActors, _ := movieModel.GetPopularActors(requests.Pagination{Page: 1})
			movieDataCh <- MovieResult{PopularActors: popularActors}
		}()

		go func() {
			moviesInTheater, _ := movieModel.GetInTheaterPreviewMovies()
			movieDataCh <- MovieResult{MoviesInTheater: moviesInTheater}
		}()

		go func() {
			moviePopularSP, _ := movieModel.GetPopularStreamingPlatforms(data.Region)
			moviePopularPC, _ := movieModel.GetPopularProductionCompanies()
			movieDataCh <- MovieResult{MoviePopularSP: moviePopularSP, MoviePopularPC: moviePopularPC}
		}()

		var result MovieResult
		for i := 0; i < 6; i++ {
			data := <-movieDataCh
			result.UpcomingMovies = append(result.UpcomingMovies, data.UpcomingMovies...)
			result.PopularMovies = append(result.PopularMovies, data.PopularMovies...)
			result.TopMovies = append(result.TopMovies, data.TopMovies...)
			result.PopularActors = append(result.PopularActors, data.PopularActors...)
			result.MoviesInTheater = append(result.MoviesInTheater, data.MoviesInTheater...)
			result.MoviePopularSP = append(result.MoviePopularSP, data.MoviePopularSP...)
			result.MoviePopularPC = append(result.MoviePopularPC, data.MoviePopularPC...)
		}

		movieCh <- result
	}()

	// TV Series
	go func() {
		defer wg.Done()
		tvDataCh := make(chan TVResult)

		go func() {
			upcomingTVSeries, _ := tvModel.GetUpcomingPreviewTVSeries()
			tvDataCh <- TVResult{UpcomingTVSeries: upcomingTVSeries}
		}()

		go func() {
			popularTVSeries, _ := tvModel.GetPopularPreviewTVSeries()
			tvDataCh <- TVResult{PopularTVSeries: popularTVSeries}
		}()

		go func() {
			topTVSeries, _ := tvModel.GetTopPreviewTVSeries()
			tvDataCh <- TVResult{TopTVSeries: topTVSeries}
		}()

		go func() {
			popularActorsTVSeries, _ := tvModel.GetPopularActors(requests.Pagination{Page: 1})
			tvDataCh <- TVResult{PopularActors: popularActorsTVSeries}
		}()

		go func() {
			dayOfWeekTVSeries, _ := tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek(int16(time.Now().UTC().Weekday()) + 1)
			tvDataCh <- TVResult{airingTVSeries: dayOfWeekTVSeries}
		}()

		go func() {
			tvSeriesPopularSP, _ := tvModel.GetPopularStreamingPlatforms(data.Region)
			tvPopularPC, _ := tvModel.GetPopularProductionCompanies()
			tvDataCh <- TVResult{TVSeriesPopularSP: tvSeriesPopularSP, TVPopularPC: tvPopularPC}
		}()

		var result TVResult
		for i := 0; i < 6; i++ {
			data := <-tvDataCh
			result.UpcomingTVSeries = append(result.UpcomingTVSeries, data.UpcomingTVSeries...)
			result.PopularTVSeries = append(result.PopularTVSeries, data.PopularTVSeries...)
			result.TopTVSeries = append(result.TopTVSeries, data.TopTVSeries...)
			result.PopularActors = append(result.PopularActors, data.PopularActors...)
			result.airingTVSeries = append(result.airingTVSeries, data.airingTVSeries...)
			result.TVSeriesPopularSP = append(result.TVSeriesPopularSP, data.TVSeriesPopularSP...)
			result.TVPopularPC = append(result.TVPopularPC, data.TVPopularPC...)
		}

		tvCh <- result
	}()

	// Anime
	go func() {
		defer wg.Done()
		animeDataCh := make(chan AnimeResult)

		go func() {
			upcomingAnimes, _ := animeModel.GetPreviewUpcomingAnimes()
			animeDataCh <- AnimeResult{UpcomingAnimes: upcomingAnimes}
		}()

		go func() {
			topRatedAnimes, _ := animeModel.GetPreviewTopAnimes()
			animeDataCh <- AnimeResult{TopRatedAnimes: topRatedAnimes}
		}()

		go func() {
			popularAnimes, _ := animeModel.GetPreviewPopularAnimes()
			animeDataCh <- AnimeResult{PopularAnimes: popularAnimes}
		}()

		go func() {
			dayOfWeekAnime, _ := animeModel.GetCurrentlyAiringAnimesByDayOfWeek(int16(time.Now().UTC().Weekday()) + 1)
			animeDataCh <- AnimeResult{AiringAnime: dayOfWeekAnime}
		}()

		go func() {
			animePopularSP, _ := animeModel.GetPopularStreamingPlatforms()
			animePopularStudios, _ := animeModel.GetPopularStudios()
			animeDataCh <- AnimeResult{AnimePopularSP: animePopularSP, AnimePopularStudios: animePopularStudios}
		}()

		var result AnimeResult
		for i := 0; i < 5; i++ {
			data := <-animeDataCh
			result.UpcomingAnimes = append(result.UpcomingAnimes, data.UpcomingAnimes...)
			result.TopRatedAnimes = append(result.TopRatedAnimes, data.TopRatedAnimes...)
			result.PopularAnimes = append(result.PopularAnimes, data.PopularAnimes...)
			result.AiringAnime = append(result.AiringAnime, data.AiringAnime...)
			result.AnimePopularSP = append(result.AnimePopularSP, data.AnimePopularSP...)
			result.AnimePopularStudios = append(result.AnimePopularStudios, data.AnimePopularStudios...)
		}

		animeCh <- result
	}()

	// Game
	go func() {
		defer wg.Done()
		gameDataCh := make(chan GameResult)

		go func() {
			upcomingGames, _ := gameModel.GetPreviewUpcomingGames()
			gameDataCh <- GameResult{UpcomingGames: upcomingGames}
		}()

		go func() {
			topRatedGames, _ := gameModel.GetPreviewTopGames()
			gameDataCh <- GameResult{TopRatedGames: topRatedGames}
		}()

		go func() {
			popularGames, _ := gameModel.GetPreviewPopularGames()
			gameDataCh <- GameResult{PopularGames: popularGames}
		}()

		var result GameResult
		for i := 0; i < 3; i++ {
			data := <-gameDataCh
			result.UpcomingGames = append(result.UpcomingGames, data.UpcomingGames...)
			result.TopRatedGames = append(result.TopRatedGames, data.TopRatedGames...)
			result.PopularGames = append(result.PopularGames, data.PopularGames...)
		}

		gameCh <- result
	}()

	movieData := <-movieCh
	tvData := <-tvCh
	animeData := <-animeCh
	gameData := <-gameCh

	c.JSON(http.StatusOK, gin.H{
		"movie": gin.H{
			"upcoming":             movieData.UpcomingMovies,
			"popular":              movieData.PopularMovies,
			"top":                  movieData.TopMovies,
			"extra":                movieData.MoviesInTheater,
			"actors":               movieData.PopularActors,
			"streaming_platforms":  movieData.MoviePopularSP,
			"production_companies": movieData.MoviePopularPC,
		},
		"tv": gin.H{
			"upcoming":             tvData.UpcomingTVSeries,
			"popular":              tvData.PopularTVSeries,
			"top":                  tvData.TopTVSeries,
			"extra":                tvData.airingTVSeries,
			"actors":               tvData.PopularActors,
			"streaming_platforms":  tvData.TVSeriesPopularSP,
			"production_companies": tvData.TVPopularPC,
		},
		"anime": gin.H{
			"upcoming":                  animeData.UpcomingAnimes,
			"top":                       animeData.TopRatedAnimes,
			"popular":                   animeData.PopularAnimes,
			"extra":                     animeData.AiringAnime,
			"anime_streaming_platforms": animeData.AnimePopularSP,
			"studios":                   animeData.AnimePopularStudios,
		},
		"game": gin.H{
			"upcoming": gameData.UpcomingGames,
			"top":      gameData.TopRatedGames,
			"popular":  gameData.PopularGames,
			"extra":    nil,
		},
	})
}

type MovieResult struct {
	UpcomingMovies  []responses.PreviewMovie
	PopularMovies   []responses.PreviewMovie
	TopMovies       []responses.PreviewMovie
	PopularActors   []responses.ActorDetails
	MoviesInTheater []responses.PreviewMovie
	MoviePopularSP  []responses.StreamingPlatform
	MoviePopularPC  []responses.StreamingPlatform
}

type TVResult struct {
	UpcomingTVSeries  []responses.PreviewTVSeries
	PopularTVSeries   []responses.PreviewTVSeries
	TopTVSeries       []responses.PreviewTVSeries
	PopularActors     []responses.ActorDetails
	airingTVSeries    []responses.PreviewTVSeries
	TVSeriesPopularSP []responses.StreamingPlatform
	TVPopularPC       []responses.StreamingPlatform
}

type AnimeResult struct {
	UpcomingAnimes      []responses.PreviewAnime
	TopRatedAnimes      []responses.PreviewAnime
	PopularAnimes       []responses.PreviewAnime
	AiringAnime         []responses.PreviewAnime
	AnimePopularSP      []responses.AnimeNameURL
	AnimePopularStudios []responses.AnimeNameURL
}

type GameResult struct {
	UpcomingGames []responses.PreviewGame
	TopRatedGames []responses.PreviewGame
	PopularGames  []responses.PreviewGame
}
