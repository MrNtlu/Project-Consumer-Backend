package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type RecommendationController struct {
	Database *db.MongoDB
}

func NewRecommendationController(mongoDB *db.MongoDB) RecommendationController {
	return RecommendationController{
		Database: mongoDB,
	}
}

// Create Recommendation
// @Summary Create Recommendation
// @Description Creates Recommendation
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param createrecommendation body requests.CreateRecommendation true "Create Recommendation"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} responses.RecommendationWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation [post]
func (rn *RecommendationController) CreateRecommendation(c *gin.Context) {
	var data requests.CreateRecommendation
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		contentTitleEn       string
		contentTitleOriginal string
		contentImage         string

		recommendationTitleEn       string
		recommendationTitleOriginal string
		recommendationImage         string
	)

	switch data.ContentType {
	case "anime":
		animeModel := models.NewAnimeModel(rn.Database)

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

		recommendationAnime, err := animeModel.GetAnimeDetails(requests.ID{
			ID: data.RecommendationID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if recommendationAnime.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		contentTitleOriginal = anime.TitleOriginal
		contentTitleEn = anime.TitleEn
		contentImage = anime.ImageURL

		recommendationTitleOriginal = recommendationAnime.TitleOriginal
		recommendationTitleEn = recommendationAnime.TitleEn
		recommendationImage = recommendationAnime.ImageURL
	case "game":
		gameModel := models.NewGameModel(rn.Database)

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

		recommendationGame, err := gameModel.GetGameDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if recommendationGame.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		contentTitleOriginal = game.TitleOriginal
		contentTitleEn = game.Title
		contentImage = game.ImageUrl

		recommendationTitleOriginal = recommendationGame.TitleOriginal
		recommendationTitleEn = recommendationGame.Title
		recommendationImage = recommendationGame.ImageUrl
	case "movie":
		movieModel := models.NewMovieModel(rn.Database)

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

		recommendationMovie, err := movieModel.GetMovieDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if recommendationMovie.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		contentTitleOriginal = movie.TitleOriginal
		contentTitleEn = movie.TitleEn
		contentImage = movie.ImageURL

		recommendationTitleOriginal = recommendationMovie.TitleOriginal
		recommendationTitleEn = recommendationMovie.TitleEn
		recommendationImage = recommendationMovie.ImageURL
	case "tv":
		tvSeriesModel := models.NewTVModel(rn.Database)

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

		recommendationTVSeries, err := tvSeriesModel.GetTVSeriesDetails(requests.ID{
			ID: data.ContentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if recommendationTVSeries.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		contentTitleOriginal = tvSeries.TitleOriginal
		contentTitleEn = tvSeries.TitleEn
		contentImage = tvSeries.ImageURL

		recommendationTitleOriginal = recommendationTVSeries.TitleOriginal
		recommendationTitleEn = recommendationTVSeries.TitleEn
		recommendationImage = recommendationTVSeries.ImageURL
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	recommendationModel := models.NewRecommendationModel(rn.Database)
	userModel := models.NewUserModel(rn.Database)

	var (
		recommendationWithContent responses.RecommendationWithContent
		err                       error
	)

	createdRecommendation, err := recommendationModel.CreateRecommendation(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	author, err := userModel.FindUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	recommendationWithContent = responses.RecommendationWithContent{
		ID:     createdRecommendation.ID,
		UserID: createdRecommendation.UserID,
		Author: responses.Author{
			ID:           author.ID,
			Image:        author.Image,
			Username:     author.Username,
			EmailAddress: author.EmailAddress,
			IsPremium:    author.IsPremium || author.IsLifetimePremium,
		},
		IsAuthor:         true,
		ContentID:        createdRecommendation.ContentID,
		ContentType:      createdRecommendation.ContentType,
		RecommendationID: createdRecommendation.RecommendationID,
		Reason:           createdRecommendation.Reason,
		IsLiked:          false,
		Likes:            []string{},
		Content: responses.ReviewContent{
			TitleEn:       contentTitleEn,
			TitleOriginal: contentTitleOriginal,
			ImageURL:      &contentImage,
		},
		RecommendationContent: responses.ReviewContent{
			TitleEn:       recommendationTitleEn,
			TitleOriginal: recommendationTitleOriginal,
			ImageURL:      &recommendationImage,
		},
		CreatedAt: createdRecommendation.CreatedAt,
	}

	logModel := models.NewLogsModel(rn.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.RecommendationLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: fmt.Sprintf("Recommended %s", recommendationTitleOriginal),
		ContentTitle:     contentTitleOriginal,
		ContentImage:     contentImage,
		ContentType:      data.ContentType,
		ContentID:        data.ContentID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": recommendationWithContent})
}
