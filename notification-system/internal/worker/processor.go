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

// ChannelProcessor handles the specific channel tasks (Email, SMS)
type ChannelProcessor struct {
	repo   repository.NotificationRepository
	sender provider.NotificationSender
	name   string // 'Email' or 'SMS' for logging
}

// NewChannelProcessor creates a new instance of the channel-specific worker
func NewChannelProcessor(name string, repo repository.NotificationRepository, sender provider.NotificationSender) *ChannelProcessor {
	return &ChannelProcessor{
		name:   name,
		repo:   repo,
		sender: sender,
	}
}

// ProcessTask handles the specific delivery job from the isolated queue
func (processor *ChannelProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ChannelDeliveryPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[%s CONSUMER] 📥 Pulled task for recipient: %s", processor.name, payload.Recipient)

	// 1. IDEMPOTENCY CHECK (Type B Deduplication - Specific to this channel)
	delivery, err := processor.repo.GetDeliveryByID(ctx, payload.DeliveryID)
	if err != nil {
		return fmt.Errorf("failed to fetch delivery from DB: %w", err)
	}

	if delivery.Status == "sent" {
		log.Printf("[%s IDEMPOTENCY ✅] Delivery %s was already sent! Skipping.", processor.name, payload.DeliveryID)
		return nil
	}

	// 2. Call the external provider (Twilio or AWS SES)
	err = processor.sender.Send(ctx, payload.Recipient, payload.Message)
	if err != nil {
		return fmt.Errorf("external sender failed: %w", err)
	}

	// 3. Mark this specific channel as sent! (Email could be sent, while SMS is still failing)
	if err := processor.repo.UpdateDeliveryStatus(ctx, payload.DeliveryID, "sent"); err != nil {
		return fmt.Errorf("failed to update status to sent: %w", err)
	}

	log.Printf("[%s CONSUMER] ✅ Successfully sent to %s and updated DB!", processor.name, payload.Recipient)
	return nil
}
