package database

import (
	"fmt"
	"log"

	"booking-system/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {

	dsn := "host=localhost user=booking_user password=booking_secret dbname=booking_db port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// -----------------------------------------------------------
	// CONNECTION POOLING FIX
	// -----------------------------------------------------------
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get sql.DB from gorm: %v", err)
	}

	// PostgreSQL's default hard limit is 100 simultaneous connections.
	// By setting MaxOpenConns to 80, we tell our Go app:
	// "If 100 people hit our server at the exact same millisecond, 
	// let 80 talk to the DB, and put the other 20 in a waiting queue."
	// This prevents the "FATAL: too many clients already" crash!
	sqlDB.SetMaxOpenConns(80)
	sqlDB.SetMaxIdleConns(10)
	// -----------------------------------------------------------

	fmt.Println("✅ Successfully connected to PostgreSQL!")

	err = db.AutoMigrate(
		&models.Theater{},
		&models.Hall{},
		&models.Movie{},
		&models.Showtime{},
		&models.ShowtimeSeat{},
		&models.Booking{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to auto-migrate database: %v", err)
	}

	fmt.Println("✅ Database Migration Completed successfully!")

	DB = db
}
