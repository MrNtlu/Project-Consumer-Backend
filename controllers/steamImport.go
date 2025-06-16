package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type SteamImportController struct {
	Database *db.MongoDB
}

func NewSteamImportController(mongoDB *db.MongoDB) SteamImportController {
	return SteamImportController{
		Database: mongoDB,
	}
}

// Import Steam Library
// @Summary Import game library from Steam
// @Description Imports user's game library from Steam using their Steam ID
// @Tags import
// @Accept application/json
// @Produce application/json
// @Param steamimport body requests.SteamImportRequest true "Steam Import Request"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.SteamImportResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /import/steam [post]
func (s *SteamImportController) ImportFromSteam(c *gin.Context) {
	var data requests.SteamImportRequest
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	// Validate that either Steam ID or username is provided
	if data.SteamID == "" && data.SteamUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either steam_id or steam_username must be provided",
		})
		return
	}

	if data.SteamID != "" && data.SteamUsername != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Provide either steam_id or steam_username, not both",
		})
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	steamImportModel := models.NewSteamImportModel(s.Database)

	// Resolve Steam ID if username was provided
	var steamID string
	if data.SteamUsername != "" {
		resolvedID, err := steamImportModel.ResolveSteamUsername(data.SteamUsername)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to resolve Steam username: %v", err),
			})
			return
		}
		steamID = resolvedID
	} else {
		steamID = data.SteamID
	}

	result, err := steamImportModel.ImportUserGameLibrary(uid, steamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Steam import completed successfully.",
		"data":    result,
	})
}
