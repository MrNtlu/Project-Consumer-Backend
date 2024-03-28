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
	createdReview.Author.IsPremium = author.IsPremium || author.IsLifetimePremium

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

// Get Review List by UID
// @Summary Get Review List by UID
// @Description Get Review List by UID
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param sortreview body requests.SortReview true "Sort Review"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.ReviewDetails
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/profile [get]
func (r *ReviewController) GetReviewsByUID(c *gin.Context) {
	var data requests.SortReview
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	reviewModel := models.NewReviewModel(r.Database)

	uid := jwt.ExtractClaims(c)["id"].(string)

	reviews, pagination, err := reviewModel.GetReviewsByUID(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reviews, "pagination": pagination})
}

// Get Review List
// @Summary Get Review List
// @Description Get Review List independent from content or user id
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param sortreview body requests.SortReview true "Sort Review"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.ReviewDetails
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/social [get]
func (r *ReviewController) GetReviewsIndependentFromContent(c *gin.Context) {
	var data requests.SortReview
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	reviewModel := models.NewReviewModel(r.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userId := uid.(string)

		reviews, pagination, err := reviewModel.GetReviewsIndependentFromContent(&userId, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": reviews, "pagination": pagination})
	} else {
		reviews, pagination, err := reviewModel.GetReviewsIndependentFromContent(nil, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": reviews, "pagination": pagination})
	}
}

// Get Review Details
// @Summary Get Review Details
// @Description Get Review Details with or without authentication
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.ReviewDetails
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/details [get]
func (r *ReviewController) GetReviewDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	reviewModel := models.NewReviewModel(r.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userId := uid.(string)

		reviews, err := reviewModel.GetReviewDetails(&userId, data.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": reviews})
	} else {
		reviews, err := reviewModel.GetReviewDetails(nil, data.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"data": reviews})
	}
}

// Get Reviews by User
// @Summary Get Reviews by User
// @Description Get Reviews by username with or without authentication
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param sortreviewbyusername body requests.SortReviewByUsername true "Sort Review by Username"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.ReviewWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/details [get]
func (r *ReviewController) GetReviewsByUsername(c *gin.Context) {
	var data requests.SortReviewByUsername
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	userModel := models.NewUserModel(r.Database)
	reviewModel := models.NewReviewModel(r.Database)

	userInfo, err := userModel.GetUserInfo(data.Username, "", true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if userInfo.EmailAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errNoUser,
		})

		return
	}

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userId := uid.(string)

		reviews, pagination, err := reviewModel.GetReviewsByUserID(&userId, requests.SortReviewByUserID{
			UserID: userInfo.ID.Hex(),
			Sort:   data.Sort,
			Page:   data.Page,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": reviews})
	} else {
		reviews, pagination, err := reviewModel.GetReviewsByUserID(nil, requests.SortReviewByUserID{
			UserID: userInfo.ID.Hex(),
			Sort:   data.Sort,
			Page:   data.Page,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": reviews})
	}
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

// Get Liked Reviews
// @Summary Get Liked by User
// @Description Get Liked by User
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param sortreview body requests.SortReview true "Sort Review"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.ReviewWithContent
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/liked [get]
func (r *ReviewController) GetLikedReviews(c *gin.Context) {
	var data requests.SortReview
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)
	reviewModel := models.NewReviewModel(r.Database)

	reviews, pagination, err := reviewModel.GetLikedReviews(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": reviews})
}

// Get User Reviews
// @Summary Get Reviews by User
// @Description Get Reviews by User
// @Tags review
// @Accept application/json
// @Produce application/json
// @Param sortreviewbyuserid body requests.SortReviewByUserID true "Sort Review by User ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.Review
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /review/user [get]
func (r *ReviewController) GetReviewsByUserID(c *gin.Context) {
	var data requests.SortReviewByUserID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	reviewModel := models.NewReviewModel(r.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userId := uid.(string)

		reviews, pagination, err := reviewModel.GetReviewsByUserID(&userId, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": reviews})
	} else {
		reviews, pagination, err := reviewModel.GetReviewsByUserID(nil, data)
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated.", "data": updatedReview})
}

// Like/Dislike Review
// @Summary Like/Dislike Review
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
// @Router /review/like [patch]
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
		ContentType:      review.ContentType,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Successfully liked.", "data": updatedReview})

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
