package controllers

import (
	"app/db"
	"app/models"
	"app/responses"
	"context"
	"net/http"
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
	aiSuggestionsModel := models.NewAISuggestionsModel(ai.Database)
	userModel := models.NewUserModel(ai.Database)
	userListModel := models.NewUserListModel(ai.Database)
	movieModel := models.NewMovieModel(ai.Database)
	tvModel := models.NewTVModel(ai.Database)
	animeModel := models.NewAnimeModel(ai.Database)
	gameModel := models.NewGameModel(ai.Database)

	aiSuggestion, _ := aiSuggestionsModel.GetAISuggestions(uid)

	isPremium, _ := userModel.IsUserPremium(uid)
	currentDate := time.Now().UTC()

	var (
		recommendations []responses.AISuggestion
		createdAt       time.Time
	)
	if aiSuggestion.UserID == "" || (aiSuggestion.UserID != "" &&
		((isPremium && (currentDate.Sub(aiSuggestion.CreatedAt).Hours()/24) >= 7) ||
			(!isPremium && (currentDate.Sub(aiSuggestion.CreatedAt).Hours()/24) >= 30))) {

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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Couldn't generate new, sorry.",
			})

			return
		} else {
			if aiSuggestion.UserID == "" {
				go aiSuggestionsModel.CreateAISuggestions(uid, movieList, tvList, animeList, gameList)
			} else {
				if _, err := aiSuggestionsModel.DeleteAISuggestionsByUserID(uid); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})

					return
				}

				go aiSuggestionsModel.CreateAISuggestions(uid, movieList, tvList, animeList, gameList)
			}
		}

		recommendations = allSuggestions
	} else {
		allSuggestions, err := ai.FetchRecommendations(
			uid,
			aiSuggestion.Movies,
			aiSuggestion.TVSeries,
			aiSuggestion.Anime,
			aiSuggestion.Games,
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
	// Extract each ID slice from the raw map
	getIDs := func(key string) []string {
		if arr, ok := rawRecs[key].([]map[string]interface{}); ok {
			ids := make([]string, len(arr))
			for i, item := range arr {
				ids[i] = item["id"].(string)
			}
			return ids
		}
		return nil
	}

	movieList := getIDs("movies")
	tvList := getIDs("tvSeries")
	animeList := getIDs("animes")
	gameList := getIDs("games")

	logrus.Println("movieList", movieList)
	logrus.Println("tvList", tvList)
	logrus.Println("animeList", animeList)
	logrus.Println("gameList", gameList)

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
) (
	[]responses.AISuggestion,
	error,
) {
	var allSuggestions []responses.AISuggestion

	// For each type, fetch details via the respective model
	if len(movieList) > 0 {
		movies, err := movieModel.GetMoviesFromOpenAI(uid, movieList, 10)
		if err != nil {
			return nil, err
		}
		allSuggestions = append(allSuggestions, movies...)
	}

	if len(tvList) > 0 {
		tvSeries, err := tvModel.GetTVSeriesFromOpenAI(uid, tvList, 10)
		if err != nil {
			return nil, err
		}
		allSuggestions = append(allSuggestions, tvSeries...)
	}

	if len(animeList) > 0 {
		anime, err := animeModel.GetAnimeFromOpenAI(uid, animeList, 10)
		if err != nil {
			return nil, err
		}
		allSuggestions = append(allSuggestions, anime...)
	}

	if len(gameList) > 0 {
		games, err := gameModel.GetGamesFromOpenAI(uid, gameList, 10)
		if err != nil {
			return nil, err
		}
		allSuggestions = append(allSuggestions, games...)
	}

	return allSuggestions, nil
}
