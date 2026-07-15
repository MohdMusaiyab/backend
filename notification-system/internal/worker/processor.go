package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
)

// NotificationProcessor is the actual "Worker" that lives in the background.
type NotificationProcessor struct {
	repo   repository.NotificationRepository
	sender provider.NotificationSender
}

// NewNotificationProcessor injects the dependencies this worker needs to do its job.
func NewNotificationProcessor(repo repository.NotificationRepository, sender provider.NotificationSender) *NotificationProcessor {
	return &NotificationProcessor{
		repo:   repo,
		sender: sender,
	}
}

// ProcessTaskSendNotification is the function that Asynq executes every time it pulls a job from Redis.
func (processor *NotificationProcessor) ProcessTaskSendNotification(ctx context.Context, t *asynq.Task) error {
	var payload SendNotificationPayload

	// 1. Deserialize the raw JSON bytes back into our Go struct
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[CONSUMER] 📥 Pulled task from Redis for recipient: %s", payload.Recipient)

	// =========================================================================
	// 2. IDEMPOTENCY CHECK (Type B Deduplication)
	// =========================================================================
	// We check the database BEFORE attempting to send the email.
	notif, err := processor.repo.GetByID(ctx, payload.NotificationID)
	if err != nil {
		// If we can't hit the DB, we safely NACK the job so it retries later.
		return fmt.Errorf("failed to fetch notification from DB: %w", err)
	}

	// If the system crashed after sending the email but before ACKing the job last time,
	// the broker will redeliver it. This check catches that exact redelivery!
	if notif.Status == "sent" {
		log.Printf("[CONSUMER IDEMPOTENCY ✅] Job %s was already sent! Skipping external call to prevent duplicate email.", payload.NotificationID)
		return nil // Returning nil acts as an ACK, instantly deleting the duplicate job from Redis.
	}

	// 3. Do the heavy lifting (Call Twilio / AWS / Mock)
	err = processor.sender.Send(ctx, payload.Recipient, payload.Message)
	if err != nil {
		// If the external provider fails, we leave the DB status as "pending" so the next retry can attempt it.
		return fmt.Errorf("external sender failed: %w", err)
	}

	// 4. Mark as sent in the database!
	// This is the physical lock that prevents future retries from double-sending.
	if err := processor.repo.UpdateStatus(ctx, payload.NotificationID, "sent"); err != nil {
		return fmt.Errorf("failed to update status to sent: %w", err)
	}

	log.Printf("[CONSUMER] ✅ Successfully sent email and updated DB for %s", payload.Recipient)
	return nil
}
