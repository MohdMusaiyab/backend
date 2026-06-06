package handlers

import (
	"booking-system/internal/models"
	"booking-system/pkg/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookSeatRequest struct {
	ShowtimeSeatID uint   `json:"showtime_seat_id" binding:"required"`
	UserEmail      string `json:"user_email" binding:"required,email"`
}

func BookSeat(c *gin.Context) {
	var req BookSeatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Need showtime_seat_id and valid user_email"})
		return
	}

	tx := database.DB.Begin()
	var seat models.ShowtimeSeat

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, req.ShowtimeSeatID).Error; err != nil {
		tx.Rollback()
		// We were blindly assuming ANY error meant "Seat Not Found". Let's fix that:
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Seat not found"})
		} else {
			// This will catch Connection Pool timeouts, database lock failures, etc!
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while acquiring lock", "details": err.Error()})
		}
		return
	}

	if seat.Status != "AVAILABLE" {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Sorry, this seat is already taken!"})
		return
	}

	time.Sleep(100 * time.Millisecond)

	seat.Status = "BOOKED"
	if err := tx.Save(&seat).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update seat status"})
		return
	}

	booking := models.Booking{
		ShowtimeSeatID: seat.ID,
		UserEmail:      req.UserEmail,
		Status:         "CONFIRMED",
		ExpiresAt:      time.Now().Add(time.Hour * 24),
	}

	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking record"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking successful!",
		"booking": booking,
	})
}
