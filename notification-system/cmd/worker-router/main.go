package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

// STAGE 10: Router Worker
// This binary strictly listens to the "critical" queue, reads user preferences,
// and fans out the notification to the Email or SMS queues.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
	slog.Info("Router Worker successfully connected to Postgres!")

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisConnOpt := asynq.RedisClientOpt{Addr: redisAddr}

	repo := repository.NewNotificationRepository(db)
	
	queueClient := asynq.NewClient(redisConnOpt)
	defer queueClient.Close()
	
	routerProcessor := worker.NewRouterProcessor(repo, queueClient)

	workerServer := asynq.NewServer(
		redisConnOpt,
		asynq.Config{
			Concurrency: 10, 
			Queues: map[string]int{
				"critical": 10, // Router ONLY listens to critical events
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				slog.Error("Router Retry. Task failed.", "task_type", task.Type(), "error", err)
			}),
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeEventNotificationRequested, routerProcessor.ProcessEventNotificationRequested)

	slog.Info("Router Worker started successfully! Listening on 'critical' queue...")
	
	if err := workerServer.Run(mux); err != nil {
		slog.Error("Router Worker server failed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
