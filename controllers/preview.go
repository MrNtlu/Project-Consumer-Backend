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

type previewResult struct {
	MovieData MovieResult
	TVData    TVResult
	AnimeData AnimeResult
	GameData  GameResult
}

const PreviewLimit = 30

// Get Previews
// @Summary Get Previews
// @Description Returns previews with optional consume later items if user is authenticated
// @Tags preview
// @Accept application/json
// @Produce application/json
// @Param regionfilters body requests.RegionFilters true "Region Filters"
// @Param Authorization header string false "Optional Authentication header for consume later items"
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

	// Check if user is logged in via optional token
	uid, userExists := c.Get("uuid")
	var userID string
	if userExists {
		userID = uid.(string)
	}

	// Initialize models once
	movieModel := models.NewMovieModel(pr.Database)
	tvModel := models.NewTVModel(pr.Database)
	animeModel := models.NewAnimeModel(pr.Database)
	gameModel := models.NewGameModel(pr.Database)
	var userInteractionModel *models.UserInteractionModel
	if userExists {
		userInteractionModel = models.NewUserInteractionModel(pr.Database)
	}

	var wg sync.WaitGroup

	// Pre-allocate result structure with optimized capacities
	result := previewResult{
		MovieData: MovieResult{
			UpcomingMovies:  make([]responses.PreviewMovie, 0, PreviewLimit),
			PopularMovies:   make([]responses.PreviewMovie, 0, PreviewLimit),
			TopMovies:       make([]responses.PreviewMovie, 0, PreviewLimit),
			PopularActors:   make([]responses.ActorDetails, 0, PreviewLimit),
			MoviesInTheater: make([]responses.PreviewMovie, 0, PreviewLimit),
			MoviePopularPC:  make([]responses.StreamingPlatform, 0, 15),
		},
		TVData: TVResult{
			UpcomingTVSeries: make([]responses.PreviewTVSeries, 0, PreviewLimit),
			PopularTVSeries:  make([]responses.PreviewTVSeries, 0, PreviewLimit),
			TopTVSeries:      make([]responses.PreviewTVSeries, 0, PreviewLimit),
			PopularActors:    make([]responses.ActorDetails, 0, PreviewLimit),
			AiringTVSeries:   make([]responses.PreviewTVSeries, 0, PreviewLimit),
			TVPopularPC:      make([]responses.StreamingPlatform, 0, 15),
		},
		AnimeData: AnimeResult{
			UpcomingAnimes:      make([]responses.PreviewAnime, 0, PreviewLimit),
			TopRatedAnimes:      make([]responses.PreviewAnime, 0, PreviewLimit),
			PopularAnimes:       make([]responses.PreviewAnime, 0, PreviewLimit),
			AiringAnime:         make([]responses.PreviewAnime, 0, PreviewLimit),
			AnimePopularStudios: make([]responses.AnimeNameURL, 0, 15),
		},
		GameData: GameResult{
			UpcomingGames: make([]responses.PreviewGame, 0, PreviewLimit),
			TopRatedGames: make([]responses.PreviewGame, 0, PreviewLimit),
			PopularGames:  make([]responses.PreviewGame, 0, PreviewLimit),
		},
	}

	// Store consume later lists - pre-allocate with expected size
	var (
		movieConsumeLaters []responses.ConsumeLaterPreview
		tvConsumeLaters    []responses.ConsumeLaterPreview
		animeConsumeLaters []responses.ConsumeLaterPreview
		gameConsumeLaters  []responses.ConsumeLaterPreview
	)

	// OPTIMIZATION 1: Batch core content operations (reduce goroutine overhead)
	// Movie operations batch
	wg.Add(1)
	go func() {
		defer wg.Done()
		var movieWg sync.WaitGroup
		movieWg.Add(5)

		go func() {
			defer movieWg.Done()
			result.MovieData.UpcomingMovies, _ = movieModel.GetUpcomingPreviewMovies()
		}()
		go func() {
			defer movieWg.Done()
			result.MovieData.PopularMovies, _ = movieModel.GetPopularPreviewMovies()
		}()
		go func() {
			defer movieWg.Done()
			result.MovieData.TopMovies, _ = movieModel.GetTopPreviewMovies()
		}()
		go func() {
			defer movieWg.Done()
			result.MovieData.MoviesInTheater, _ = movieModel.GetInTheaterPreviewMovies()
		}()
		go func() {
			defer movieWg.Done()
			result.MovieData.MoviePopularPC, _ = movieModel.GetPopularProductionCompanies()
		}()
		movieWg.Wait()
	}()

	// OPTIMIZATION 2: Combined actor query (replaces separate movie + TV actor queries)
	wg.Add(1)
	go func() {
		defer wg.Done()
		result.MovieData.PopularActors, result.TVData.PopularActors, _ = movieModel.GetCombinedPopularActors(requests.Pagination{Page: 1})
	}()

	// TV operations batch (without actors - handled above)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var tvWg sync.WaitGroup
		tvWg.Add(5)

		go func() {
			defer tvWg.Done()
			result.TVData.UpcomingTVSeries, _ = tvModel.GetUpcomingPreviewTVSeries()
		}()
		go func() {
			defer tvWg.Done()
			result.TVData.PopularTVSeries, _ = tvModel.GetPopularPreviewTVSeries()
		}()
		go func() {
			defer tvWg.Done()
			result.TVData.TopTVSeries, _ = tvModel.GetTopPreviewTVSeries()
		}()
		go func() {
			defer tvWg.Done()
			result.TVData.AiringTVSeries, _ = tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek(int16(time.Now().UTC().Weekday()) + 1)
		}()
		go func() {
			defer tvWg.Done()
			result.TVData.TVPopularPC, _ = tvModel.GetPopularProductionCompanies()
		}()
		tvWg.Wait()
	}()

	// Anime operations batch
	wg.Add(1)
	go func() {
		defer wg.Done()
		var animeWg sync.WaitGroup
		animeWg.Add(5)

		go func() {
			defer animeWg.Done()
			result.AnimeData.UpcomingAnimes, _ = animeModel.GetPreviewUpcomingAnimes()
		}()
		go func() {
			defer animeWg.Done()
			result.AnimeData.TopRatedAnimes, _ = animeModel.GetPreviewTopAnimes()
		}()
		go func() {
			defer animeWg.Done()
			result.AnimeData.PopularAnimes, _ = animeModel.GetPreviewPopularAnimes()
		}()
		go func() {
			defer animeWg.Done()
			result.AnimeData.AiringAnime, _ = animeModel.GetCurrentlyAiringAnimesByDayOfWeek(int16(time.Now().UTC().Weekday()) + 1)
		}()
		go func() {
			defer animeWg.Done()
			result.AnimeData.AnimePopularStudios, _ = animeModel.GetPopularStudios()
		}()
		animeWg.Wait()
	}()

	// Game operations batch
	wg.Add(1)
	go func() {
		defer wg.Done()
		var gameWg sync.WaitGroup
		gameWg.Add(3)

		go func() {
			defer gameWg.Done()
			result.GameData.UpcomingGames, _ = gameModel.GetPreviewUpcomingGames()
		}()
		go func() {
			defer gameWg.Done()
			result.GameData.TopRatedGames, _ = gameModel.GetPreviewTopGames()
		}()
		go func() {
			defer gameWg.Done()
			result.GameData.PopularGames, _ = gameModel.GetPreviewPopularGames()
		}()
		gameWg.Wait()
	}()

	// OPTIMIZATION 3: Consume later operations only if user exists (single combined query)
	if userExists && userInteractionModel != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			movieConsumeLaters, tvConsumeLaters, animeConsumeLaters, gameConsumeLaters, _ = userInteractionModel.GetAllConsumeLaterForPreview(userID, 10)
		}()
	}

	wg.Wait()

	// OPTIMIZATION 3: Streamlined response building with pre-allocated map
	responseData := gin.H{
		"movie": gin.H{
			"upcoming":             result.MovieData.UpcomingMovies,
			"popular":              result.MovieData.PopularMovies,
			"top":                  result.MovieData.TopMovies,
			"extra":                result.MovieData.MoviesInTheater,
			"actors":               result.MovieData.PopularActors,
			"production_companies": result.MovieData.MoviePopularPC,
		},
		"tv": gin.H{
			"upcoming":             result.TVData.UpcomingTVSeries,
			"popular":              result.TVData.PopularTVSeries,
			"top":                  result.TVData.TopTVSeries,
			"extra":                result.TVData.AiringTVSeries,
			"actors":               result.TVData.PopularActors,
			"production_companies": result.TVData.TVPopularPC,
		},
		"anime": gin.H{
			"upcoming": result.AnimeData.UpcomingAnimes,
			"top":      result.AnimeData.TopRatedAnimes,
			"popular":  result.AnimeData.PopularAnimes,
			"extra":    result.AnimeData.AiringAnime,
			"studios":  result.AnimeData.AnimePopularStudios,
		},
		"game": gin.H{
			"upcoming": result.GameData.UpcomingGames,
			"top":      result.GameData.TopRatedGames,
			"popular":  result.GameData.PopularGames,
			"extra":    nil,
		},
	}

	// OPTIMIZATION 4: Only add consume later data if user exists (avoid map modifications)
	if userExists {
		responseData["movie"].(gin.H)["consume_laters"] = movieConsumeLaters
		responseData["tv"].(gin.H)["consume_laters"] = tvConsumeLaters
		responseData["anime"].(gin.H)["consume_laters"] = animeConsumeLaters
		responseData["game"].(gin.H)["consume_laters"] = gameConsumeLaters
	}

	c.JSON(http.StatusOK, responseData)
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

	// Initialize models once
	movieModel := models.NewMovieModel(pr.Database)
	tvModel := models.NewTVModel(pr.Database)
	animeModel := models.NewAnimeModel(pr.Database)
	gameModel := models.NewGameModel(pr.Database)

	var wg sync.WaitGroup
	result := previewResult{}

	// Movie operations (7 operations for V2)
	wg.Add(7)
	go func() {
		defer wg.Done()
		result.MovieData.UpcomingMovies, _ = movieModel.GetUpcomingPreviewMovies()
	}()

	go func() {
		defer wg.Done()
		result.MovieData.PopularMovies, _ = movieModel.GetPopularPreviewMovies()
	}()

	go func() {
		defer wg.Done()
		result.MovieData.TopMovies, _ = movieModel.GetTopPreviewMovies()
	}()

	go func() {
		defer wg.Done()
		result.MovieData.MoviesInTheater, _ = movieModel.GetInTheaterPreviewMovies()
	}()

	go func() {
		defer wg.Done()
		result.MovieData.MoviePopularSP, _ = movieModel.GetPopularStreamingPlatforms(data.Region)
	}()

	go func() {
		defer wg.Done()
		result.MovieData.MoviePopularPC, _ = movieModel.GetPopularProductionCompanies()
	}()

	// TV operations (7 operations for V2)
	wg.Add(7)
	go func() {
		defer wg.Done()
		result.TVData.UpcomingTVSeries, _ = tvModel.GetUpcomingPreviewTVSeries()
	}()

	go func() {
		defer wg.Done()
		result.TVData.PopularTVSeries, _ = tvModel.GetPopularPreviewTVSeries()
	}()

	go func() {
		defer wg.Done()
		result.TVData.TopTVSeries, _ = tvModel.GetTopPreviewTVSeries()
	}()

	go func() {
		defer wg.Done()
		result.TVData.AiringTVSeries, _ = tvModel.GetCurrentlyAiringTVSeriesByDayOfWeek(int16(time.Now().UTC().Weekday()) + 1)
	}()

	go func() {
		defer wg.Done()
		result.TVData.TVSeriesPopularSP, _ = tvModel.GetPopularStreamingPlatforms(data.Region)
	}()

	go func() {
		defer wg.Done()
		result.TVData.TVPopularPC, _ = tvModel.GetPopularProductionCompanies()
	}()

	// Anime operations (6 operations for V2)
	wg.Add(6)
	go func() {
		defer wg.Done()
		result.AnimeData.UpcomingAnimes, _ = animeModel.GetPreviewUpcomingAnimes()
	}()

	go func() {
		defer wg.Done()
		result.AnimeData.TopRatedAnimes, _ = animeModel.GetPreviewTopAnimes()
	}()

	go func() {
		defer wg.Done()
		result.AnimeData.PopularAnimes, _ = animeModel.GetPreviewPopularAnimes()
	}()

	go func() {
		defer wg.Done()
		result.AnimeData.AiringAnime, _ = animeModel.GetCurrentlyAiringAnimesByDayOfWeek(int16(time.Now().UTC().Weekday()) + 1)
	}()

	go func() {
		defer wg.Done()
		result.AnimeData.AnimePopularSP, _ = animeModel.GetPopularStreamingPlatforms()
	}()

	go func() {
		defer wg.Done()
		result.AnimeData.AnimePopularStudios, _ = animeModel.GetPopularStudios()
	}()

	// Game operations (3 operations)
	wg.Add(3)
	go func() {
		defer wg.Done()
		result.GameData.UpcomingGames, _ = gameModel.GetPreviewUpcomingGames()
	}()

	go func() {
		defer wg.Done()
		result.GameData.TopRatedGames, _ = gameModel.GetPreviewTopGames()
	}()

	go func() {
		defer wg.Done()
		result.GameData.PopularGames, _ = gameModel.GetPreviewPopularGames()
	}()

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"movie": gin.H{
			"upcoming":             result.MovieData.UpcomingMovies,
			"popular":              result.MovieData.PopularMovies,
			"top":                  result.MovieData.TopMovies,
			"extra":                result.MovieData.MoviesInTheater,
			"actors":               result.MovieData.PopularActors,
			"streaming_platforms":  result.MovieData.MoviePopularSP,
			"production_companies": result.MovieData.MoviePopularPC,
		},
		"tv": gin.H{
			"upcoming":             result.TVData.UpcomingTVSeries,
			"popular":              result.TVData.PopularTVSeries,
			"top":                  result.TVData.TopTVSeries,
			"extra":                result.TVData.AiringTVSeries,
			"actors":               result.TVData.PopularActors,
			"streaming_platforms":  result.TVData.TVSeriesPopularSP,
			"production_companies": result.TVData.TVPopularPC,
		},
		"anime": gin.H{
			"upcoming":                  result.AnimeData.UpcomingAnimes,
			"top":                       result.AnimeData.TopRatedAnimes,
			"popular":                   result.AnimeData.PopularAnimes,
			"extra":                     result.AnimeData.AiringAnime,
			"anime_streaming_platforms": result.AnimeData.AnimePopularSP,
			"studios":                   result.AnimeData.AnimePopularStudios,
		},
		"game": gin.H{
			"upcoming": result.GameData.UpcomingGames,
			"top":      result.GameData.TopRatedGames,
			"popular":  result.GameData.PopularGames,
			"extra":    nil,
		},
	})
}
