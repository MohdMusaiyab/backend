package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

var ErrDuplicateRequest = errors.New("duplicate request detected")
var ErrSystemOverloaded = errors.New("system overloaded (backpressure applied)") // 🔥 NEW!

type NotificationService interface {
	ProcessNotification(ctx context.Context, recipient, message, idempotencyKey string) error
}

type notificationService struct {
	repo        repository.NotificationRepository
	queueClient *asynq.Client
	inspector   *asynq.Inspector // 🔥 NEW! Allows us to peek into Redis Queues
}

// NewNotificationService creates a new instance of the service
func NewNotificationService(repo repository.NotificationRepository, queueClient *asynq.Client, inspector *asynq.Inspector) NotificationService {
	return &notificationService{
		repo:        repo,
		queueClient: queueClient,
		inspector:   inspector,
	}
}

// ProcessNotification handles the business rules before pushing to the queue
func (s *notificationService) ProcessNotification(ctx context.Context, recipient, message, idempotencyKey string) error {
	
	// =========================================================================
	// BACKPRESSURE: Load Shedding Check
	// =========================================================================
	// Before we even touch the database, check if the system is drowning in work.
	queues, err := s.inspector.Queues()
	if err == nil {
		totalPending := 0
		for _, q := range queues {
			info, _ := s.inspector.GetQueueInfo(q)
			if info != nil {
				// We care about tasks sitting waiting (Pending) and tasks currently executing (Active)
				totalPending += info.Pending + info.Active
			}
		}
		
		// If there are more than 5,000 tasks in the system, we aggressively reject new traffic.
		// (In a real system, you'd put this threshold in an env var instead of hardcoding)
		if totalPending > 5000 {
			log.Printf("[BACKPRESSURE ⚠️] System overloaded! Total tasks: %d. Rejecting traffic.", totalPending)
			return ErrSystemOverloaded
		}
	}

	// 1. Create the database record
	notif := &model.Notification{
		Recipient:      recipient,
		Message:        message,
		Status:         "pending",
		IdempotencyKey: idempotencyKey,
	}

	if err := s.repo.Save(ctx, notif); err != nil {
		if errors.Is(err, repository.ErrDuplicateIdempotencyKey) {
			return ErrDuplicateRequest
		}
		return fmt.Errorf("could not save pending notification: %w", err)
	}

	// 2. Package the generic event payload
	task, err := worker.NewEventNotificationRequestedTask(notif.ID.String(), recipient, message)
	if err != nil {
		return fmt.Errorf("could not create task: %w", err)
	}

	// 3. Push the task to the Redis Queue
	info, err := s.queueClient.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Queue("critical"))
	if err != nil {
		return fmt.Errorf("could not enqueue task: %w", err)
	}

	fmt.Printf("[PRODUCER] Enqueued task: id=%s type=%s queue=%s\n", info.ID, info.Type, info.Queue)
	return nil
}
