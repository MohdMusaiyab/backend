package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

// ErrDuplicateRequest lets the handler know the client sent the exact same request twice
var ErrDuplicateRequest = errors.New("duplicate request detected")

// NotificationService defines the core business logic interface
type NotificationService interface {
	ProcessNotification(ctx context.Context, recipient, message, idempotencyKey string) error
}

type notificationService struct {
	repo        repository.NotificationRepository
	queueClient *asynq.Client
}

// NewNotificationService creates a new instance of the service
func NewNotificationService(repo repository.NotificationRepository, queueClient *asynq.Client) NotificationService {
	return &notificationService{
		repo:        repo,
		queueClient: queueClient,
	}
}

// ProcessNotification handles the business rules before pushing to the queue
func (s *notificationService) ProcessNotification(ctx context.Context, recipient, message, idempotencyKey string) error {
	// 1. Create the database record
	notif := &model.Notification{
		Recipient:      recipient,
		Message:        message,
		Status:         "pending",
		IdempotencyKey: idempotencyKey, // Pass the key to the DB!
	}

	if err := s.repo.Save(ctx, notif); err != nil {
		// If Postgres rejected this exact key, we stop immediately! We DO NOT push to the queue.
		if errors.Is(err, repository.ErrDuplicateIdempotencyKey) {
			return ErrDuplicateRequest
		}
		return fmt.Errorf("could not save pending notification: %w", err)
	}

	// 2. Package the JSON payload. We explicitly pass the database ID down to the worker for idempotency checks!
	task, err := worker.NewSendNotificationTask(notif.ID.String(), recipient, message)
	if err != nil {
		return fmt.Errorf("could not create task: %w", err)
	}

	// 3. Push the task to the Redis Queue!
	// We explicitly set MaxRetry to 3. If it fails 3 times, it gets moved to the DLQ (Archived).
	// We also assign it to the "critical" queue so the worker prioritizes it over normal jobs.
	info, err := s.queueClient.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Queue("critical"))
	if err != nil {
		return fmt.Errorf("could not enqueue task: %w", err)
	}

	// Log it for our own visibility
	fmt.Printf("[PRODUCER] Enqueued task: id=%s type=%s queue=%s\n", info.ID, info.Type, info.Queue)
	
	// 4. Return instantly! No waiting 500ms for Twilio/Mock to respond.
	return nil
}
