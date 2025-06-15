package controllers

import (
	"app/db"
	"app/models"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type AchievementController struct {
	Database *db.MongoDB
}

func NewAchievementController(mongoDB *db.MongoDB) AchievementController {
	return AchievementController{
		Database: mongoDB,
	}
}

// Get User Achievements
// @Summary Get user achievements
// @Description Returns all achievements with unlocked status for the user
// @Tags achievements
// @Accept application/json
// @Produce application/json
// @Security ApiKeyAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.UserAchievementResponse "User Achievements"
// @Router /achievements [get]
func (a *AchievementController) GetUserAchievements(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	achievementModel := models.NewAchievementModel(a.Database)
	achievements, err := achievementModel.GetUserAchievements(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully fetched achievements.",
		"data":    achievements,
	})
}

// Get All Achievements
// @Summary Get all achievements
// @Description Returns all achievements without user-specific unlock status
// @Tags achievements
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.Achievement "All Achievements"
// @Router /achievements/all [get]
func (a *AchievementController) GetAllAchievements(c *gin.Context) {
	achievementModel := models.NewAchievementModel(a.Database)
	achievements, err := achievementModel.GetAllAchievements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully fetched all achievements.",
		"data":    achievements,
	})
}
