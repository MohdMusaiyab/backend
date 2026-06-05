package handlers

import (
	"booking-system/internal/models"
	"booking-system/pkg/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type BookSeatRequest struct {
	ShowtimeSeatID uint   `json:"showtime_seat_id" binding:"required"`
	UserEmail      string `json:"user_email" binding:"required,email"`
}

func BookSeat(c *gin.Context) {
	var req BookSeatRequest

	// Validate JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Need showtime_seat_id and valid user_email"})
		return
	}

	// 1. Begin Database Transaction
	tx := database.DB.Begin()

	var seat models.ShowtimeSeat

	// 2. THE MOST IMPORTANT LINE IN THE PROJECT: PESSIMISTIC LOCKING
	// clause.Locking{Strength: "UPDATE"} translates to 'SELECT FOR UPDATE'.
	// It physically locks this specific row in Postgres. If User B tries to run this
	// while User A's transaction is still running, Postgres forces User B to wait!
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, req.ShowtimeSeatID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Seat not found"})
		return
	}

	// 3. Check if someone else already took it!
	if seat.Status != "AVAILABLE" {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Sorry, this seat is already taken!"})
		return
	}

	// 4. Update the seat status to BOOKED
	seat.Status = "BOOKED"
	if err := tx.Save(&seat).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update seat status"})
		return
	}

	// 5. Create the Booking receipt
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

	// 6. Commit the transaction (Saves everything AND releases the lock!)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking successful!",
		"booking": booking,
	})
}
