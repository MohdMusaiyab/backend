package main

import (
	"fmt"
	"math/rand"
	"time"

	"booking-system/internal/models"
	"booking-system/pkg/database"
)

func main() {
	fmt.Println("Running Database Seeder...")

	// 1. Connect to our Postgres Database
	database.ConnectDB()

	// 2. Control Panel: Comment or Uncomment the functions you want to run!
	// seedTheaters()  
	// seedHalls()
	// seedMovies() 
	seedShowtimes() // <-- Uncommented to run the Showtime & Seat generation!
	
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
		title := fmt.Sprintf("%s: %s (%d)", baseTitles[rand.Intn(len(baseTitles))], adjectives[rand.Intn(len(adjectives))], rand.Intn(999))
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

func seedShowtimes() {
	fmt.Println("Seeding Showtimes & Generating Seats...")
	
	var movies []models.Movie
	var halls []models.Hall
	
	// Fetch all existing movies and halls
	database.DB.Find(&movies)
	database.DB.Find(&halls)

	if len(movies) == 0 || len(halls) == 0 {
		fmt.Println("⚠️ Missing Movies or Halls in the DB. Run those seeders first!")
		return
	}

	showtimesCreated := 0
	seatsCreated := 0

	// We will create exactly 2 showtimes for EVERY hall in the database
	for _, hall := range halls {
		for i := 0; i < 2; i++ {
			// Pick a random movie
			randomMovie := movies[rand.Intn(len(movies))]
			
			// Give it a random start time (anywhere from right now to 7 days in the future)
			futureHours := rand.Intn(24 * 7)
			startTime := time.Now().Add(time.Duration(futureHours) * time.Hour)

			showtime := models.Showtime{
				MovieID:   randomMovie.ID,
				HallID:    hall.ID,
				StartTime: startTime,
			}

			// 1. Insert the Showtime
			if err := database.DB.Create(&showtime).Error; err != nil {
				fmt.Printf("❌ Failed to create showtime: %v\n", err)
				continue
			}
			showtimesCreated++

			// 2. Instantly generate all the individual physical seats for this specific showtime!
			var seats []models.ShowtimeSeat
			for s := 1; s <= hall.TotalSeats; s++ {
				seats = append(seats, models.ShowtimeSeat{
					ShowtimeID: showtime.ID,
					SeatNumber: fmt.Sprintf("S-%d", s),
					Status:     "AVAILABLE",
				})
			}

			// 3. Bulk insert the seats (this is extremely fast in Postgres)
			if err := database.DB.Create(&seats).Error; err != nil {
				fmt.Printf("❌ Failed to generate seats: %v\n", err)
			} else {
				seatsCreated += len(seats)
			}
		}
	}

	fmt.Printf("✅ Successfully created %d Showtimes and generated %d individual Seats!\n", showtimesCreated, seatsCreated)
}
