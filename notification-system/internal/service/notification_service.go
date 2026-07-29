package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/worker"
)

var ErrDuplicateRequest = errors.New("duplicate request detected")
var ErrSystemOverloaded = errors.New("system overloaded (backpressure applied)") 

type NotificationService interface {
	// Updated signature for Stage 7: We now accept UserID, TemplateName, and a highly dynamic Data map
	ProcessNotification(ctx context.Context, userID, templateName string, data map[string]interface{}, idempotencyKey string) error
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

func (s *notificationService) ProcessNotification(ctx context.Context, userID, templateName string, data map[string]interface{}, idempotencyKey string) error {
	
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
			log.Printf("[BACKPRESSURE ⚠️] System overloaded! Total tasks: %d. Rejecting traffic.", totalPending)
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
	}

	if err := s.repo.Save(ctx, notif); err != nil {
		if errors.Is(err, repository.ErrDuplicateIdempotencyKey) {
			return ErrDuplicateRequest
		}
		return fmt.Errorf("could not save pending notification: %w", err)
	}

	// =========================================================================
	// 4. THE QUEUE PAYLOAD
	// =========================================================================
	// Notice we physically stamp `latestTemplate.Version` onto the job payload.
	// Even if marketing releases v2 while this job is stuck in the queue, 
	// the worker will guarantee it uses the correct v1 template!
	task, err := worker.NewEventNotificationRequestedTask(
		notif.ID.String(), 
		userID, 
		templateName, 
		latestTemplate.Version, 
		data,
	)
	if err != nil {
		return fmt.Errorf("could not create task: %w", err)
	}

	info, err := s.queueClient.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Queue("critical"))
	if err != nil {
		return fmt.Errorf("could not enqueue task: %w", err)
	}

	log.Printf("[PRODUCER] Enqueued task: id=%s template=%s version=%s", info.ID, templateName, latestTemplate.Version)
	return nil
}
