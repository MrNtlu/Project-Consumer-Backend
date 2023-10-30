package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
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
// @Success 201 {object} responses.Review
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
	reviewModel := models.NewReviewModel(r.Database)
	userModel := models.NewUserModel(r.Database)

	var (
		createdReview responses.Review
		err           error
	)

	if createdReview, err = reviewModel.CreateReview(uid, data); err != nil {
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

	createdReview.IsAuthor = true
	createdReview.Author.ID = author.ID
	createdReview.Author.EmailAddress = author.EmailAddress
	createdReview.Author.Image = author.Image
	createdReview.Author.Username = author.Username
	createdReview.Author.IsPremium = author.IsPremium

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

// Get Reviews
// @Summary Get Review
// @Description Get Reviews with or without authentication
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param sortreviewbycontentid body requests.SortReviewByContentID true "Sort Review by Content ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.Review
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review [get]
func (r *ReviewController) GetReviewsByContentID(c *gin.Context) {
	var data requests.SortReviewByContentID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	reviewModel := models.NewReviewModel(r.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		reviews, pagination, err := reviewModel.GetReviewsByContentIDAndUserID(uid.(string), data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": reviews})
	} else {
		reviews, pagination, err := reviewModel.GetReviewsByContentID(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": reviews})
	}
}

// Update Review
// @Summary Update Review
// @Description Update Review
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param createreview body requests.CreateReview true "Create Review"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.Review
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review [patch]
func (r *ReviewController) UpdateReview(c *gin.Context) {
	var data requests.UpdateReview
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	reviewModel := models.NewReviewModel(r.Database)

	var (
		updatedReview responses.Review
		err           error
	)

	review, err := reviewModel.GetBaseReviewResponse(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if review.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if updatedReview, err = reviewModel.UpdateReview(uid, data, review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(r.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.ReviewLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: "",
		ContentID:        review.ContentID,
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully updated.", "data": updatedReview})
}

// Vote Review
// @Summary Vote Review
// @Description Like Review
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.Review
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/vote [patch]
func (r *ReviewController) VoteReview(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	reviewModel := models.NewReviewModel(r.Database)

	var (
		updatedReview responses.Review
		err           error
	)

	review, err := reviewModel.GetBaseReviewResponse(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if review.UserID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	if review.UserID == uid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot like your own review.",
		})

		return
	}

	if updatedReview, err = reviewModel.VoteReview(uid, data, review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(r.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.ReviewLogType,
		LogAction:        models.UpdateLogAction,
		LogActionDetails: "Vote",
		ContentID:        review.ContentID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Successfully voted.", "data": updatedReview})

}

// Delete Review
// @Summary Delete Review
// @Description Deletes Review
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review [delete]
func (r *ReviewController) DeleteReviewByID(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	reviewModel := models.NewReviewModel(r.Database)

	isDeleted, err := reviewModel.DeleteReviewByID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		logModel := models.NewLogsModel(r.Database)

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.ReviewLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
		})

		c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully."})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}
