package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"context"
	"net/http"
	"sync"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/sirupsen/logrus"
)

type AISuggestionsController struct {
	Database      *db.MongoDB
	Pinecone      *pinecone.Client
	PineconeCtrl  *PineconeController
	PineconeIndex *pinecone.IndexConnection
}

func NewAISuggestionsController(
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
) AISuggestionsController {

	pineconeCtrl := NewPineconeController(mongoDB, pinecone, pineconeIndex)

	return AISuggestionsController{
		Database:      mongoDB,
		Pinecone:      pinecone,
		PineconeCtrl:  &pineconeCtrl,
		PineconeIndex: pineconeIndex,
	}
}

var (
	errNotEnoughUserList = "You need to have more content in your list. Total of 5 content is required. e.g. 1 Movie, 2 TV Series, 2 Anime, 0 Game."
)

// Mark Content as Not Interested
// @Summary Mark Content as Not Interested
// @Description Marks a content as not interested
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /suggestions/not-interested [post]
func (ai *AISuggestionsController) NotInterested(c *gin.Context) {
	var data requests.NotInterestedRequest
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	aiSuggestionsModel := models.NewAISuggestionsModel(ai.Database)
	if data.IsDelete {
		err := aiSuggestionsModel.DeleteNotInterested(uid, data.ContentID, data.ContentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Content removed from not interested"})
		return
	}

	err := aiSuggestionsModel.CreateNotInterested(uid, data.ContentID, data.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Content marked as not interested"})
}

// Generate AI Recommendations
// @Summary Generate AI Recommendations from OpenAI
// @Description Generates and returns ai recommendations from OpenAI
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Success 200 {object} responses.AISuggestionResponse
// @Failure 500 {string} string
// @Router /suggestions/ [get]
func (ai *AISuggestionsController) GenerateAISuggestions(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	suggestModel := models.NewAISuggestionsModel(ai.Database)
	userModel := models.NewUserModel(ai.Database)
	userListModel := models.NewUserListModel(ai.Database)
	movieModel := models.NewMovieModel(ai.Database)
	tvModel := models.NewTVModel(ai.Database)
	animeModel := models.NewAnimeModel(ai.Database)
	gameModel := models.NewGameModel(ai.Database)

	var (
		aiRec     models.AISuggestions
		isPremium bool
		wgInit    sync.WaitGroup
	)
	wgInit.Add(2)
	go func() {
		defer wgInit.Done()
		aiRec, _ = suggestModel.GetAISuggestions(uid)
	}()
	go func() {
		defer wgInit.Done()
		isPremium, _ = userModel.IsUserPremium(uid)
	}()
	wgInit.Wait()

	currentDate := time.Now().UTC()
	ageDays := currentDate.Sub(aiRec.CreatedAt).Hours() / 24

	var (
		recommendations []responses.AISuggestion
		createdAt       time.Time
	)
	needRefresh := aiRec.UserID == "" || (isPremium && ageDays >= 3) || (!isPremium && ageDays >= 10)

	if needRefresh {
		count, err := userListModel.GetUserListCount(uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if count < 5 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": errNotEnoughUserList,
			})

			return
		}

		rawRecs, err := ai.PineconeCtrl.GetRecommendationsByType(uid, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		allSuggestions, movieList, tvList, animeList, gameList, err := ai.ApplyRecommendations(
			c.Request.Context(),
			uid,
			rawRecs,
			movieModel,
			tvModel,
			animeModel,
			gameModel,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if len(allSuggestions) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't generate new, sorry."})
			return
		}

		go func() {
			if aiRec.UserID != "" {
				suggestModel.DeleteAISuggestionsByUserID(uid)
			}
			suggestModel.CreateAISuggestions(uid, movieList, tvList, animeList, gameList)
		}()

		recommendations = allSuggestions
		createdAt = currentDate
	} else {
		allSuggestions, err := ai.FetchRecommendations(
			uid,
			aiRec.Movies,
			aiRec.TVSeries,
			aiRec.Anime,
			aiRec.Games,
			movieModel,
			tvModel,
			animeModel,
			gameModel,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		recommendations = allSuggestions
		createdAt = aiRec.CreatedAt
	}

	c.JSON(http.StatusOK, gin.H{"data": responses.AISuggestionResponse{
		Suggestions: recommendations,
		CreatedAt:   createdAt,
	}})
}

// ApplyRecommendations takes the raw Pinecone rec map and fetches detailed AISuggestions
// by delegating to the OpenAI-backed model fetchers (or DB fetchers).
func (ai *AISuggestionsController) ApplyRecommendations(ctx context.Context, uid string,
	rawRecs map[string]interface{},
	movieModel *models.MovieModel,
	tvModel *models.TVModel,
	animeModel *models.AnimeModel,
	gameModel *models.GameModel,
) (
	[]responses.AISuggestion,
	[]string,
	[]string,
	[]string,
	[]string,
	error,
) {
	var (
		movieList []string
		tvList    []string
		animeList []string
		gameList  []string
		mu        sync.Mutex
		wg        sync.WaitGroup
	)
	wg.Add(4)

	go func() {
		defer wg.Done()
		if arr, ok := rawRecs["movies"].([]map[string]interface{}); ok {
			ids := make([]string, len(arr))
			for i, item := range arr {
				if id, ok2 := item["id"].(string); ok2 {
					ids[i] = id
				}
			}
			mu.Lock()
			movieList = ids
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		if arr, ok := rawRecs["tvSeries"].([]map[string]interface{}); ok {
			ids := make([]string, len(arr))
			for i, item := range arr {
				if id, ok2 := item["id"].(string); ok2 {
					ids[i] = id
				}
			}
			mu.Lock()
			tvList = ids
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		if arr, ok := rawRecs["animes"].([]map[string]interface{}); ok {
			ids := make([]string, len(arr))
			for i, item := range arr {
				if id, ok2 := item["id"].(string); ok2 {
					ids[i] = id
				}
			}
			mu.Lock()
			animeList = ids
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		if arr, ok := rawRecs["games"].([]map[string]interface{}); ok {
			ids := make([]string, len(arr))
			for i, item := range arr {
				if id, ok2 := item["id"].(string); ok2 {
					ids[i] = id
				}
			}
			mu.Lock()
			gameList = ids
			mu.Unlock()
		}
	}()
	wg.Wait()

	allSuggestions, err := ai.FetchRecommendations(
		uid,
		movieList,
		tvList,
		animeList,
		gameList,
		movieModel,
		tvModel,
		animeModel,
		gameModel,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	logrus.Infof("Applied detailed fetch for %d recommendations", len(allSuggestions))
	return allSuggestions, movieList, tvList, animeList, gameList, nil
}

func (ai *AISuggestionsController) FetchRecommendations(
	uid string,
	movieList []string,
	tvList []string,
	animeList []string,
	gameList []string,
	movieModel *models.MovieModel,
	tvModel *models.TVModel,
	animeModel *models.AnimeModel,
	gameModel *models.GameModel,
) ([]responses.AISuggestion, error) {
	var (
		allSuggestions []responses.AISuggestion
		mu             sync.Mutex
		wg             sync.WaitGroup
		errs           = make(chan error, 4)
	)

	// Helper to fetch and append suggestions
	fetch := func(list []string, fetchFunc func(string, []string, int) ([]responses.AISuggestion, error)) {
		defer wg.Done()
		if len(list) == 0 {
			return
		}
		suggestions, err := fetchFunc(uid, list, 10)
		if err != nil {
			errs <- err
			return
		}
		mu.Lock()
		allSuggestions = append(allSuggestions, suggestions...)
		mu.Unlock()
	}

	// Launch concurrent fetches
	wg.Add(4)
	go fetch(movieList, movieModel.GetMoviesFromOpenAI)
	go fetch(tvList, tvModel.GetTVSeriesFromOpenAI)
	go fetch(animeList, animeModel.GetAnimeFromOpenAI)
	go fetch(gameList, gameModel.GetGamesFromOpenAI)
	wg.Wait()

	// Check errors
	close(errs)
	if err, ok := <-errs; ok {
		return nil, err
	}

	return allSuggestions, nil
}
