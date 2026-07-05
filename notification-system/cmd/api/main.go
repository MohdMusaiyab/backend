package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mohdMusaiyab/notification-system/internal/handler"
	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/service"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error reading it.")
	}

	// 2. Setup PostgreSQL Connection
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("DB_PORT"),
	)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Critical Error: Failed to connect to database: %v", err)
	}
	log.Println("✅ Successfully connected to the Postgres Database!")

	// =========================================================================
	// 3. Dependency Injection (Wiring the Application)
	// We build our application layer by layer, from the inside out.
	// =========================================================================
	
	// A. Initialize the Data Access Layer (Repository)
	repo := repository.NewNotificationRepository(db)
	
	// B. Initialize the External Provider (Sender)
	sender := provider.NewMockSender()
	
	// C. Inject both into the Core Business Logic (Service)
	notificationService := service.NewNotificationService(repo, sender)
	
	// D. Inject the Core Service into the HTTP Transport Layer (Handler)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// =========================================================================
	// 4. HTTP Router Setup
	// =========================================================================
	router := gin.Default()

	// Register our single endpoint, hooking it up to our handler
	router.POST("/notification", notificationHandler.HandleSendNotification)

	// 5. Start the Server
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}
	
	log.Printf("🚀 Server is flying on port %s...", appPort)
	if err := router.Run(":" + appPort); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
