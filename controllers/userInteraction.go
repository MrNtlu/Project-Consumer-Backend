package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type UserInteractionController struct {
	Database *db.MongoDB
}

func NewUserInteractionController(mongoDB *db.MongoDB) UserInteractionController {
	return UserInteractionController{
		Database: mongoDB,
	}
}

const errConsumeLaterPremium = "Free members can add up to 50 content to their watch later, you can get premium membership for unlimited access."

// Create Consume Later
// @Summary Create Consume Later
// @Description Creates Consume Later
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Param createconsumelater body requests.CreateConsumeLater true "Create Consume Later"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 201 {object} models.ConsumeLaterList
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /consume [post]
func (ui *UserInteractionController) CreateConsumeLater(c *gin.Context) {
	var data requests.CreateConsumeLater
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		contentTitle string
		contentImage string
	)

	switch data.ContentType {
	case "anime":
		animeModel := models.NewAnimeModel(ui.Database)

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
		gameModel := models.NewGameModel(ui.Database)

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
		movieModel := models.NewMovieModel(ui.Database)

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
		tvSeriesModel := models.NewTVModel(ui.Database)

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
	userModel := models.NewUserModel(ui.Database)
	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count := userInteractionModel.GetConsumeLaterCount(uid)

	if !isPremium && count >= models.ConsumeLaterLimit {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errConsumeLaterPremium,
		})

		return
	}

	var (
		createdConsumeLater models.ConsumeLaterList
		err                 error
	)

	if createdConsumeLater, err = userInteractionModel.CreateConsumeLater(uid, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	logModel := models.NewLogsModel(ui.Database)

	go logModel.CreateLog(uid, requests.CreateLog{
		LogType:          models.ConsumeLaterLogType,
		LogAction:        models.AddLogAction,
		LogActionDetails: "",
		ContentTitle:     contentTitle,
		ContentImage:     contentImage,
		ContentType:      data.ContentType,
		ContentID:        data.ContentID,
	})

	// Check watch later achievements in background
	achievementModel := models.NewAchievementModel(ui.Database)
	achievementModel.CheckAndUnlockAchievements(uid, "watch_later")
	achievementModel.CheckAndUnlockAchievements(uid, "first_activity")

	c.JSON(http.StatusCreated, gin.H{"message": "Successfully created.", "data": createdConsumeLater})
}

// Move Consume Later List as User List
// @Summary Move consume later as user list
// @Description Deletes consume later and creates user list
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Param markconsumelater body requests.MarkConsumeLater true "Mark Consume Later"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /consume/move [post]
func (ui *UserInteractionController) MarkConsumeLaterAsUserList(c *gin.Context) {
	var data requests.MarkConsumeLater
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	var (
		contentTitle string
		contentImage string
	)

	uid := jwt.ExtractClaims(c)["id"].(string)
	logModel := models.NewLogsModel(ui.Database)
	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	consumeLater, err := userInteractionModel.GetBaseConsumeLater(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if consumeLater.ContentID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
		return
	}

	userModel := models.NewUserModel(ui.Database)
	userListModel := models.NewUserListModel(ui.Database)

	isPremium, _ := userModel.IsUserPremium(uid)
	count, _ := userListModel.GetUserListCount(uid)

	if !isPremium && count >= models.UserListLimit {
		c.JSON(http.StatusForbidden, gin.H{
			"error": errUserListPremium,
		})

		return
	}

	var timesFinished = 1
	var status = "finished"

	switch consumeLater.ContentType {
	case "anime":
		animeModel := models.NewAnimeModel(ui.Database)
		anime, err := animeModel.GetAnimeDetails(requests.ID{
			ID: consumeLater.ContentID,
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

		animeList := userListModel.GetAnimeListByUserIdAndAnimeId(uid, consumeLater.ContentID)

		var episodes int64
		if anime.Episodes != nil {
			episodes = *anime.Episodes
		} else {
			episodes = 0
		}

		if animeList.UserID != "" {
			timesFinished = animeList.TimesFinished + 1

			var score float32
			if data.Score == nil && animeList.Score != nil {
				score = *animeList.Score
			} else {
				score = *data.Score
			}

			if _, err := userListModel.UpdateAnimeListByID(animeList, requests.UpdateAnimeList{
				ID:              consumeLater.ContentID,
				IsUpdatingScore: true,
				Score:           &score,
				TimesFinished:   &timesFinished,
				Status:          &status,
				WatchedEpisodes: &episodes,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		} else {
			if _, err = userListModel.CreateAnimeList(uid, requests.CreateAnimeList{
				AnimeID:         consumeLater.ContentID,
				AnimeMALID:      anime.MalID,
				Status:          status,
				TimesFinished:   &timesFinished,
				WatchedEpisodes: &episodes,
				Score:           data.Score,
			}, anime); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		}

		userInteractionModel.DeleteConsumeLaterByID(uid, data.ID)

		contentTitle = anime.TitleOriginal
		contentImage = anime.ImageURL

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.ConsumeLaterLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})
		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.UserListLogType,
			LogAction:        models.AddLogAction,
			LogActionDetails: models.FinishedActionDetails,
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})

		// Check achievements for finished anime
		achievementModel := models.NewAchievementModel(ui.Database)
		achievementModel.CheckAndUnlockAchievements(uid, "anime_finished")

		c.JSON(http.StatusCreated, gin.H{"message": "Successfully moved to user list."})
		return
	case "game":
		gameModel := models.NewGameModel(ui.Database)
		game, err := gameModel.GetGameDetails(requests.ID{
			ID: consumeLater.ContentID,
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

		gameList := userListModel.GetGameListByUserIdAndGameId(uid, consumeLater.ContentID)

		if gameList.UserID != "" {
			timesFinished = gameList.TimesFinished + 1

			var score float32
			if data.Score == nil && gameList.Score != nil {
				score = *gameList.Score
			} else {
				score = *data.Score
			}

			if _, err := userListModel.UpdateGameListByID(gameList, requests.UpdateGameList{
				ID:              consumeLater.ContentID,
				IsUpdatingScore: true,
				Score:           &score,
				TimesFinished:   &timesFinished,
				Status:          &status,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		} else {
			if _, err = userListModel.CreateGameList(uid, requests.CreateGameList{
				GameID:        consumeLater.ContentID,
				GameRAWGID:    *consumeLater.ContentExternalIntID,
				Status:        status,
				Score:         data.Score,
				TimesFinished: &timesFinished,
			}, game); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		}

		userInteractionModel.DeleteConsumeLaterByID(uid, data.ID)

		contentTitle = game.Title
		contentImage = game.ImageUrl

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.ConsumeLaterLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})
		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.UserListLogType,
			LogAction:        models.AddLogAction,
			LogActionDetails: models.FinishedActionDetails,
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})

		// Check achievements for finished game
		achievementModel := models.NewAchievementModel(ui.Database)
		achievementModel.CheckAndUnlockAchievements(uid, "game_finished")

		c.JSON(http.StatusCreated, gin.H{"message": "Successfully moved to user list."})
		return
	case "movie":
		movieModel := models.NewMovieModel(ui.Database)

		movie, err := movieModel.GetMovieDetails(requests.ID{
			ID: consumeLater.ContentID,
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

		movieList := userListModel.GetMovieListByUserIdAndMovieId(uid, consumeLater.ContentID)

		if movieList.UserID != "" {
			timesFinished = movieList.TimesFinished + 1

			var score float32
			if data.Score == nil && movieList.Score != nil {
				score = *movieList.Score
			} else {
				score = *data.Score
			}

			if _, err := userListModel.UpdateMovieListByID(movieList, requests.UpdateMovieList{
				ID:              consumeLater.ContentID,
				IsUpdatingScore: true,
				Score:           &score,
				TimesFinished:   &timesFinished,
				Status:          &status,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		} else {
			if _, err = userListModel.CreateMovieWatchList(uid, requests.CreateMovieWatchList{
				MovieID:     consumeLater.ContentID,
				MovieTmdbID: *consumeLater.ContentExternalID,
				Status:      "finished",
				Score:       data.Score,
			}, movie); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		}

		userInteractionModel.DeleteConsumeLaterByID(uid, data.ID)

		contentTitle = movie.TitleEn
		contentImage = movie.ImageURL

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.ConsumeLaterLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})
		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.UserListLogType,
			LogAction:        models.AddLogAction,
			LogActionDetails: models.FinishedActionDetails,
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})

		// Check achievements for finished movie
		achievementModel := models.NewAchievementModel(ui.Database)
		achievementModel.CheckAndUnlockAchievements(uid, "movie_finished")

		c.JSON(http.StatusCreated, gin.H{"message": "Successfully moved to user list."})
		return
	case "tv":
		tvSeriesModel := models.NewTVModel(ui.Database)
		tvSeries, err := tvSeriesModel.GetTVSeriesDetails(requests.ID{
			ID: consumeLater.ContentID,
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

		tvList := userListModel.GetTVSeriesListByUserIdAndTVId(uid, consumeLater.ContentID)

		episodes := tvSeries.TotalEpisodes
		seasons := tvSeries.TotalSeasons

		if tvList.UserID != "" {
			timesFinished = tvList.TimesFinished + 1

			var score float32
			if data.Score == nil && tvList.Score != nil {
				score = *tvList.Score
			} else {
				score = *data.Score
			}

			if _, err := userListModel.UpdateTVSeriesListByID(tvList, requests.UpdateTVSeriesList{
				ID:              consumeLater.ContentID,
				IsUpdatingScore: true,
				Score:           &score,
				TimesFinished:   &timesFinished,
				Status:          &status,
				WatchedEpisodes: &episodes,
				WatchedSeasons:  &seasons,
			}); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		} else {
			if _, err = userListModel.CreateTVSeriesWatchList(uid, requests.CreateTVSeriesWatchList{
				TvID:            consumeLater.ContentID,
				TvTmdbID:        tvSeries.TmdbID,
				Status:          "finished",
				TimesFinished:   &timesFinished,
				WatchedEpisodes: &episodes,
				WatchedSeasons:  &seasons,
				Score:           data.Score,
			}, tvSeries); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})

				return
			}
		}

		userInteractionModel.DeleteConsumeLaterByID(uid, data.ID)

		contentTitle = tvSeries.TitleEn
		contentImage = tvSeries.ImageURL

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.ConsumeLaterLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})
		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.UserListLogType,
			LogAction:        models.AddLogAction,
			LogActionDetails: models.FinishedActionDetails,
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})

		// Check achievements for finished TV series
		achievementModel := models.NewAchievementModel(ui.Database)
		achievementModel.CheckAndUnlockAchievements(uid, "tv_finished")

		c.JSON(http.StatusCreated, gin.H{"message": "Successfully moved to user list."})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown error!"})
}

// Get Consume Later List
// @Summary Get Consume Later
// @Description Returns Consume Later by optional filter
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Param sortfilterconsumelater body requests.SortFilterConsumeLater true "Sort Filter Consume Later"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {array} responses.ConsumeLater
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /consume [get]
func (ui *UserInteractionController) GetConsumeLater(c *gin.Context) {
	var data requests.SortFilterConsumeLater
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	consumeLaterList, err := userInteractionModel.GetConsumeLater(uid, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": consumeLaterList})
}

// Delete Consume Later
// @Summary Delete Consume Later
// @Description Deletes Consume Later
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /consume [delete]
func (ui *UserInteractionController) DeleteConsumeLaterById(c *gin.Context) {
	var data requests.ID
	if shouldReturn := bindJSONData(&data, c); shouldReturn {
		return
	}

	uid := jwt.ExtractClaims(c)["id"].(string)

	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	consumeLater, _ := userInteractionModel.GetBaseConsumeLater(uid, data.ID)
	isDeleted, err := userInteractionModel.DeleteConsumeLaterByID(uid, data.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if isDeleted {
		var (
			contentTitle string
			contentImage string
		)

		switch consumeLater.ContentType {
		case "anime":
			animeModel := models.NewAnimeModel(ui.Database)

			anime, _ := animeModel.GetAnimeDetails(requests.ID{
				ID: consumeLater.ContentID,
			})

			contentTitle = anime.TitleOriginal
			contentImage = anime.ImageURL
		case "game":
			gameModel := models.NewGameModel(ui.Database)

			game, _ := gameModel.GetGameDetails(requests.ID{
				ID: consumeLater.ContentID,
			})

			contentTitle = game.Title
			contentImage = game.ImageUrl
		case "movie":
			movieModel := models.NewMovieModel(ui.Database)

			movie, _ := movieModel.GetMovieDetails(requests.ID{
				ID: consumeLater.ContentID,
			})

			contentTitle = movie.TitleEn
			contentImage = movie.ImageURL
		case "tv":
			tvSeriesModel := models.NewTVModel(ui.Database)

			tvSeries, _ := tvSeriesModel.GetTVSeriesDetails(requests.ID{
				ID: consumeLater.ContentID,
			})

			contentTitle = tvSeries.TitleEn
			contentImage = tvSeries.ImageURL
		}

		logModel := models.NewLogsModel(ui.Database)

		go logModel.CreateLog(uid, requests.CreateLog{
			LogType:          models.ConsumeLaterLogType,
			LogAction:        models.DeleteLogAction,
			LogActionDetails: "",
			ContentTitle:     contentTitle,
			ContentImage:     contentImage,
			ContentType:      consumeLater.ContentType,
			ContentID:        consumeLater.ContentID,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully."})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
}

// Delete All Consume Later
// @Summary Delete All Consume Later
// @Description Deletes All Consume Later
// @Tags consume_later
// @Accept application/json
// @Produce application/json
// @Security BearerAuth
// @Param Authorization header string true "Authentication header"
// @Success 200 {string} string
// @Failure 500 {string} string
// @Router /consume/all [delete]
func (ui *UserInteractionController) DeleteAllConsumeLaterByUserID(c *gin.Context) {
	uid := jwt.ExtractClaims(c)["id"].(string)

	userInteractionModel := models.NewUserInteractionModel(ui.Database)

	if err := userInteractionModel.DeleteAllConsumeLaterByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Deleted successfully."})
}
