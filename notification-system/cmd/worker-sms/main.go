package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/redis/go-redis/v9"

	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/telemetry"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

// STAGE 10: SMS Worker
func main() {
	if err := godotenv.Load(); err != nil {
		// Silent load error
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
	slog.Info("SMS Worker successfully connected to Postgres!")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisConnOpt := asynq.RedisClientOpt{Addr: redisAddr}
	
	rawRedisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rawRedisClient.Close()
	
	// STAGE 10.5: Centralized Logging Publisher
	redisWriter := telemetry.NewRedisPubSubWriter(rawRedisClient, "global_telemetry")
	multiWriter := io.MultiWriter(os.Stdout, redisWriter)
	logger := slog.New(slog.NewJSONHandler(multiWriter, nil))
	slog.SetDefault(logger)

	repo := repository.NewNotificationRepository(db)
	smsSender := provider.NewMockSMSSender()
	smsProcessor := worker.NewChannelProcessor("SMS", repo, smsSender, rawRedisClient)

	workerServer := asynq.NewServer(
		redisConnOpt,
		asynq.Config{
			Concurrency: 5, // Strict low concurrency to protect SMS providers (e.g. Twilio)
			Queues: map[string]int{
				"sms": 10, // ONLY listens to SMS queue
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				slog.Error("SMS Retry. Task failed.", "task_type", task.Type(), "error", err)
			}),
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeSendSMS, smsProcessor.ProcessTask)

	slog.Info("SMS Worker started successfully! Listening on 'sms' queue...")
	
	if err := workerServer.Run(mux); err != nil {
		slog.Error("SMS Worker server failed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
