package controllers

import (
	"app/db"
	"app/models"
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
			c.JSON(http.StatusBadRequest, errNotEnoughUserList)

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
		movieList = aiSuggestion.Movies
		tvList = aiSuggestion.TVSeries
		animeList = aiSuggestion.Anime
		gameList = aiSuggestion.Games
		createdAt = aiSuggestion.CreatedAt
	}

	movies, err := movieModel.GetMoviesFromOpenAI(uid, movieList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	tvSeries, err := tvModel.GetTVSeriesFromOpenAI(uid, tvList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	anime, err := animeModel.GetAnimeFromOpenAI(uid, animeList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	games, err := gameModel.GetGamesFromOpenAI(uid, gameList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
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
