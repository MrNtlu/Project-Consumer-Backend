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
	AiringTVSeries    []responses.PreviewTVSeries
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

// Get Previews
// @Summary Get Previews
// @Description Returns previews
// @Tags preview
// @Accept application/json
// @Produce application/json
// @Param regionfilters body requests.RegionFilters true "Region Filters"
// @Success 200 {object} AnimeResult
// @Success 200 {object} MovieResult
// @Success 200 {object} GameResult
// @Success 200 {object} TVResult
// @Router /preview [get]
func (pr *PreviewController) GetHomePreview(c *gin.Context) {
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
			moviePopularPC, _ := movieModel.GetPopularProductionCompanies()
			movieDataCh <- MovieResult{MoviePopularPC: moviePopularPC}
		}()

		var result MovieResult
		for i := 0; i < 6; i++ {
			data := <-movieDataCh
			result.UpcomingMovies = append(result.UpcomingMovies, data.UpcomingMovies...)
			result.PopularMovies = append(result.PopularMovies, data.PopularMovies...)
			result.TopMovies = append(result.TopMovies, data.TopMovies...)
			result.PopularActors = append(result.PopularActors, data.PopularActors...)
			result.MoviesInTheater = append(result.MoviesInTheater, data.MoviesInTheater...)
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
			tvDataCh <- TVResult{AiringTVSeries: dayOfWeekTVSeries}
		}()

		go func() {
			tvPopularPC, _ := tvModel.GetPopularProductionCompanies()
			tvDataCh <- TVResult{TVPopularPC: tvPopularPC}
		}()

		var result TVResult
		for i := 0; i < 6; i++ {
			data := <-tvDataCh
			result.UpcomingTVSeries = append(result.UpcomingTVSeries, data.UpcomingTVSeries...)
			result.PopularTVSeries = append(result.PopularTVSeries, data.PopularTVSeries...)
			result.TopTVSeries = append(result.TopTVSeries, data.TopTVSeries...)
			result.PopularActors = append(result.PopularActors, data.PopularActors...)
			result.AiringTVSeries = append(result.AiringTVSeries, data.AiringTVSeries...)
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
			animePopularStudios, _ := animeModel.GetPopularStudios()
			animeDataCh <- AnimeResult{AnimePopularStudios: animePopularStudios}
		}()

		var result AnimeResult
		for i := 0; i < 5; i++ {
			data := <-animeDataCh
			result.UpcomingAnimes = append(result.UpcomingAnimes, data.UpcomingAnimes...)
			result.TopRatedAnimes = append(result.TopRatedAnimes, data.TopRatedAnimes...)
			result.PopularAnimes = append(result.PopularAnimes, data.PopularAnimes...)
			result.AiringAnime = append(result.AiringAnime, data.AiringAnime...)
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
			"production_companies": movieData.MoviePopularPC,
		},
		"tv": gin.H{
			"upcoming":             tvData.UpcomingTVSeries,
			"popular":              tvData.PopularTVSeries,
			"top":                  tvData.TopTVSeries,
			"extra":                tvData.AiringTVSeries,
			"actors":               tvData.PopularActors,
			"production_companies": tvData.TVPopularPC,
		},
		"anime": gin.H{
			"upcoming": animeData.UpcomingAnimes,
			"top":      animeData.TopRatedAnimes,
			"popular":  animeData.PopularAnimes,
			"extra":    animeData.AiringAnime,
			"studios":  animeData.AnimePopularStudios,
		},
		"game": gin.H{
			"upcoming": gameData.UpcomingGames,
			"top":      gameData.TopRatedGames,
			"popular":  gameData.PopularGames,
			"extra":    nil,
		},
	})
}

// Get Previews Older Version
// @Summary Get Previews Older Version
// @Description Returns previews older version
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
			movieDataCh <- MovieResult{MoviePopularSP: moviePopularSP}
		}()

		go func() {
			moviePopularPC, _ := movieModel.GetPopularProductionCompanies()
			movieDataCh <- MovieResult{MoviePopularPC: moviePopularPC}
		}()

		var result MovieResult
		for i := 0; i < 7; i++ {
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
			tvDataCh <- TVResult{AiringTVSeries: dayOfWeekTVSeries}
		}()

		go func() {
			tvSeriesPopularSP, _ := tvModel.GetPopularStreamingPlatforms(data.Region)
			tvDataCh <- TVResult{TVSeriesPopularSP: tvSeriesPopularSP}
		}()

		go func() {
			tvPopularPC, _ := tvModel.GetPopularProductionCompanies()
			tvDataCh <- TVResult{TVPopularPC: tvPopularPC}
		}()

		var result TVResult
		for i := 0; i < 7; i++ {
			data := <-tvDataCh
			result.UpcomingTVSeries = append(result.UpcomingTVSeries, data.UpcomingTVSeries...)
			result.PopularTVSeries = append(result.PopularTVSeries, data.PopularTVSeries...)
			result.TopTVSeries = append(result.TopTVSeries, data.TopTVSeries...)
			result.PopularActors = append(result.PopularActors, data.PopularActors...)
			result.AiringTVSeries = append(result.AiringTVSeries, data.AiringTVSeries...)
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
			animeDataCh <- AnimeResult{AnimePopularSP: animePopularSP}
		}()

		go func() {
			animePopularStudios, _ := animeModel.GetPopularStudios()
			animeDataCh <- AnimeResult{AnimePopularStudios: animePopularStudios}
		}()

		var result AnimeResult
		for i := 0; i < 6; i++ {
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
			"extra":                tvData.AiringTVSeries,
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
