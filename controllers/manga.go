package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MangaController struct {
	Database *db.MongoDB
}

func NewMangaController(mongoDB *db.MongoDB) MangaController {
	return MangaController{
		Database: mongoDB,
	}
}

// Get Mangas
// @Summary Get Manga by Sort and Filter
// @Description Returns manga by sort and filter
// @Tags manga
// @Accept application/json
// @Produce application/json
// @Param sortfiltermanga body requests.SortFilterManga true "Sort and Filter Manga"
// @Success 200 {array} responses.Manga
// @Failure 500 {string} string
// @Router /manga [get]
func (m *MangaController) GetMangaBySortAndFilter(c *gin.Context) {
	var data requests.SortFilterManga
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	mangaModel := models.NewMangaModel(m.Database)

	mangas, pagination, err := mangaModel.GetMangaBySortAndFilter(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": mangas})
}

// Get Manga Details
// @Summary Get Manga Details
// @Description Returns Manga details with optional authentication
// @Tags manga
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Success 200 {array} responses.Manga
// @Success 200 {array} responses.MangaDetails
// @Failure 500 {string} string
// @Router /manga/details [get]
func (m *MangaController) GetMangaDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	mangaModel := models.NewMangaModel(m.Database)
	reviewModel := models.NewReviewModel(m.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userID := uid.(string)

		mangaDetailsWithWatchList, err := mangaModel.GetMangaDetailsWithWatchList(data, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if mangaDetailsWithWatchList.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, userID, nil, &mangaDetailsWithWatchList.MalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		var review *responses.Review

		if reviewSummary.IsReviewed {
			reviewResponse, _ := reviewModel.GetBaseReviewResponseByUserIDAndContentID(data.ID, userID)
			review = &reviewResponse
		} else {
			review = nil
		}

		reviewSummary.Review = review
		mangaDetailsWithWatchList.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": mangaDetailsWithWatchList,
		})
	} else {
		mangaDetails, err := mangaModel.GetMangaDetails(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if mangaDetails.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, "-1", nil, &mangaDetails.MalID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		mangaDetails.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": mangaDetails,
		})
	}
}
