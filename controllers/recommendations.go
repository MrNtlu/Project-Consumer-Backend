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

const errCannotRecommendSelf = "You cannot recommend a content for itself."

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

	if data.RecommendationID == data.ContentID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errCannotRecommendSelf,
		})

		return
	}

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

// Get Recommendations by Content
// @Summary Get Recommendations by Content
// @Description Get Recommendations by Content
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param sortrecommendation body requests.SortRecommendation true "Sort Recommendation"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.RecommendationWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation [get]
func (rn *RecommendationController) GetRecommendationsByContentID(c *gin.Context) {
	var data requests.SortRecommendation
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	recommendationModel := models.NewRecommendationModel(rn.Database)

	var (
		userID    string
		isUIDNull bool
	)
	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		isUIDNull = false
		userID = uid.(string)
	} else {
		isUIDNull = true
		userID = ""
	}

	recommendations, pagination, err := recommendationModel.GetRecommendationsByContentID(userID, isUIDNull, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": recommendations})
}

// Get Recommendations by User ID
// @Summary Get Recommendations by User ID
// @Description Get Recommendations by User ID
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param sortrecommendationbycontentid body requests.SortRecommendationByUserID true "Sort Recommendation by Content ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.RecommendationWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation/profile [get]
func (rn *RecommendationController) GetRecommendationsByUserID(c *gin.Context) {
	var data requests.SortRecommendationByUserID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	recommendationModel := models.NewRecommendationModel(rn.Database)

	uid := jwt.ExtractClaims(c)["id"].(string)
	recommendations, pagination, err := recommendationModel.GetRecommendationsByUserID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": recommendations})
}

// Get Recommendations for Social
// @Summary Get Recommendations for Social
// @Description Get Recommendations for Social
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param sortrecommendationforsocial body requests.SortRecommendationsForSocial true "Sort Recommendation for Social"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.RecommendationWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation/social [get]
func (rn *RecommendationController) GetRecommendationsForSocial(c *gin.Context) {
	var data requests.SortRecommendationsForSocial
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	recommendationModel := models.NewRecommendationModel(rn.Database)

	var (
		userID    string
		isUIDNull bool
	)
	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		isUIDNull = false
		userID = uid.(string)
	} else {
		isUIDNull = true
		userID = ""
	}

	recommendations, pagination, err := recommendationModel.GetRecommendationsForSocial(userID, isUIDNull, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": recommendations})
}

// Get Liked Recommendations
// @Summary Get Liked Recommendations
// @Description Get Liked Recommendations
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param sortreview body requests.SortReview true "Sort Review"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.RecommendationWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation/liked [get]
func (rn *RecommendationController) GetLikedRecommendations(c *gin.Context) {
	var data requests.SortReview
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	recommendationModel := models.NewRecommendationModel(rn.Database)

	uid := jwt.ExtractClaims(c)["id"].(string)
	recommendations, pagination, err := recommendationModel.GetLikedRecommendations(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": recommendations})
}

// Like/Dislike Recommendation
// @Summary Like/Dislike Recommendation
// @Description Like Recommendation
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param likerecommendation body requests.LikeRecommendation true "Like Recommendation"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.RecommendationWithContent
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation/like [patch]
func (rn *RecommendationController) LikeRecommendation(c *gin.Context) {
	var data requests.LikeRecommendation
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	recommendationModel := models.NewRecommendationModel(rn.Database)
	uid := jwt.ExtractClaims(c)["id"].(string)

	var (
		updatedRecommendation responses.RecommendationWithContent
		err                   error
	)

	recommendation, err := recommendationModel.GetBaseRecommendationWithContent(uid, false, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if recommendation.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if recommendation.UserID == uid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot like your own recommendation.",
		})

		return
	}

	if updatedRecommendation, err = recommendationModel.LikeRecommendation(uid, data.RecommendationID, recommendation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(rn.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.RecommendationLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: "Vote",
		ContentID:        recommendation.ContentID,
		ContentTitle:     recommendation.Content.TitleOriginal,
		ContentType:      recommendation.ContentType,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Successfully liked.", "data": updatedRecommendation})
}

// Delete Recommendation
// @Summary Delete Recommendation
// @Description Delete Recommendation
// @Tags recommendation
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.RecommendationWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /recommendation [delete]
func (rn *RecommendationController) DeleteRecommendationByID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	recommendationModel := models.NewRecommendationModel(rn.Database)

	isDeleted, err := recommendationModel.DeleteRecommendationByID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		logModel := models.NewLogsModel(rn.Database)

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.RecommendationLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
		})

		c.JSON(http.StatusOK, gin.H{"message": "Recommendation deleted successfully."})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}
