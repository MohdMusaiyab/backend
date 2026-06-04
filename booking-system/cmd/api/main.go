package main

import (
	"fmt"
	"log"

	"booking-system/pkg/database"
	"booking-system/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Booking System Backend is starting...")

	database.ConnectDB()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "server is healthy and running",
		})
	})

	// --- Theater Endpoints ---
	r.POST("/theaters", handlers.CreateTheater)
	r.GET("/theaters", handlers.GetTheaters)
	r.GET("/theaters/:id", handlers.GetTheaterByID)

	// --- Hall Endpoints ---
	r.POST("/halls", handlers.CreateHall)
	r.GET("/halls", handlers.GetHalls)
	r.GET("/halls/:id", handlers.GetHallByID)

	// --- Movie Endpoints ---
	r.POST("/movies", handlers.CreateMovie)
	r.GET("/movies", handlers.GetMovies)
	r.GET("/movies/:id", handlers.GetMovieByID)

	// 4. Start the server
	fmt.Println("🚀 Server is running on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
