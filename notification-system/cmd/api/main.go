package main

import (
	"context"
	"fmt"
	"log/slog"
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
	"github.com/mohdMusaiyab/notification-system/internal/telemetry"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
	
	"github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// STAGE 9.5: Initialize the WebSocket Hub for Live Log Streaming!
	wsHub := telemetry.NewHub()
	go wsHub.Run()

	// STAGE 9: Configure Global Structured JSON Logging
	// Instead of writing strictly to os.Stdout, we write to our custom HubWriter!
	hubWriter := telemetry.NewHubWriter(wsHub)
	logger := slog.New(slog.NewJSONHandler(hubWriter, nil))
	slog.SetDefault(logger)

	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found or error reading it.")
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
		slog.Error("Critical Error: Failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to Postgres!")

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
	
	// C.1 Queue Inspector (For Backpressure Load Shedding)
	inspector := asynq.NewInspector(redisConnOpt)
	defer inspector.Close()
	
	// D. Core Service (API Brain)
	notificationService := service.NewNotificationService(repo, queueClient, inspector)
	
	// E. HTTP Handler
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// F. Raw Redis Client (For the Global Rate Limiter inside the workers)
	rawRedisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rawRedisClient.Close()

	// G. Worker Processors
	routerProcessor := worker.NewRouterProcessor(repo, queueClient)
	emailProcessor := worker.NewChannelProcessor("Email", repo, emailSender, rawRedisClient)
	smsProcessor := worker.NewChannelProcessor("SMS", repo, smsSender, rawRedisClient)

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
					slog.Error("Worker Retry. Task failed.", "task_type", task.Type(), "error", err)
				}),
			},
		)

		mux := asynq.NewServeMux()
		
		// Map the Specific Task Types to their Specialized Processors!
		mux.HandleFunc(worker.TypeEventNotificationRequested, routerProcessor.ProcessEventNotificationRequested)
		mux.HandleFunc(worker.TypeSendEmail, emailProcessor.ProcessTask)
		mux.HandleFunc(worker.TypeSendSMS, smsProcessor.ProcessTask)

		slog.Info("Background Worker Pools started successfully!", "email_workers", 40, "sms_workers", 5, "router_workers", 10)
		
		if err := workerServer.Run(mux); err != nil {
			slog.Error("Worker server failed", "error", err)
			os.Exit(1)
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

	// STAGE 9: Expose the Prometheus metrics endpoint!
	// Prometheus will automatically ping this URL every 15 seconds to scrape our system's health.
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// STAGE 9.5: Expose the WebSocket endpoint for the Frontend Dashboard!
	router.GET("/ws", func(c *gin.Context) {
		telemetry.ServeWs(wsHub, c.Writer, c.Request)
	})

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}
	
	slog.Info("HTTP Server flying...", "port", appPort)
	if err := router.Run(":" + appPort); err != nil {
		slog.Error("HTTP Server crashed", "error", err)
		os.Exit(1)
	}
}
