package controllers

import (
	"app/db"
	"app/models"
	"app/requests"
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
// @Param sortupcoming body requests.SortUpcoming true "Sort Upcoming"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/upcoming [get]
func (m *MovieController) GetUpcomingMoviesBySort(c *gin.Context) {
	var data requests.SortUpcoming
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

// Get Popular Movies by Decade
// @Summary Get Popular Movies by decade
// @Description Returns popular movies by decade with pagination
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param filterbydecade body requests.FilterByDecade true "Filter by Decade"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/decade [get]
func (m *MovieController) GetPopularMoviesByDecade(c *gin.Context) {
	var data requests.FilterByDecade
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	var dateTo = data.Decade + 10

	movies, pagination, err := movieModel.GetMoviesBySortAndFilter(requests.SortFilterMovie{
		ReleaseDateFrom: &data.Decade,
		ReleaseDateTo:   &dateTo,
		Sort:            "popularity",
		Page:            data.Page,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": movies})
}

// Get Popular Movies by Genre
// @Summary Get Popular Movies by genre
// @Description Returns popular movies by genre with pagination
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param filterbygenre body requests.FilterByGenre true "Filter by Genre"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/genre [get]
func (m *MovieController) GetPopularMoviesByGenre(c *gin.Context) {
	var data requests.FilterByGenre
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	movies, pagination, err := movieModel.GetMoviesBySortAndFilter(requests.SortFilterMovie{
		Genres: &data.Genre,
		Sort:   "popularity",
		Page:   data.Page,
	})
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

	movieDetails, err := movieModel.GetMovieDetails(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	uuid, OK := c.Get("uuid")
	if OK && uuid != nil && movieDetails.TitleOriginal != "" {
		movieDetailsWithWatchList, err := movieModel.GetMovieDetailsWithWatchListAndWatchLater(data, uuid.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": movieDetailsWithWatchList,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": movieDetails,
		})
	}
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
