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

// Get Preview Movies
// @Summary Get Preview Movies
// @Description Returns preview movies
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/preview [get]
func (m *MovieController) GetPreviewMovies(c *gin.Context) {
	movieModel := models.NewMovieModel(m.Database)

	upcomingMovies, _, err := movieModel.GetUpcomingMoviesBySort(requests.SortUpcoming{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	popularMovies, _, err := movieModel.GetMoviesBySortAndFilter(requests.SortFilterMovie{
		Sort: "popularity",
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	topMovies, _, err := movieModel.GetTopRatedMoviesBySort(requests.Pagination{
		Page: 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"upcoming": upcomingMovies, "popular": popularMovies, "top": topMovies})
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

// Get Top Rated Movies
// @Summary Get Top Rated Movies
// @Description Returns top rated movies with pagination
// @Tags movie
// @Accept application/json
// @Produce application/json
// @Param pagination body requests.Pagination true "Pagination"
// @Success 200 {array} responses.Movie
// @Failure 500 {string} string
// @Router /movie/top [get]
func (m *MovieController) GetTopRatedMoviesBySort(c *gin.Context) {
	var data requests.Pagination
	if err := c.ShouldBindQuery(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validatorErrorHandler(err),
		})

		return
	}

	movieModel := models.NewMovieModel(m.Database)

	topRatedMovies, pagination, err := movieModel.GetTopRatedMoviesBySort(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"pagination": pagination, "data": topRatedMovies})
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
