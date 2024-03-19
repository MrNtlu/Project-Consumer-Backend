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

type SocialResult struct {
	PopularReviews []responses.ReviewDetails
	CustomLists    []responses.CustomList
	Leaderboard    []responses.Leaderboard
	Error          error
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

	uid, ok := c.Get("uuid")

	resultCh := make(chan SocialResult)

	go func() {
		defer close(resultCh)

		socialResult := SocialResult{}

		reviewCh := make(chan []responses.ReviewDetails)
		go func() {
			var (
				err            error
				popularReviews []responses.ReviewDetails
			)
			sortRequest := requests.SortReview{
				Sort: "popularity",
				Page: 1,
			}

			if ok && uid != nil {
				userId := uid.(string)
				popularReviews, _, err = reviewModel.GetReviewsIndependentFromContent(&userId, sortRequest)
				if err != nil {
					resultCh <- SocialResult{Error: err}
					return
				}
			} else {
				popularReviews, _, err = reviewModel.GetReviewsIndependentFromContent(nil, sortRequest)
				if err != nil {
					resultCh <- SocialResult{Error: err}
					return
				}
			}

			reviewCh <- popularReviews
		}()

		customListCh := make(chan []responses.CustomList)
		go func() {
			var (
				err         error
				customLists []responses.CustomList
			)
			sortCustomListUID := requests.SortCustomListUID{
				Sort: "popularity",
			}

			if ok && uid != nil {
				userId := uid.(string)
				customLists, err = customListModel.GetCustomListsByUserID(&userId, sortCustomListUID, true, true)
				if err != nil {
					resultCh <- SocialResult{Error: err}
					return
				}
			} else {
				customLists, err = customListModel.GetCustomListsByUserID(nil, sortCustomListUID, true, true)
				if err != nil {
					resultCh <- SocialResult{Error: err}
					return
				}
			}
			customListCh <- customLists
		}()

		leaderboardCh := make(chan []responses.Leaderboard)
		go func() {
			var err error

			leaderboard, err := userModel.GetLeaderboard()
			if err != nil {
				resultCh <- SocialResult{Error: err}
				return
			}

			leaderboardCh <- leaderboard
		}()

		socialResult.PopularReviews = <-reviewCh
		socialResult.CustomLists = <-customListCh
		socialResult.Leaderboard = <-leaderboardCh

		resultCh <- socialResult
	}()

	result := <-resultCh
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": responses.SocialPreview{
		Reviews:     result.PopularReviews,
		Leaderboard: result.Leaderboard,
		CustomLists: result.CustomLists,
	}})
}
