package main

import (
	"fmt"
	"log"

	"booking-system/internal/handlers"
	"booking-system/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting Booking System API...")

	database.ConnectDB()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/theaters", handlers.CreateTheater)
	r.GET("/theaters", handlers.GetTheaters)
	r.GET("/theaters/:id", handlers.GetTheaterByID)

	r.POST("/halls", handlers.CreateHall)
	r.GET("/halls", handlers.GetHalls)
	r.GET("/halls/:id", handlers.GetHallByID)

	r.POST("/movies", handlers.CreateMovie)
	r.GET("/movies", handlers.GetMovies)
	r.GET("/movies/:id", handlers.GetMovieByID)

	r.POST("/showtimes", handlers.CreateShowtime)
	r.GET("/showtimes", handlers.GetShowtimes)
	r.GET("/showtimes/:id", handlers.GetShowtimeByID)
	r.GET("/showtimes/:id/seats", handlers.GetShowtimeSeats)

	r.POST("/book", handlers.BookSeat)

	fmt.Println("🚀 Server is running on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
