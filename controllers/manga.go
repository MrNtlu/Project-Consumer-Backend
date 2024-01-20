package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
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
// @Router /anime [get]
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
