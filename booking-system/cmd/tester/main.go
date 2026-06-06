package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"booking-system/internal/models"
	"booking-system/pkg/database"
)

func main() {
	fmt.Println("🚀 Starting Extreme Concurrency Load Test...")

	database.ConnectDB()

	var seat models.ShowtimeSeat
	if err := database.DB.Where("status = ?", "AVAILABLE").Order("RANDOM()").First(&seat).Error; err != nil {
		fmt.Println("❌ Failed to find any available seats. Did you run the seeder?")
		return
	}

	targetSeatID := seat.ID
	fmt.Printf("🎯 Dynamically selected Seat ID %d (Seat %s for Showtime %d)\n", targetSeatID, seat.SeatNumber, seat.ShowtimeID)
	
	time.Sleep(2 * time.Second)

	numberOfUsers := 100

	var wg sync.WaitGroup
	wg.Add(numberOfUsers)

	successCount := 0
	conflictCount := 0
	errorCount := 0

	var mu sync.Mutex

	fmt.Printf("🔥 Firing 100 concurrent requests at Seat %d...\n\n", targetSeatID)

	for i := 1; i <= numberOfUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			email := fmt.Sprintf("user%d@test.com", userID)
			jsonPayload := []byte(fmt.Sprintf(`{"showtime_seat_id": %d, "user_email": "%s"}`, targetSeatID, email))

			resp, err := http.Post("http://localhost:8080/book", "application/json", bytes.NewBuffer(jsonPayload))
			
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errorCount++
				fmt.Printf("⚠️ Request Failed: %v\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				successCount++
				fmt.Printf("🏆 %s WON THE RACE AND GOT THE SEAT!\n", email)
			} else if resp.StatusCode == 409 {
				conflictCount++
			} else {
				errorCount++
				
				// Read the server's response body to see the exact error
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				fmt.Printf("⚠️ Unexpected Status %d: %s\n", resp.StatusCode, buf.String())
			}
		}(i)
	}

	wg.Wait()

	fmt.Println("\n📊 TEST RESULTS:")
	fmt.Printf("Targeted Seat ID: %d\n", targetSeatID)
	fmt.Printf("Total Requests Sent: %d\n", numberOfUsers)
	fmt.Printf("Successful Bookings: %d\n", successCount)
	fmt.Printf("Rejected (Seat Taken): %d\n", conflictCount)
	fmt.Printf("Other Errors: %d\n", errorCount)
	
	if successCount == 1 && conflictCount == numberOfUsers-1 {
		fmt.Println("\n✅ PASS: System perfectly handled 100 concurrent requests with ZERO double-bookings!")
	} else {
		fmt.Println("\n❌ FAIL: Double booking detected or server crashed!")
	}
}
