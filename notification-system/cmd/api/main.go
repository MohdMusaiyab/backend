package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mohdMusaiyab/notification-system/internal/handler"
	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/service"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
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
	log.Println("✅ Successfully connected to Postgres!")

	// 3. Setup Redis Connection Details
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisConnOpt := asynq.RedisClientOpt{Addr: redisAddr}

	// =========================================================================
	// 4. Dependency Injection (Wiring everything up)
	// =========================================================================
	
	// A. Data Access
	repo := repository.NewNotificationRepository(db)
	
	// B. External Provider
	sender := provider.NewMockSender()
	
	// C. Queue Client (Producer side)
	queueClient := asynq.NewClient(redisConnOpt)
	defer queueClient.Close() // Ensure the connection closes when the app shuts down
	
	// D. Core Service (API Brain)
	notificationService := service.NewNotificationService(repo, queueClient)
	
	// E. HTTP Handler
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// F. Worker Processor (Consumer Brain)
	notificationProcessor := worker.NewNotificationProcessor(repo, sender)

	// =========================================================================
	// 5. Start the Background Worker (Consumer)
	// We run this in a Goroutine so it operates concurrently in the background!
	// =========================================================================
	go func() {
		// Initialize the Asynq Server with a concurrency of 10 workers
		workerServer := asynq.NewServer(
			redisConnOpt,
			asynq.Config{Concurrency: 10},
		)

		// Create a router for our tasks (just like Gin for HTTP)
		mux := asynq.NewServeMux()
		
		// Tell the router: If you see a "notification:send" job, send it to our processor!
		mux.HandleFunc(worker.TypeSendNotification, notificationProcessor.ProcessTaskSendNotification)

		log.Println("⚙️  Background Worker Pool started successfully!")
		
		// Run the server (This is a blocking call, which is why it's in a Goroutine)
		if err := workerServer.Run(mux); err != nil {
			log.Fatalf("Worker server failed: %v", err)
		}
	}()

	// =========================================================================
	// 6. Start the HTTP API Server (Producer) on the main thread
	// =========================================================================
	router := gin.Default()
	router.POST("/notification", notificationHandler.HandleSendNotification)

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}
	
	log.Printf("🚀 HTTP Server flying on port %s...", appPort)
	if err := router.Run(":" + appPort); err != nil {
		log.Fatalf("HTTP Server crashed: %v", err)
	}
}
