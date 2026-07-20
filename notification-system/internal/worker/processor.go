package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/redis/go-redis/v9"
)

// ChannelProcessor handles the specific channel tasks (Email, SMS)
type ChannelProcessor struct {
	repo        repository.NotificationRepository
	sender      provider.NotificationSender
	name        string
	redisClient *redis.Client // 🔥 NEW: Raw Redis client for Distributed Rate Limiting
}

// NewChannelProcessor creates a new instance of the channel-specific worker
func NewChannelProcessor(name string, repo repository.NotificationRepository, sender provider.NotificationSender, redisClient *redis.Client) *ChannelProcessor {
	return &ChannelProcessor{
		name:        name,
		repo:        repo,
		sender:      sender,
		redisClient: redisClient,
	}
}

// ProcessTask handles the specific delivery job from the isolated queue
func (processor *ChannelProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ChannelDeliveryPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[%s CONSUMER] 📥 Pulled task for recipient: %s", processor.name, payload.Recipient)

	// 1. IDEMPOTENCY CHECK
	delivery, err := processor.repo.GetDeliveryByID(ctx, payload.DeliveryID)
	if err != nil {
		return fmt.Errorf("failed to fetch delivery from DB: %w", err)
	}

	if delivery.Status == "sent" {
		log.Printf("[%s IDEMPOTENCY ✅] Delivery %s was already sent! Skipping.", processor.name, payload.DeliveryID)
		return nil
	}

	// =========================================================================
	// 2. GLOBAL WORKER RATE LIMITING (Distributed Fixed Window Counter)
	// =========================================================================
	if processor.name == "SMS" && processor.redisClient != nil {
		// We use a simple Redis Fixed Window Counter for rate limiting SMS globally
		// We use the current second as the key (e.g., "rate_limit:sms:1634567890")
		now := time.Now().Unix()
		key := fmt.Sprintf("rate_limit:sms:%d", now)

		// Increment the counter. Since Redis is single-threaded, this is 100% thread-safe globally!
		count, err := processor.redisClient.Incr(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis rate limiter failed: %w", err)
		}

		// If it's the first request in this exact second, set it to automatically delete itself 
		// after 5 seconds so our Redis server doesn't slowly run out of memory with millions of keys.
		if count == 1 {
			processor.redisClient.Expire(ctx, key, 5*time.Second)
		}

		// Twilio Limit: Maximum 2 SMS per second globally!
		// If we are request #3, we immediately abort.
		if count > 2 {
			log.Printf("[SMS GLOBAL RATE LIMIT ⚠️] Twilio capacity reached! Backing off...")
			// Returning an error tells Asynq to put the task back in the queue and try again later
			return fmt.Errorf("global SMS rate limit exceeded (2/sec)")
		}
	}

	// 3. Call the external provider (Twilio or AWS SES)
	err = processor.sender.Send(ctx, payload.Recipient, payload.Message)
	if err != nil {
		return fmt.Errorf("external sender failed: %w", err)
	}

	// 4. Mark this specific channel as sent!
	if err := processor.repo.UpdateDeliveryStatus(ctx, payload.DeliveryID, "sent"); err != nil {
		return fmt.Errorf("failed to update status to sent: %w", err)
	}

	log.Printf("[%s CONSUMER] ✅ Successfully sent to %s and updated DB!", processor.name, payload.Recipient)
	return nil
}
