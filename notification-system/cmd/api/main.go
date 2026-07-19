package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mohdMusaiyab/notification-system/internal/handler"
	"github.com/mohdMusaiyab/notification-system/internal/middleware"
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
	
	// B. External Providers (Notice we have two now!)
	emailSender := provider.NewMockEmailSender()
	smsSender := provider.NewMockSMSSender()
	
	// C. Queue Client (Producer side)
	queueClient := asynq.NewClient(redisConnOpt)
	defer queueClient.Close()
	
	// D. Core Service (API Brain)
	notificationService := service.NewNotificationService(repo, queueClient)
	
	// E. HTTP Handler
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// F. Worker Processors (Notice we have three now!)
	routerProcessor := worker.NewRouterProcessor(repo, queueClient)
	emailProcessor := worker.NewChannelProcessor("Email", repo, emailSender)
	smsProcessor := worker.NewChannelProcessor("SMS", repo, smsSender)

	// =========================================================================
	// 5. Start the Background Worker (Consumer)
	// =========================================================================
	go func() {
		// We define a total of 55 concurrent workers for our Node.
		workerServer := asynq.NewServer(
			redisConnOpt,
			asynq.Config{
				Concurrency: 55, 
				// The Magic of Rate Limiting via Queue Priorities:
				// Email queue gets 40 "weight". SMS queue gets only 5 "weight" to protect Twilio.
				Queues: map[string]int{
					"critical": 10, // Router events must be processed instantly
					"email":    40, // 40 concurrent workers chewing through fast AWS SES tasks
					"sms":      5,  // ONLY 5 concurrent workers talking to slow/rate-limited Twilio!
				},
				ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
					log.Printf("[WORKER RETRY ⚠️] Task %s failed. Reason: %v", task.Type(), err)
				}),
			},
		)

		mux := asynq.NewServeMux()
		
		// Map the Specific Task Types to their Specialized Processors!
		mux.HandleFunc(worker.TypeEventNotificationRequested, routerProcessor.ProcessEventNotificationRequested)
		mux.HandleFunc(worker.TypeSendEmail, emailProcessor.ProcessTask)
		mux.HandleFunc(worker.TypeSendSMS, smsProcessor.ProcessTask)

		log.Println("⚙️  Background Worker Pools started successfully! (Email: 40 | SMS: 5 | Router: 10)")
		
		if err := workerServer.Run(mux); err != nil {
			log.Fatalf("Worker server failed: %v", err)
		}
	}()

	// =========================================================================
	// 6. Start the HTTP API Server (Producer) on the main thread
	// =========================================================================
	router := gin.Default()
	
	// Create our API Gateway Rate Limiter (5 requests per second, burst of 10)
	limiter := middleware.NewIPRateLimiter(5, 10)
	
	// Apply the middleware strictly to the notification endpoint
	router.POST("/notification", middleware.RateLimit(limiter), notificationHandler.HandleSendNotification)

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}
	
	log.Printf("🚀 HTTP Server flying on port %s...", appPort)
	if err := router.Run(":" + appPort); err != nil {
		log.Fatalf("HTTP Server crashed: %v", err)
	}
}
