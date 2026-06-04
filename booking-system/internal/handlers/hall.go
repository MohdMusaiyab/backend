package handlers

import (
	"booking-system/internal/models"
	"booking-system/pkg/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateHall(c *gin.Context) {
	var hall models.Hall

	if err := c.ShouldBindJSON(&hall); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := database.DB.Create(&hall).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hall. Ensure TheaterID exists."})
		return
	}

	c.JSON(http.StatusCreated, hall)
}

func GetHalls(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var halls []models.Hall

	if err := database.DB.Limit(limit).Offset(offset).Find(&halls).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch halls"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  halls,
	})
}

func GetHallByID(c *gin.Context) {
	id := c.Param("id")

	var hall models.Hall

	if err := database.DB.First(&hall, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hall not found"})
		return
	}

	c.JSON(http.StatusOK, hall)
}
