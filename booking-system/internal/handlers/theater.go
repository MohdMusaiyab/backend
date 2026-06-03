package handlers

import (
	"booking-system/internal/models"
	"booking-system/pkg/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateTheater(c *gin.Context) {
	var theater models.Theater

	if err := c.ShouldBindJSON(&theater); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := database.DB.Create(&theater).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create theater"})
		return
	}

	c.JSON(http.StatusCreated, theater)
}

func GetTheaters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var theaters []models.Theater

	if err := database.DB.Limit(limit).Offset(offset).Find(&theaters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch theaters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  theaters,
	})
}

func GetTheaterByID(c *gin.Context) {
	id := c.Param("id")

	var theater models.Theater

	if err := database.DB.Preload("Halls").First(&theater, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Theater not found"})
		return
	}

	c.JSON(http.StatusOK, theater)
}
