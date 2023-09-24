package controllers

import (
	"app/db"
	"app/models"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type OpenAIController struct {
	Database *db.MongoDB
}

func NewOpenAiController(mongoDB *db.MongoDB) OpenAIController {
	return OpenAIController{
		Database: mongoDB,
	}
}

//TODO Add for other contents

// Get Movie Recommendations
// @Summary Get Movie Recommendations from OpenAI
// @Description Returns movie recommendations from OpenAI
// @Tags openai
// @Accept application/json
// @Produce application/json
// @Success 200 {array} models.OpenAIMovieResponse
// @Failure 500 {string} string
// @Router /openai [get]
func (ai *OpenAIController) GetRecommendation(c *gin.Context) {
	movieModel := models.NewMovieModel(ai.Database)
	tvModel := models.NewTVModel(ai.Database)
	animeModel := models.NewAnimeModel(ai.Database)
	gameModel := models.NewGameModel(ai.Database)
	userListModel := models.NewUserListModel(ai.Database)
	openAIModel := models.CreateOpenAIClient()

	uid := jwt.ExtractClaims(c)["id"].(string)
	movieWatchList, err := userListModel.GetMovieListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	tvWatchList, err := userListModel.GetTVSeriesListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	animeWatchList, err := userListModel.GetAnimeListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	gamePlayList, err := userListModel.GetGameListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
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

	resp, err := openAIModel.GetRecommendation(inputString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	var (
		movieList []string
		tvList    []string
		animeList []string
		gameList  []string
	)
	for _, str := range resp.Recommendation {
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

	movies, err := movieModel.GetMoviesFromOpenAI(movieList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	tvSeries, err := tvModel.GetTVSeriesFromOpenAI(tvList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	anime, err := animeModel.GetAnimeFromOpenAI(animeList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	games, err := gameModel.GetGamesFromOpenAI(gameList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.OpenAIMovieResponse{
		OpenAIResponse: resp,
		Movies:         movies,
		TVSeries:       tvSeries,
		Anime:          anime,
		Games:          games,
	}})
}
