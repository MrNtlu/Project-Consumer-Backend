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
	PopularReviews  []responses.ReviewDetails
	Recommendations []responses.RecommendationWithContent
	CustomLists     []responses.CustomList
	Leaderboard     []responses.Leaderboard
	Error           error
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
	userModel := models.NewUserModel(s.Database)

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

			reviewModel := models.NewReviewModel(s.Database)
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

			customListModel := models.NewCustomListModel(s.Database)
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

		recommendationCh := make(chan []responses.RecommendationWithContent)
		go func() {
			var (
				err             error
				recommendations []responses.RecommendationWithContent
			)
			sortRecommendations := requests.SortRecommendationsForSocial{
				Sort: "popularity",
				Page: 1,
			}

			recommendationModel := models.NewRecommendationModel(s.Database)
			if ok && uid != nil {
				var userId string = uid.(string)
				recommendations, _, err = recommendationModel.GetRecommendationsForSocial(userId, false, sortRecommendations)
				if err != nil {
					resultCh <- SocialResult{Error: err}
					return
				}
			} else {
				recommendations, _, err = recommendationModel.GetRecommendationsForSocial("", true, sortRecommendations)
				if err != nil {
					resultCh <- SocialResult{Error: err}
					return
				}
			}
			recommendationCh <- recommendations
		}()

		socialResult.PopularReviews = <-reviewCh
		socialResult.CustomLists = <-customListCh
		socialResult.Leaderboard = <-leaderboardCh
		socialResult.Recommendations = <-recommendationCh

		resultCh <- socialResult
	}()

	result := <-resultCh
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": responses.SocialPreview{
		Reviews:         result.PopularReviews,
		Leaderboard:     result.Leaderboard,
		CustomLists:     result.CustomLists,
		Recommendations: result.Recommendations,
	}})
}
