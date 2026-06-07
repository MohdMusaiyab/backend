package database

import (
	"fmt"
	"log"
	"os"

	"booking-system/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {

		dsn = "host=localhost user=booking_user password=booking_secret dbname=booking_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get sql.DB from gorm: %v", err)
	}

	sqlDB.SetMaxOpenConns(80)
	sqlDB.SetMaxIdleConns(10)

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
