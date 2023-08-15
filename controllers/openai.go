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
	userListModel := models.NewUserListModel(ai.Database)
	openAIModel := models.CreateOpenAIClient()

	uid := jwt.ExtractClaims(c)["id"].(string)
	watchList, err := userListModel.GetMovieListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	watchListAsStringList := make([]string, len(watchList))
	for i, movieWatchList := range watchList {
		score := *movieWatchList.Score
		watchListAsStringList[i] = fmt.Sprintf("%s, %.0f.", movieWatchList.TitleOriginal, score)
	}

	watchListAsString := strings.Join(watchListAsStringList, "\n")

	resp, err := openAIModel.GetRecommendation(watchListAsString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	movies, err := movieModel.GetMoviesFromOpenAI(resp.Recommendation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.OpenAIMovieResponse{
		OpenAIResponse: resp,
		Movies:         movies,
	}})
}
