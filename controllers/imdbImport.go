package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type IMDBImportController struct {
	Database *db.MongoDB
}

type TMDBImportController struct {
	Database *db.MongoDB
}

type AniListImportController struct {
	Database *db.MongoDB
}

type TraktImportController struct {
	Database *db.MongoDB
}

func NewIMDBImportController(mongoDB *db.MongoDB) IMDBImportController {
	return IMDBImportController{
		Database: mongoDB,
	}
}

func NewTMDBImportController(mongoDB *db.MongoDB) TMDBImportController {
	return TMDBImportController{
		Database: mongoDB,
	}
}

func NewAniListImportController(mongoDB *db.MongoDB) AniListImportController {
	return AniListImportController{
		Database: mongoDB,
	}
}

func NewTraktImportController(mongoDB *db.MongoDB) TraktImportController {
	return TraktImportController{
		Database: mongoDB,
	}
}

// Import IMDB Watchlist
// @Summary Import watchlist from IMDB
// @Description Imports user's watchlist from IMDB using their User ID or List ID
// @Tags import
// @Accept application/json
// @Produce application/json
// @Param imdbimport body requests.IMDBImportRequest true "IMDB Import Request"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.IMDBImportResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /import/imdb [post]
func (c *IMDBImportController) ImportWatchlist(ctx *gin.Context) {
	var req requests.IMDBImportRequest
	if shouldReturn := bindJSONData(&req, ctx); shouldReturn {
		return
	}

	if req.IMDBUserID == "" && req.IMDBListID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Either IMDB User ID or List ID must be provided",
		})
		return
	}

	uid := jwt.ExtractClaims(ctx)["id"].(string)
	imdbImportModel := models.NewIMDBImportModel(c.Database)

	logrus.WithFields(logrus.Fields{
		"user_id":      uid,
		"imdb_user_id": req.IMDBUserID,
		"imdb_list_id": req.IMDBListID,
	}).Info("IMDB import request received")

	result, err := imdbImportModel.ImportUserWatchlist(uid, req.IMDBUserID, req.IMDBListID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
			"error":   err.Error(),
		}).Error("IMDB import failed")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        uid,
		"imported_count": result.ImportedCount,
		"skipped_count":  result.SkippedCount,
		"error_count":    result.ErrorCount,
	}).Info("IMDB import completed successfully")

	ctx.JSON(http.StatusOK, gin.H{
		"message": "IMDB import completed successfully.",
		"data":    result,
	})
}

// Import TMDB Watchlist
// @Summary Import watchlist from TMDB
// @Description Imports user's watchlist and ratings from TMDB using their username
// @Tags import
// @Accept application/json
// @Produce application/json
// @Param tmdbimport body requests.TMDBImportRequest true "TMDB Import Request"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.TMDBImportResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /import/tmdb [post]
func (c *TMDBImportController) ImportUserData(ctx *gin.Context) {
	var req requests.TMDBImportRequest
	if shouldReturn := bindJSONData(&req, ctx); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(ctx)["id"].(string)
	tmdbImportModel := models.NewTMDBImportModel(c.Database)

	logrus.WithFields(logrus.Fields{
		"user_id":       uid,
		"tmdb_username": req.TMDBUsername,
	}).Info("TMDB import request received")

	result, err := tmdbImportModel.ImportUserData(uid, req.TMDBUsername)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
			"error":   err.Error(),
		}).Error("TMDB import failed")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        uid,
		"imported_count": result.ImportedCount,
		"skipped_count":  result.SkippedCount,
		"error_count":    result.ErrorCount,
	}).Info("TMDB import completed successfully")

	ctx.JSON(http.StatusOK, gin.H{
		"message": "TMDB import completed successfully.",
		"data":    result,
	})
}

// Import AniList Lists
// @Summary Import anime and manga lists from AniList
// @Description Imports user's anime and manga lists from AniList using their username
// @Tags import
// @Accept application/json
// @Produce application/json
// @Param anilistimport body requests.AniListImportRequest true "AniList Import Request"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.AniListImportResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /import/anilist [post]
func (c *AniListImportController) ImportUserLists(ctx *gin.Context) {
	var req requests.AniListImportRequest
	if shouldReturn := bindJSONData(&req, ctx); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(ctx)["id"].(string)
	anilistImportModel := models.NewAniListImportModel(c.Database)

	logrus.WithFields(logrus.Fields{
		"user_id":          uid,
		"anilist_username": req.AniListUsername,
	}).Info("AniList import request received")

	result, err := anilistImportModel.ImportUserLists(uid, req.AniListUsername)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
			"error":   err.Error(),
		}).Error("AniList import failed")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        uid,
		"imported_count": result.ImportedCount,
		"skipped_count":  result.SkippedCount,
		"error_count":    result.ErrorCount,
	}).Info("AniList import completed successfully")

	ctx.JSON(http.StatusOK, gin.H{
		"message": "AniList import completed successfully.",
		"data":    result,
	})
}

// Import Trakt Data
// @Summary Import watched movies and shows from Trakt
// @Description Imports user's watched movies, shows, and watchlist from Trakt using their username
// @Tags import
// @Accept application/json
// @Produce application/json
// @Param traktimport body requests.TraktImportRequest true "Trakt Import Request"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {object} responses.TraktImportResponse
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /import/trakt [post]
func (c *TraktImportController) ImportUserData(ctx *gin.Context) {
	var req requests.TraktImportRequest
	if shouldReturn := bindJSONData(&req, ctx); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(ctx)["id"].(string)
	traktImportModel := models.NewTraktImportModel(c.Database)

	logrus.WithFields(logrus.Fields{
		"user_id":        uid,
		"trakt_username": req.TraktUsername,
	}).Info("Trakt import request received")

	result, err := traktImportModel.ImportUserData(uid, req.TraktUsername)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": uid,
			"error":   err.Error(),
		}).Error("Trakt import failed")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        uid,
		"imported_count": result.ImportedCount,
		"skipped_count":  result.SkippedCount,
		"error_count":    result.ErrorCount,
	}).Info("Trakt import completed successfully")

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Trakt import completed successfully.",
		"data":    result,
	})
}
