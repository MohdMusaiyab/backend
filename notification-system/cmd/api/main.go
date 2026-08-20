package main

import (
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
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/service"
	"github.com/mohdMusaiyab/notification-system/internal/telemetry"
	
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// STAGE 10: The Gateway (API Only)
// This binary is now strictly responsible for handling HTTP traffic and pushing events to Redis.
// It DOES NOT process any background jobs. 
func main() {
	wsHub := telemetry.NewHub()
	go wsHub.Run()

	hubWriter := telemetry.NewHubWriter(wsHub)
	logger := slog.New(slog.NewJSONHandler(hubWriter, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found or error reading it.")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
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
	slog.Info("API Gateway successfully connected to Postgres!")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisConnOpt := asynq.RedisClientOpt{Addr: redisAddr}

	repo := repository.NewNotificationRepository(db)
	
	// Queue Client (Producer side)
	queueClient := asynq.NewClient(redisConnOpt)
	defer queueClient.Close()
	
	// Queue Inspector (For Backpressure Load Shedding)
	inspector := asynq.NewInspector(redisConnOpt)
	defer inspector.Close()
	
	notificationService := service.NewNotificationService(repo, queueClient, inspector)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// API Gateway Server
	router := gin.Default()
	
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID, Idempotency-Key")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	
	limiter := middleware.NewIPRateLimiter(5, 10)
	router.POST("/notification", middleware.RateLimit(limiter), notificationHandler.HandleSendNotification)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/ws", func(c *gin.Context) {
		telemetry.ServeWs(wsHub, c.Writer, c.Request)
	})

	appPort := getEnv("APP_PORT", "8080")
	
	slog.Info("HTTP API Gateway flying...", "port", appPort)
	if err := router.Run(":" + appPort); err != nil {
		slog.Error("HTTP Server crashed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
