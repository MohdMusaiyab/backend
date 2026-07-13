package service

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

// NotificationService defines the use-cases for our application
type NotificationService interface {
	ProcessNotification(ctx context.Context, recipient, message string) error
}

// notificationService struct holds our dependencies.
// Notice we replaced the slow 'sender' with the blazing fast 'queueClient'.
type notificationService struct {
	repo        repository.NotificationRepository
	queueClient *asynq.Client
}

// NewNotificationService injects the repository and the Redis queue client
func NewNotificationService(repo repository.NotificationRepository, queueClient *asynq.Client) NotificationService {
	return &notificationService{
		repo:        repo,
		queueClient: queueClient,
	}
}

// ProcessNotification is now a "Producer". It saves state and offloads the heavy work.
func (s *notificationService) ProcessNotification(ctx context.Context, recipient, message string) error {
	// 1. Instantly save to DB with "pending" status.
	// The background worker will update this to "sent" or "failed" later.
	notif := &model.Notification{
		Recipient: recipient,
		Message:   message,
		Status:    "pending",
	}

	if err := s.repo.Save(ctx, notif); err != nil {
		return fmt.Errorf("could not save pending notification: %w", err)
	}

	// 2. Package the JSON payload using our helper from Step 2
	task, err := worker.NewSendNotificationTask(recipient, message)
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
