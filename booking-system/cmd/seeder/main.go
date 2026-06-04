package main

import (
	"fmt"
	"math/rand"

	"booking-system/internal/models"
	"booking-system/pkg/database"
)

func main() {
	fmt.Println("Running Database Seeder...")

	database.ConnectDB()

	// 2. Control Panel: Comment or Uncomment the functions you want to run!
	// seedTheaters()
	// seedHalls()
	seedMovies() // <-- Uncommented to run the movies seeder!
	
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
	
	if err := database.DB.Find(&theaters).Error; err != nil {
		fmt.Printf("❌ Failed to fetch theaters: %v\n", err)
		return
	}

	if len(theaters) == 0 {
		fmt.Println("⚠️ No theaters found! Please run seedTheaters() first.")
		return
	}

	totalHallsCreated := 0

	for _, theater := range theaters {
		for i := 1; i <= 5; i++ {
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

func seedMovies() {
	fmt.Println("Seeding 25 Random Movies...")

	baseTitles := []string{"Inception", "The Dark Knight", "Interstellar", "Dune", "Avatar", "Oppenheimer", "The Matrix", "Gladiator", "Titanic"}
	adjectives := []string{"Returns", "Part II", "The Awakening", "Origins", "Uncut", "Remastered"}

	moviesCreated := 0

	for i := 1; i <= 25; i++ {
		// Create a dynamic movie title like "Inception: Returns (421)"
		title := fmt.Sprintf("%s: %s (%d)", baseTitles[rand.Intn(len(baseTitles))], adjectives[rand.Intn(len(adjectives))], rand.Intn(999))
		
		// Random duration between 90 and 180 minutes
		duration := rand.Intn(91) + 90 

		movie := models.Movie{
			Title:       title,
			Description: fmt.Sprintf("An amazing cinematic experience about %s.", title),
			DurationMin: duration,
		}

		if err := database.DB.Create(&movie).Error; err != nil {
			fmt.Printf("❌ Failed to create movie: %v\n", err)
		} else {
			moviesCreated++
			fmt.Printf("✅ Created: %s (%d mins)\n", movie.Title, movie.DurationMin)
		}
	}

	fmt.Printf("✅ Successfully seeded %d Movies!\n", moviesCreated)
}
