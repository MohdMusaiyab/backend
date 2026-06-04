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
	// seedTheaters()  <-- I commented this out so you don't create another 25 theaters by accident!
	seedHalls()
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

func seedHalls() {
	fmt.Println("Seeding 5 Halls for each existing Theater...")
	
	var theaters []models.Theater
	
	// 1. Fetch all theaters currently in the DB
	if err := database.DB.Find(&theaters).Error; err != nil {
		fmt.Printf("❌ Failed to fetch theaters: %v\n", err)
		return
	}

	if len(theaters) == 0 {
		fmt.Println("⚠️ No theaters found! Please run seedTheaters() first.")
		return
	}

	totalHallsCreated := 0

	// 2. Loop through every single theater
	for _, theater := range theaters {
		
		// 3. Create exactly 5 Halls for the current theater
		for i := 1; i <= 5; i++ {
			// Randomize seat capacity between 50 and 150 seats per hall
			seatCapacity := rand.Intn(101) + 50 

			hall := models.Hall{
				TheaterID:  theater.ID,
				Name:       fmt.Sprintf("Screen %d", i),
				TotalSeats: seatCapacity,
			}

			if err := database.DB.Create(&hall).Error; err != nil {
				fmt.Printf("❌ Failed to create hall for theater %d: %v\n", theater.ID, err)
			} else {
				totalHallsCreated++
			}
		}
	}

	fmt.Printf("✅ Successfully seeded %d Halls!\n", totalHallsCreated)
}
