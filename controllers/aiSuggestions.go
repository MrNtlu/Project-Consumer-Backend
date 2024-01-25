package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type AISuggestionsController struct {
	Database *db.MongoDB
}

func NewAISuggestionsController(mongoDB *db.MongoDB) AISuggestionsController {
	return AISuggestionsController{
		Database: mongoDB,
	}
}

var (
	errNotEnoughUserList = "You need to have more content in your list. Total of 5 content is required. e.g. 1 Movie, 2 TV Series, 2 Anime, 0 Game."
)

// Get Content Summary
// @Summary Get Content Summary
// @Description Returns content summary
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Param assistantrequest body requests.AssistantRequest true "Assistant Request"
// @Success 200 {object} string
// @Failure 500 {string} string
// @Router /assistant/summary [get]
func (ai *AISuggestionsController) GetSummary(c *gin.Context) {
	var data requests.AssistantRequest
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	userModel := models.NewUserModel(ai.Database)
	openAIModel := models.CreateOpenAIClient()

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium, _ := userModel.IsUserPremium(uid)

	if isPremium {
		summary, err := openAIModel.GetSummary(data.ContentName, data.ContentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": summary.Response})
		return
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "You need to have a premium membership to do this action.",
		})

		return
	}
}

// Get Content Public Opinion
// @Summary Get Content Public Opinion
// @Description Returns content Public Opinion
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Param assistantrequest body requests.AssistantRequest true "Assistant Request"
// @Success 200 {object} string
// @Failure 500 {string} string
// @Router /assistant/opinion [get]
func (ai *AISuggestionsController) GetPublicOpinion(c *gin.Context) {
	var data requests.AssistantRequest
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	userModel := models.NewUserModel(ai.Database)
	openAIModel := models.CreateOpenAIClient()

	uid := jwt.ExtractClaims(c)["id"].(string)
	isPremium, _ := userModel.IsUserPremium(uid)

	if isPremium {
		summary, err := openAIModel.GetPublicOpinion(data.ContentName, data.ContentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": summary.Response})

		return
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "You need to have a premium membership to do this action.",
		})

		return
	}

}

// Get AI Recommendations
// @Summary Get AI Recommendations from OpenAI
// @Description Returns ai recommendations from OpenAI
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Success 200 {object} responses.AISuggestionResponse
// @Failure 500 {string} string
// @Router /suggestions [get]
func (ai *AISuggestionsController) GetAISuggestions(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	aiSuggestionsModel := models.NewAISuggestionsModel(ai.Database)
	userModel := models.NewUserModel(ai.Database)
	movieModel := models.NewMovieModel(ai.Database)
	tvModel := models.NewTVModel(ai.Database)
	animeModel := models.NewAnimeModel(ai.Database)
	gameModel := models.NewGameModel(ai.Database)

	aiSuggestion, _ := aiSuggestionsModel.GetAISuggestions(uid)

	isPremium, _ := userModel.IsUserPremium(uid)
	currentDate := time.Now().UTC()

	var (
		movieList []string
		tvList    []string
		animeList []string
		gameList  []string
		createdAt time.Time
	)
	if aiSuggestion.UserID == "" || (aiSuggestion.UserID != "" &&
		((isPremium && (currentDate.Sub(aiSuggestion.CreatedAt).Hours()/24) >= 7) ||
			(!isPremium && (currentDate.Sub(aiSuggestion.CreatedAt).Hours()/24) >= 30))) {

		c.JSON(http.StatusOK, gin.H{"data": nil, "message": "You can generate your recommendations now!"})

		return
	} else {
		movieList = aiSuggestion.Movies
		tvList = aiSuggestion.TVSeries
		animeList = aiSuggestion.Anime
		gameList = aiSuggestion.Games
		createdAt = aiSuggestion.CreatedAt
	}

	var (
		movies []responses.AISuggestion
		err    error
	)
	if len(gameList) > 0 {
		movies, err = movieModel.GetMoviesFromOpenAI(uid, movieList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	var tvSeries []responses.AISuggestion
	if len(gameList) > 0 {
		tvSeries, err = tvModel.GetTVSeriesFromOpenAI(uid, tvList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	var anime []responses.AISuggestion
	if len(gameList) > 0 {
		anime, err = animeModel.GetAnimeFromOpenAI(uid, animeList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	var games []responses.AISuggestion
	if len(gameList) > 0 {
		games, err = gameModel.GetGamesFromOpenAI(uid, gameList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	suggestions := append(movies, tvSeries...)
	suggestions = append(suggestions, anime...)
	suggestions = append(suggestions, games...)

	c.JSON(http.StatusOK, gin.H{"data": responses.AISuggestionResponse{
		Suggestions: suggestions,
		CreatedAt:   createdAt,
	}, "message": "Successfully returned."})

	return
}

// Generate AI Recommendations
// @Summary Generate AI Recommendations from OpenAI
// @Description Generates and returns ai recommendations from OpenAI
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Success 200 {object} responses.AISuggestionResponse
// @Failure 500 {string} string
// @Router /suggestions/generate [post]
func (ai *AISuggestionsController) GenerateAISuggestions(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)
	aiSuggestionsModel := models.NewAISuggestionsModel(ai.Database)
	userModel := models.NewUserModel(ai.Database)
	userListModel := models.NewUserListModel(ai.Database)
	openAIModel := models.CreateOpenAIClient()
	movieModel := models.NewMovieModel(ai.Database)
	tvModel := models.NewTVModel(ai.Database)
	animeModel := models.NewAnimeModel(ai.Database)
	gameModel := models.NewGameModel(ai.Database)

	aiSuggestion, _ := aiSuggestionsModel.GetAISuggestions(uid)

	isPremium, _ := userModel.IsUserPremium(uid)
	currentDate := time.Now().UTC()

	var (
		movieList []string
		tvList    []string
		animeList []string
		gameList  []string
		createdAt time.Time
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

		inputString, err := getUserListAsString(userListModel, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		resp, err := openAIModel.GetRecommendation(inputString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		movieList, tvList, animeList, gameList = parseResponseString(resp.Recommendation)
		createdAt = time.Now().UTC()

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
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't generate new, sorry.",
		})

		return
	}

	var (
		movies []responses.AISuggestion
		err    error
	)
	if len(gameList) > 0 {
		movies, err = movieModel.GetMoviesFromOpenAI(uid, movieList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	var tvSeries []responses.AISuggestion
	if len(gameList) > 0 {
		tvSeries, err = tvModel.GetTVSeriesFromOpenAI(uid, tvList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	var anime []responses.AISuggestion
	if len(gameList) > 0 {
		anime, err = animeModel.GetAnimeFromOpenAI(uid, animeList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	var games []responses.AISuggestion
	if len(gameList) > 0 {
		games, err = gameModel.GetGamesFromOpenAI(uid, gameList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	suggestions := append(movies, tvSeries...)
	suggestions = append(suggestions, anime...)
	suggestions = append(suggestions, games...)

	c.JSON(http.StatusOK, gin.H{"data": responses.AISuggestionResponse{
		Suggestions: suggestions,
		CreatedAt:   createdAt,
	}})

	return
}

func getUserListAsString(userListModel *models.UserListModel, uid string) (string, error) {
	movieWatchList, err := userListModel.GetMovieListByUserID(uid)
	if err != nil {
		return "", err
	}

	tvWatchList, err := userListModel.GetTVSeriesListByUserID(uid)
	if err != nil {
		return "", err
	}

	animeWatchList, err := userListModel.GetAnimeListByUserID(uid)
	if err != nil {
		return "", err
	}

	gamePlayList, err := userListModel.GetGameListByUserID(uid)
	if err != nil {
		return "", err
	}

	movieWatchListAsStringList := make([]string, len(movieWatchList))
	for i, movieWatchList := range movieWatchList {
		var score string
		if movieWatchList.Score != nil {
			score = fmt.Sprintf("%.0f", *movieWatchList.Score)
		} else {
			score = "*"
		}
		movieWatchListAsStringList[i] = fmt.Sprintf("%s, %s.", movieWatchList.TitleOriginal, score)
	}

	tvWatchListAsStringList := make([]string, len(tvWatchList))
	for i, tvWatchList := range tvWatchList {
		var score string
		if tvWatchList.Score != nil {
			score = fmt.Sprintf("%.0f", *tvWatchList.Score)
		} else {
			score = "*"
		}
		tvWatchListAsStringList[i] = fmt.Sprintf("%s, %s.", tvWatchList.TitleOriginal, score)
	}

	animeWatchListAsStringList := make([]string, len(animeWatchList))
	for i, animeWatchList := range animeWatchList {
		var score string
		if animeWatchList.Score != nil {
			score = fmt.Sprintf("%.0f", *animeWatchList.Score)
		} else {
			score = "*"
		}
		animeWatchListAsStringList[i] = fmt.Sprintf("%s, %s.", animeWatchList.TitleOriginal, score)
	}

	gamePlayListAsStringList := make([]string, len(animeWatchList))
	for i, gamePlayList := range gamePlayList {
		var score string
		if gamePlayList.Score != nil {
			score = fmt.Sprintf("%.0f", *gamePlayList.Score)
		} else {
			score = "*"
		}
		gamePlayListAsStringList[i] = fmt.Sprintf("%s, %s.", gamePlayList.TitleOriginal, score)
	}

	movieWatchListAsString := strings.Join(movieWatchListAsStringList, "\n")
	tvWatchListAsString := strings.Join(tvWatchListAsStringList, "\n")
	animeWatchListAsString := strings.Join(animeWatchListAsStringList, "\n")
	gamePlayListAsString := strings.Join(gamePlayListAsStringList, "\n")

	inputString := fmt.Sprintf("Movies:\n%s\nTV Series:\n%s\nAnime:\n%s\nGames:\n%s", movieWatchListAsString, tvWatchListAsString, animeWatchListAsString, gamePlayListAsString)

	return inputString, nil
}

func parseResponseString(response []string) ([]string, []string, []string, []string) {
	var (
		movieList []string
		tvList    []string
		animeList []string
		gameList  []string
	)
	for _, str := range response {
		if strings.HasPrefix(str, "Movie:") {
			_, movieName, _ := strings.Cut(str, "Movie:")
			if strings.HasPrefix(movieName, " ") {
				movieName = strings.TrimPrefix(movieName, " ")
			}

			movieList = append(movieList, movieName)
		} else if strings.HasPrefix(str, "TV Series:") {
			_, tvName, _ := strings.Cut(str, "TV Series:")
			if strings.HasPrefix(tvName, " ") {
				tvName = strings.TrimPrefix(tvName, " ")
			}

			tvList = append(tvList, tvName)
		} else if strings.HasPrefix(str, "Anime:") {
			_, animeName, _ := strings.Cut(str, "Anime:")
			if strings.HasPrefix(animeName, " ") {
				animeName = strings.TrimPrefix(animeName, " ")
			}

			animeList = append(animeList, animeName)
		} else if strings.HasPrefix(str, "Game:") {
			_, gameName, _ := strings.Cut(str, "Game:")
			if strings.HasPrefix(gameName, " ") {
				gameName = strings.TrimPrefix(gameName, " ")
			}

			gameList = append(gameList, gameName)
		}
	}

	return movieList, tvList, animeList, gameList
}
