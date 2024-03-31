package controllers

import (
	"app/db"
	"app/helpers"
	"app/requests"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeedbackController struct {
	Database *db.MongoDB
}

func NewFeedbackController(mongoDB *db.MongoDB) FeedbackController {
	return FeedbackController{
		Database: mongoDB,
	}
}

// Send feedback
// @Summary Send feedback
// @Description Send feedback
// @Tags feedback
// @Accept application/json
// @Produce application/json
// @Param feedback body requests.Feedback true "Feedback"
// @Success 200 {object} string
// @Failure 500 {string} string
// @Router /feedback [patch]
func (feeedback *FeedbackController) SendFeedback(c *gin.Context) {
	var data requests.Feedback
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	if err := helpers.SendFeedbackMail(data.Feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully send feedback."})
}
