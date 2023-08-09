package controllers

import (
	"app/models"
	"app/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OpenAIController struct{}

func NewOpenAiController() OpenAIController {
	return OpenAIController{}
}

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

	c.JSON(http.StatusOK, gin.H{"data": resp})
}
