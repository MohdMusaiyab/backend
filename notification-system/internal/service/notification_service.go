package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

var ErrDuplicateRequest = errors.New("duplicate request detected")
var ErrSystemOverloaded = errors.New("system overloaded (backpressure applied)") 

type NotificationService interface {
	// Updated signature for Stage 8: We now accept an optional sendAt timestamp and Stage 9 RequestID
	ProcessNotification(ctx context.Context, userID, templateName string, data map[string]interface{}, idempotencyKey string, sendAt *time.Time, requestID string) error
}

type notificationService struct {
	repo        repository.NotificationRepository
	queueClient *asynq.Client
	inspector   *asynq.Inspector 
}

func NewNotificationService(repo repository.NotificationRepository, queueClient *asynq.Client, inspector *asynq.Inspector) NotificationService {
	return &notificationService{
		repo:        repo,
		queueClient: queueClient,
		inspector:   inspector,
	}
}

func (s *notificationService) ProcessNotification(ctx context.Context, userID, templateName string, data map[string]interface{}, idempotencyKey string, sendAt *time.Time, requestID string) error {
	
	// =========================================================================
	// 1. BACKPRESSURE: Load Shedding Check
	// =========================================================================
	queues, err := s.inspector.Queues()
	if err == nil {
		totalPending := 0
		for _, q := range queues {
			info, _ := s.inspector.GetQueueInfo(q)
			if info != nil {
				totalPending += info.Pending + info.Active
			}
		}
		
		if totalPending > 5000 {
			slog.Warn("System overloaded! Rejecting traffic.", "total_pending", totalPending)
			return ErrSystemOverloaded
		}
	}

	// =========================================================================
	// 2. TEMPLATE RESOLUTION (The Mid-Flight Crash Fix)
	// =========================================================================
	// Before doing anything, we ask the DB: "What is the absolute newest version of this template?"
	latestTemplate, err := s.repo.GetLatestTemplateVersion(ctx, templateName)
	if err != nil {
		return fmt.Errorf("failed to resolve template '%s' (it might not exist): %w", templateName, err)
	}

	// =========================================================================
	// 3. DATABASE RECORD (Idempotency)
	// =========================================================================
	// We serialize the data payload purely so we have a historical record in Postgres
	dataBytes, _ := json.Marshal(data)
	
	notif := &model.Notification{
		Recipient:      userID, // We store the UserID here instead of an email address
		Message:        fmt.Sprintf("Template: %s, Version: %s, Data: %s", templateName, latestTemplate.Version, string(dataBytes)),
		Status:         "pending",
		IdempotencyKey: idempotencyKey,
		SendAt:         sendAt, // STAGE 8: Save the target execution time for historical records
	}

	if err := s.repo.Save(ctx, notif); err != nil {
		if errors.Is(err, repository.ErrDuplicateIdempotencyKey) {
			return ErrDuplicateRequest
		}
		return fmt.Errorf("could not save pending notification: %w", err)
	}

	// =========================================================================
	// 4. THE QUEUE PAYLOAD & DELAYED SCHEDULING (Stage 8)
	// =========================================================================
	// Notice we physically stamp `latestTemplate.Version` onto the job payload.
	task, err := worker.NewEventNotificationRequestedTask(
		notif.ID.String(), 
		userID, 
		templateName, 
		latestTemplate.Version, 
		data,
		requestID, // Pass the trace ID to the queue!
	)
	if err != nil {
		return fmt.Errorf("could not create task: %w", err)
	}

	// Default options: retry up to 3 times, throw it in the critical queue
	opts := []asynq.Option{
		asynq.MaxRetry(3), 
		asynq.Queue("critical"),
	}

	// STAGE 9 MAGIC: Bind the trace ID to this logger!
	logger := slog.With("request_id", requestID)

	if sendAt != nil {
		opts = append(opts, asynq.ProcessAt(*sendAt))
		logger.Info("Time-Travel engaged! Scheduling task", "send_at", sendAt.Format(time.RFC1123))
	}

	info, err := s.queueClient.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("could not enqueue task: %w", err)
	}

	logger.Info("Enqueued task", "id", info.ID, "template", templateName, "version", latestTemplate.Version)
	return nil
}
