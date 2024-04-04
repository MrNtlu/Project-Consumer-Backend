package controllers

import (
	"app/db"
	"app/helpers"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
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
func (feedback *FeedbackController) SendFeedback(c *gin.Context) {
	var data requests.Feedback
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	feedbackModel := models.NewFeedbackModel(feedback.Database)

	if err := feedbackModel.CreateFeedback(uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

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
