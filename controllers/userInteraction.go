package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type UserInteractionController struct {
	Database *db.MongoDB
}

func NewUserInteractionController(mongoDB *db.MongoDB) UserInteractionController {
	return UserInteractionController{
		Database: mongoDB,
	}
}

// Create Consume Later
// @Summary Create Consume Later
// @Description Creates Consume Later
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Param createconsumelater body requests.CreateConsumeLater true "Create Consume Later"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.ConsumeLaterList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /consume [post]
func (ui *UserInteractionController) CreateConsumeLater(c *gin.Context) {
	var data requests.CreateConsumeLater
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	switch data.ContentType {
	case "anime":
		animeModel := models.NewAnimeModel(ui.Database)

		anime, err := animeModel.GetAnimeDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if anime.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}
	case "game":
		gameModel := models.NewGameModel(ui.Database)

		game, err := gameModel.GetGameDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if game.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}
	case "movie":
		movieModel := models.NewMovieModel(ui.Database)

		movie, err := movieModel.GetMovieDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if movie.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}
	case "tv":
		tvSeriesModel := models.NewTVModel(ui.Database)

		tvSeries, err := tvSeriesModel.GetTVSeriesDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if tvSeries.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	var (
		createdConsumeLater models.ConsumeLaterList
		err                 error
	)

	if createdConsumeLater, err = userInteractionModel.CreateConsumeLater(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdConsumeLater})
}

// Delete Consume Later
// @Summary Delete Consume Later
// @Description Deletes Consume Later
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /consume [delete]
func (ui *UserInteractionController) DeleteConsumeLaterById(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	isDeleted, err := userInteractionModel.DeleteConsumeLaterByID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully."})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}

// Delete All Consume Later
// @Summary Delete All Consume Later
// @Description Deletes All Consume Later
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /consume/all [delete]
func (ui *UserInteractionController) DeleteAllConsumeLaterByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	if err := userInteractionModel.DeleteAllConsumeLaterByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Deleted successfully."})
}
