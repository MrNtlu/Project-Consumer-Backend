package routes

import (
	"app/controllers"
	"app/db"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func searchRouter(
	router *gin.RouterGroup,
	mongoDB *db.MongoDB,
	pinecone *pinecone.Client,
	pineconeIndex *pinecone.IndexConnection,
	redisClient *redis.Client,
) {
	pineconeController := controllers.NewPineconeController(mongoDB, pinecone, pineconeIndex, redisClient)

	searchGroup := router.Group("/search")
	{
		searchGroup.GET("/content", func(c *gin.Context) {
			handleSearchContent(c, &pineconeController)
		})
		searchGroup.GET("/cache/stats", func(c *gin.Context) {
			handleCacheStats(c, &pineconeController)
		})
		searchGroup.DELETE("/cache", func(c *gin.Context) {
			handleClearCache(c, &pineconeController)
		})
	}
}

func handleSearchContent(c *gin.Context, pineconeController *controllers.PineconeController) {
	// Get query parameters
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	contentType := c.Query("type")
	if contentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'type' is required (movie, tvseries, anime, game)",
		})
		return
	}

	// Validate content type
	validTypes := map[string]bool{
		"movie":    true,
		"tvseries": true,
		"anime":    true,
		"game":     true,
	}
	if !validTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid content type. Must be one of: movie, tvseries, anime, game",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	logrus.WithFields(logrus.Fields{
		"query":       query,
		"contentType": contentType,
		"page":        page,
		"limit":       limit,
	}).Info("processing search request")

	// Perform search using Pinecone
	results, totalResults, err := pineconeController.SearchContentByQuery(query, contentType, page, limit)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"query":       query,
			"contentType": contentType,
			"error":       err,
		}).Error("search failed")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed. Please try again.",
		})
		return
	}

	// Format results to match your existing response format
	formattedResults := make([]map[string]interface{}, 0, len(results))
	for _, result := range results {
		if data, ok := result["data"].(map[string]interface{}); ok {
			// Create a clean response similar to your existing format
			cleanResult := map[string]interface{}{
				"_id":              data["_id"],
				"title_en":         data["title_en"],
				"title_original":   data["title_original"],
				"image_url":        data["image_url"],
				"description":      data["description"],
				"search_score":     result["hybrid_score"],
				"vector_score":     result["vector_score"],
				"popularity_score": result["popularity_score"],
				"title_relevance":  result["title_relevance"],
			}

			// Add content-type specific fields
			switch contentType {
			case "movie":
				if tmdbID, ok := data["tmdb_id"]; ok {
					cleanResult["tmdb_id"] = tmdbID
				}
				if vote, ok := data["tmdb_vote"]; ok {
					cleanResult["tmdb_vote"] = vote
				}
				if releaseDate, ok := data["release_date"]; ok {
					cleanResult["release_date"] = releaseDate
				}
			case "tvseries":
				if tmdbID, ok := data["tmdb_id"]; ok {
					cleanResult["tmdb_id"] = tmdbID
				}
				if vote, ok := data["tmdb_vote"]; ok {
					cleanResult["tmdb_vote"] = vote
				}
				if firstAirDate, ok := data["first_air_date"]; ok {
					cleanResult["first_air_date"] = firstAirDate
				}
				if totalSeasons, ok := data["total_seasons"]; ok {
					cleanResult["total_seasons"] = totalSeasons
				}
			case "anime":
				if malID, ok := data["mal_id"]; ok {
					cleanResult["mal_id"] = malID
				}
				if malScore, ok := data["mal_score"]; ok {
					cleanResult["mal_score"] = malScore
				}
				if titleJp, ok := data["title_jp"]; ok {
					cleanResult["title_jp"] = titleJp
				}
				if episodes, ok := data["episodes"]; ok {
					cleanResult["episodes"] = episodes
				}
			case "game":
				if rawgID, ok := data["rawg_id"]; ok {
					cleanResult["rawg_id"] = rawgID
				}
				if rawgRating, ok := data["rawg_rating"]; ok {
					cleanResult["rawg_rating"] = rawgRating
				}
				if releaseDate, ok := data["release_date"]; ok {
					cleanResult["release_date"] = releaseDate
				}
				if metacriticScore, ok := data["metacritic_score"]; ok {
					cleanResult["metacritic_score"] = metacriticScore
				}
			}

			formattedResults = append(formattedResults, cleanResult)
		}
	}

	// Calculate pagination info
	totalPages := (totalResults + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	// Create response
	response := map[string]interface{}{
		"data": formattedResults,
		"pagination": map[string]interface{}{
			"current_page":  page,
			"per_page":      limit,
			"total_results": totalResults,
			"total_pages":   totalPages,
			"has_next":      hasNext,
			"has_previous":  hasPrev,
		},
		"search_meta": map[string]interface{}{
			"query":        query,
			"content_type": contentType,
			"search_type":  "semantic",
		},
	}

	logrus.WithFields(logrus.Fields{
		"query":         query,
		"contentType":   contentType,
		"page":          page,
		"limit":         limit,
		"totalResults":  totalResults,
		"returnedCount": len(formattedResults),
	}).Info("search completed successfully")

	c.JSON(http.StatusOK, response)
}

func handleCacheStats(c *gin.Context, pineconeController *controllers.PineconeController) {
	stats := pineconeController.GetCacheStats()

	logrus.WithFields(logrus.Fields{
		"embedding_cache_size": stats["embedding_cache_size"],
		"search_cache_size":    stats["search_cache_size"],
	}).Info("cache stats requested")

	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
	})
}

func handleClearCache(c *gin.Context, pineconeController *controllers.PineconeController) {
	pineconeController.ClearCache()

	logrus.Info("cache cleared via API")

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}
