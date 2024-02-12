package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SocialController struct {
	Database *db.MongoDB
}

func NewSocialController(mongoDB *db.MongoDB) SocialController {
	return SocialController{
		Database: mongoDB,
	}
}

// Get Socials
// @Summary Get Socials
// @Description Returns reviews, custom lists and leaderboard for social page.
// @Tags social
// @Accept application/json
// @Produce application/json
// @Success 200 {object} responses.SocialPreview
// @Failure 500 {string} string
// @Router /social [get]
func (s *SocialController) GetSocials(c *gin.Context) {
	reviewModel := models.NewReviewModel(s.Database)
	userModel := models.NewUserModel(s.Database)
	customListModel := models.NewCustomListModel(s.Database)

	var (
		popularReviews []responses.ReviewDetails
		customLists    []responses.CustomList
		err            error
	)

	sortRequest := requests.SortReview{
		Sort: "popularity",
		Page: 1,
	}

	uid, OK := c.Get("uuid")

	if OK && uid != nil {
		userId := uid.(string)

		popularReviews, _, err = reviewModel.GetReviewsIndependentFromContent(&userId, sortRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	} else {
		popularReviews, _, err = reviewModel.GetReviewsIndependentFromContent(nil, sortRequest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	if OK && uid != nil {
		userId := uid.(string)

		customLists, err = customListModel.GetCustomListsByUserID(&userId, requests.SortCustomList{
			Sort: "popularity",
		}, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	} else {
		customLists, err = customListModel.GetCustomListsByUserID(nil, requests.SortCustomList{
			Sort: "popularity",
		}, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}
	}

	leaderboard, err := userModel.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": responses.SocialPreview{
		Reviews:     popularReviews,
		Leaderboard: leaderboard,
		CustomLists: customLists,
	}})
}
