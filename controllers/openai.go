package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

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
	var data requests.OpenAIRecommendation
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	openAIModel := models.CreateOpenAIClient()

	resp, err := openAIModel.GetRecommendation(data.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	movieModel := models.NewMovieModel(ai.Database)

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
