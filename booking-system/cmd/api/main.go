package main

import (
	"booking-system/pkg/database"
	"fmt"
)

func main() {
	fmt.Println("Booking System Backend is starting...")

	database.ConnectDB()
}
