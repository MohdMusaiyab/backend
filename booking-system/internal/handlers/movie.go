package handlers

import (
	"booking-system/internal/models"
	"booking-system/pkg/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateMovie(c *gin.Context) {
	var movie models.Movie

	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create movie"})
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func GetMovies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	title := c.Query("title")
	maxDuration := c.Query("max_duration")

	query := database.DB.Model(&models.Movie{})

	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	if maxDuration != "" {
		duration, err := strconv.Atoi(maxDuration)
		if err == nil {
			query = query.Where("duration_min <= ?", duration)
		}
	}

	var movies []models.Movie
	if err := query.Limit(limit).Offset(offset).Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  movies,
	})
}

func GetMovieByID(c *gin.Context) {
	id := c.Param("id")

	var movie models.Movie

	if err := database.DB.First(&movie, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, movie)
}
