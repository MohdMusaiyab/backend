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
		// If the JSON is completely broken, retrying won't magically fix it.
		// We use asynq.SkipRetry to instantly move it to the Dead Letter Queue.
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[CONSUMER] 📥 Pulled task from Redis for recipient: %s", payload.Recipient)

	// 2. Do the heavy lifting (Call Twilio / AWS / Mock)
	// This takes 500ms, but the HTTP API doesn't care anymore because this runs in the background!
	err := processor.sender.Send(ctx, payload.Recipient, payload.Message)
	if err != nil {
		// If Twilio is down, we return the error.
		// This acts as a NACK (Negative Acknowledgment). Asynq will put it back in the queue and retry later.
		return fmt.Errorf("external sender failed: %w", err)
	}

	// 3. (Optional) We could update the database record to "sent" here.

	// 4. Return nil acts as an ACK (Acknowledgment). 
	// It tells Redis: "Job completely finished, you can delete it from RAM."
	return nil
}
