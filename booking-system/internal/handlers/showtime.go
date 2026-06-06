package handlers

import (
	"booking-system/internal/models"
	"booking-system/pkg/database"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateShowtime(c *gin.Context) {
	var showtime models.Showtime

	if err := c.ShouldBindJSON(&showtime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Create(&showtime).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create showtime"})
		return
	}

	var hall models.Hall
	if err := tx.First(&hall, showtime.HallID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Hall ID"})
		return
	}

	var seats []models.ShowtimeSeat
	for i := 1; i <= hall.TotalSeats; i++ {
		seats = append(seats, models.ShowtimeSeat{
			ShowtimeID: showtime.ID,
			SeatNumber: fmt.Sprintf("S-%d", i),
			Status:     "AVAILABLE",
		})
	}

	if err := tx.Create(&seats).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate seats for showtime"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Showtime and seats generated successfully",
		"showtime":        showtime,
		"seats_generated": hall.TotalSeats,
	})
}

func GetShowtimes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var showtimes []models.Showtime

	if err := database.DB.Limit(limit).Offset(offset).Find(&showtimes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch showtimes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  showtimes,
	})
}

func GetShowtimeByID(c *gin.Context) {
	id := c.Param("id")
	var showtime models.Showtime

	if err := database.DB.First(&showtime, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
		return
	}

	c.JSON(http.StatusOK, showtime)
}

func GetShowtimeSeats(c *gin.Context) {
	id := c.Param("id")
	var seats []models.ShowtimeSeat

	if err := database.DB.Where("showtime_id = ?", id).Find(&seats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"showtime_id": id,
		"total_seats": len(seats),
		"seats":       seats,
	})
}
