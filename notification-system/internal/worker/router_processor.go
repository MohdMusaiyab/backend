package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
)

// RouterProcessor is the middleman. It listens for events and fans them out to specific channel queues.
type RouterProcessor struct {
	repo        repository.NotificationRepository
	queueClient *asynq.Client
}

// NewRouterProcessor creates a new instance of the router
func NewRouterProcessor(repo repository.NotificationRepository, queueClient *asynq.Client) *RouterProcessor {
	return &RouterProcessor{
		repo:        repo,
		queueClient: queueClient,
	}
}

// ProcessEventNotificationRequested grabs the generic event and fans it out!
func (p *RouterProcessor) ProcessEventNotificationRequested(ctx context.Context, t *asynq.Task) error {
	var payload EventNotificationRequestedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[ROUTER] 🔀 Pulled Event for %s. Analyzing routing preferences...", payload.Recipient)

	// Fetch the parent notification
	notif, err := p.repo.GetByID(ctx, payload.NotificationID)
	if err != nil {
		return fmt.Errorf("failed to fetch parent notification: %w", err)
	}

	// 1. ROUTER IDEMPOTENCY CHECK
	// If the router crashes mid-fan-out, we don't want it to duplicate the deliveries when it retries.
	if len(notif.Deliveries) > 0 {
		log.Printf("[ROUTER IDEMPOTENCY ✅] Event %s already fanned out! Skipping.", payload.NotificationID)
		return nil
	}

	// 2. Define the routing logic (In reality, we'd check user preferences in the DB)
	// For this system, we will fan-out to BOTH Email and SMS to prove it works.
	channels := []string{"email", "sms"}
	var deliveries []model.NotificationDelivery

	for _, ch := range channels {
		deliveries = append(deliveries, model.NotificationDelivery{
			NotificationID: notif.ID,
			Channel:        ch,
			Status:         "pending",
		})
	}

	// 3. Save the specific deliveries to the database
	if err := p.repo.SaveDeliveries(ctx, deliveries); err != nil {
		return fmt.Errorf("failed to save deliveries: %w", err)
	}

	// 4. THE FAN-OUT
	// GORM automatically populates the UUIDs in our 'deliveries' slice after saving them.
	// Now we loop through them and push them into completely isolated queues!
	for _, delivery := range deliveries {
		var task *asynq.Task
		var err error
		queueName := "default"

		if delivery.Channel == "email" {
			task, err = NewSendEmailTask(delivery.ID.String(), payload.Recipient, payload.Message)
			queueName = "email" // Routing to the 'email' queue!
		} else if delivery.Channel == "sms" {
			task, err = NewSendSMSTask(delivery.ID.String(), payload.Recipient, payload.Message)
			queueName = "sms"   // Routing to the 'sms' queue!
		}

		if err != nil {
			return fmt.Errorf("failed to create task for %s: %w", delivery.Channel, err)
		}

		// Enqueue the task to its specific queue. We use MaxRetry(3).
		_, err = p.queueClient.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Queue(queueName))
		if err != nil {
			return fmt.Errorf("failed to enqueue task for %s: %w", delivery.Channel, err)
		}
		
		log.Printf("[ROUTER] ➡️  Routed task to '%s' queue (DeliveryID: %s)", queueName, delivery.ID)
	}

	// 5. Mark the parent event as fully routed
	if err := p.repo.UpdateStatus(ctx, payload.NotificationID, "routed"); err != nil {
		return fmt.Errorf("failed to update parent status: %w", err)
	}

	log.Printf("[ROUTER] ✅ Successfully fanned out Event %s to %d isolated queues!", payload.NotificationID, len(channels))
	return nil
}
