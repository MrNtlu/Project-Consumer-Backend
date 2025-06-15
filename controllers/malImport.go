package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type MALImportController struct {
	Database *db.MongoDB
}

func NewMALImportController(mongoDB *db.MongoDB) MALImportController {
	return MALImportController{
		Database: mongoDB,
	}
}

// Import MyAnimeList
// @Summary Import anime list from MyAnimeList
// @Description Imports user's anime list from MyAnimeList using their username
// @Tags import
// @Accept application/json
// @Produce application/json
// @Param malimport body requests.MALImportRequest true "MAL Import Request"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.MALImportResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /import/mal [post]
func (m *MALImportController) ImportFromMAL(c *gin.Context) {
	var data requests.MALImportRequest
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	malImportModel := models.NewMALImportModel(m.Database)

	result, err := malImportModel.ImportUserAnimeList(uid, data.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MyAnimeList import completed successfully.",
		"data":    result,
	})
}
