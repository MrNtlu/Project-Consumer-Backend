package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
	"app/responses"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MovieController struct {
	Database *db.MongoDB
}

func NewMovieController(mongoDB *db.MongoDB) MovieController {
	return MovieController{
		Database: mongoDB,
	}
}

// Get Upcoming Movies
// @Summary Get Upcoming Movies by Sort
// @Description Returns upcoming movies by sort with pagination
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/upcoming [get]
func (m *MovieController) GetUpcomingMoviesBySort(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	upcomingMovies, pagination, err := movieModel.GetUpcomingMoviesBySort(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": upcomingMovies})
}

// Get Movies in Theater
// @Summary Get Movies that are in theaters
// @Description Returns movies in theaters
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/theaters [get]
func (m *MovieController) GetMoviesInTheater(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	movies, pagination, err := movieModel.GetMoviesInTheater(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": movies})
}

// Get Movies
// @Summary Get Movies by Sort and Filter
// @Description Returns movies by sort and filter with pagination
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param sortfiltermovie body requests.SortFilterMovie true "Sort and Filter Movie"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie [get]
func (m *MovieController) GetMoviesBySortAndFilter(c *gin.Context) {
	var data requests.SortFilterMovie
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	movies, pagination, err := movieModel.GetMoviesBySortAndFilter(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": movies})
}

// Get Movie Details
// @Summary Get Movie Details
// @Description Returns movie details with optional authentication
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param id body requests.ID true "ID"
// @Success 200 {array} responses.Movie
// @Success 200 {array} responses.MovieDetails
// @Failure 500 {string} string
// @Router /movie/details [get]
func (m *MovieController) GetMovieDetails(c *gin.Context) {
	var data requests.ID
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)
	reviewModel := models.NewReviewModel(m.Database)

	uid, OK := c.Get("uuid")
	if OK && uid != nil {
		userID := uid.(string)

		movieDetailsWithWatchList, err := movieModel.GetMovieDetailsWithWatchListAndWatchLater(data, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if movieDetailsWithWatchList.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, userID, &movieDetailsWithWatchList.TmdbID, nil)
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
		movieDetailsWithWatchList.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": movieDetailsWithWatchList,
		})
	} else {
		movieDetails, err := movieModel.GetMovieDetails(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if movieDetails.TitleOriginal == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": ErrNotFound})
			return
		}

		reviewSummary, err := reviewModel.GetReviewSummaryForDetails(data.ID, "-1", &movieDetails.TmdbID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		movieDetails.Review = reviewSummary

		c.JSON(http.StatusOK, gin.H{
			"data": movieDetails,
		})
	}
}

// Get Popular Actors
// @Summary Get Popular Actors
// @Description Returns Popular Actors
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.ActorDetails
// @Failure 500 {string} string
// @Router /movie/popular-actors [get]
func (m *MovieController) GetPopularActors(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	actors, err := movieModel.GetPopularActors(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": actors})
}

// Get Movies by Actors
// @Summary Get Movies by Actors
// @Description Returns Movies by Actors
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param idpagination body requests.IDPagination true "ID Pagination"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/actor [get]
func (m *MovieController) GetMoviesByActor(c *gin.Context) {
	var data requests.IDPagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	movies, pagination, err := movieModel.GetMoviesByActor(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": movies})
}

// Search Movie
// @Summary Search Movie
// @Description Search movies
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param search body requests.Search true "Search"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/search [get]
func (m *MovieController) SearchMovieByTitle(c *gin.Context) {
	var data requests.Search
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	movies, pagination, err := movieModel.SearchMovieByTitle(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": movies})
}
