package main

import (
	"fmt"
	"math/rand"

	"booking-system/internal/models"
	"booking-system/pkg/database"
)

func main() {
	fmt.Println("Running Database Seeder...")

	// 1. Connect to our Postgres Database
	database.ConnectDB()

	// 2. Control Panel: Comment or Uncomment the functions you want to run!
	seedTheaters()
	// seedHalls()
	// seedMovies()
	
	fmt.Println("🎉 Seeding completely finished!")
}

// -----------------------------------------------------
// SEEDER FUNCTIONS
// -----------------------------------------------------

func seedTheaters() {
	fmt.Println("Seeding 25 Random Theaters...")
	names := []string{"PVR", "Cinepolis", "INOX", "AMC", "Regal"}
	locations := []string{"Downtown", "North Mall", "City Center", "Sector 18", "West End"}

	for i := 1; i <= 25; i++ {
		randomName := names[rand.Intn(len(names))]
		randomLoc := locations[rand.Intn(len(locations))]

		theater := models.Theater{
			Name:     fmt.Sprintf("%s Cinema %d", randomName, rand.Intn(9999)),
			Location: fmt.Sprintf("%s - Area %d", randomLoc, rand.Intn(9999)),
		}

		if err := database.DB.Create(&theater).Error; err != nil {
			fmt.Printf("❌ Failed to create theater: %v\n", err)
		} else {
			fmt.Printf("✅ Created: %s at %s\n", theater.Name, theater.Location)
		}
	}
}
