package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type ReviewController struct {
	Database *db.MongoDB
}

func NewReviewController(mongoDB *db.MongoDB) ReviewController {
	return ReviewController{
		Database: mongoDB,
	}
}

//POST, PUT, DELETE, REVIEWS BY CONTENT

// Create Review
// @Summary Create Review
// @Description Creates Review
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param createreview body requests.CreateReview true "Create Review"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.Review
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review [post]
func (r *ReviewController) CreateReview(c *gin.Context) {
	var data requests.CreateReview
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		contentTitle string
		contentImage string
	)

	switch data.ContentType {
	case "anime":
		animeModel := models.NewAnimeModel(r.Database)

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

		contentTitle = anime.TitleOriginal
		contentImage = anime.ImageURL
	case "game":
		gameModel := models.NewGameModel(r.Database)

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

		contentTitle = game.Title
		contentImage = game.ImageUrl
	case "movie":
		movieModel := models.NewMovieModel(r.Database)

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

		contentTitle = movie.TitleEn
		contentImage = movie.ImageURL
	case "tv":
		tvSeriesModel := models.NewTVModel(r.Database)

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

		contentTitle = tvSeries.TitleEn
		contentImage = tvSeries.ImageURL
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	reviewController := models.NewReviewModel(r.Database)

	var (
		createdReview models.Review
		err           error
	)

	if createdReview, err = reviewController.CreateReview(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(r.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.ReviewLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: "",
		ContentTitle:     contentTitle,
		ContentImage:     contentImage,
		ContentType:      data.ContentType,
		ContentID:        data.ContentID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdReview})
}
