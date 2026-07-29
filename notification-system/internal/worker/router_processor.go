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

// RouterProcessor is the intelligent middleman. It listens for events, checks user preferences, and fans them out.
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

// ProcessEventNotificationRequested grabs the generic event and fans it out intelligently!
func (p *RouterProcessor) ProcessEventNotificationRequested(ctx context.Context, t *asynq.Task) error {
	var payload EventNotificationRequestedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[ROUTER] 🔀 Pulled Event for User: %s. Analyzing routing preferences...", payload.UserID)

	// Fetch the parent notification
	notif, err := p.repo.GetByID(ctx, payload.NotificationID)
	if err != nil {
		return fmt.Errorf("failed to fetch parent notification: %w", err)
	}

	// 1. ROUTER IDEMPOTENCY CHECK
	if len(notif.Deliveries) > 0 {
		log.Printf("[ROUTER IDEMPOTENCY ✅] Event %s already fanned out! Skipping.", payload.NotificationID)
		return nil
	}

	// =========================================================================
	// 2. FETCH USER PREFERENCES (The Stage 7 Magic!)
	// =========================================================================
	user, err := p.repo.GetUserByID(ctx, payload.UserID)
	if err != nil {
		// If the user doesn't exist, we skip retrying. It's a fatal error.
		return fmt.Errorf("failed to fetch user %s: %w", payload.UserID, asynq.SkipRetry)
	}

	// 3. INTELLIGENT ROUTING LOGIC
	var activeChannels []string
	
	// We read the JSONB map directly as a strongly-typed Go struct!
	if user.Preferences.Channels.Email {
		activeChannels = append(activeChannels, "email")
	}
	if user.Preferences.Channels.SMS {
		activeChannels = append(activeChannels, "sms")
	}

	// If the user turned EVERYTHING off, we just mark it as suppressed and gracefully stop.
	if len(activeChannels) == 0 {
		log.Printf("[ROUTER 🛑] User %s opted out of all channels. Suppressing notification.", payload.UserID)
		p.repo.UpdateStatus(ctx, payload.NotificationID, "suppressed_by_preference")
		return nil
	}

	var deliveries []model.NotificationDelivery
	for _, ch := range activeChannels {
		deliveries = append(deliveries, model.NotificationDelivery{
			NotificationID: notif.ID,
			Channel:        ch,
			Status:         "pending",
		})
	}

	// 4. Save the specific deliveries to the database
	if err := p.repo.SaveDeliveries(ctx, deliveries); err != nil {
		return fmt.Errorf("failed to save deliveries: %w", err)
	}

	// 5. THE FAN-OUT
	for _, delivery := range deliveries {
		var task *asynq.Task
		var err error
		queueName := "default"

		// We pass the incredibly rich payload straight down to the individual channel workers!
		if delivery.Channel == "email" {
			task, err = NewSendEmailTask(delivery.ID.String(), payload.UserID, payload.TemplateName, payload.TemplateVersion, payload.Data)
			queueName = "email"
		} else if delivery.Channel == "sms" {
			task, err = NewSendSMSTask(delivery.ID.String(), payload.UserID, payload.TemplateName, payload.TemplateVersion, payload.Data)
			queueName = "sms"
		}

		if err != nil {
			return fmt.Errorf("failed to create task for %s: %w", delivery.Channel, err)
		}

		_, err = p.queueClient.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Queue(queueName))
		if err != nil {
			return fmt.Errorf("failed to enqueue task for %s: %w", delivery.Channel, err)
		}
		
		log.Printf("[ROUTER] ➡️  Routed task to '%s' queue (DeliveryID: %s)", queueName, delivery.ID)
	}

	// 6. Mark the parent event as fully routed
	if err := p.repo.UpdateStatus(ctx, payload.NotificationID, "routed"); err != nil {
		return fmt.Errorf("failed to update parent status: %w", err)
	}

	log.Printf("[ROUTER] ✅ Successfully fanned out Event %s to %d isolated queues!", payload.NotificationID, len(activeChannels))
	return nil
}
