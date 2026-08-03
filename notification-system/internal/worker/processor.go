package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
	"github.com/mohdMusaiyab/notification-system/internal/telemetry"
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
func (processor *ChannelProcessor) ProcessTask(ctx context.Context, t *asynq.Task) (err error) {
	// STAGE 9: Telemetry Timers
	startTime := time.Now()
	
	// Defer block mathematically guarantees we record metrics even if a panic happens or we return early
	defer func() {
		status := "success"
		if err != nil {
			status = "failed"
		}
		
		duration := time.Since(startTime).Seconds()
		// 1. Record exact latency
		telemetry.NotificationLatency.WithLabelValues(processor.name).Observe(duration)
		// 2. Increment success/fail counter
		telemetry.NotificationsProcessed.WithLabelValues(processor.name, status).Inc()
	}()

	var payload ChannelDeliveryPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[%s CONSUMER] 📥 Pulled task for User: %s", processor.name, payload.UserID)

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
		now := time.Now().Unix()
		key := fmt.Sprintf("rate_limit:sms:%d", now)

		count, err := processor.redisClient.Incr(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis rate limiter failed: %w", err)
		}

		if count == 1 {
			processor.redisClient.Expire(ctx, key, 5*time.Second)
		}

		if count > 2 {
			log.Printf("[SMS GLOBAL RATE LIMIT ⚠️] Twilio capacity reached! Backing off...")
			return fmt.Errorf("global SMS rate limit exceeded (2/sec)")
		}
	}

	// =========================================================================
	// 3. FETCH USER DATA & TEMPLATE (The Stage 7 Magic!)
	// =========================================================================
	// We no longer pass raw emails through the queue. We fetch the user NOW.
	user, err := processor.repo.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s: %w", payload.UserID, asynq.SkipRetry)
	}

	// Dynamically determine the destination
	contactInfo := ""
	if processor.name == "Email" {
		contactInfo = user.Email
	} else if processor.name == "SMS" {
		contactInfo = user.Phone
	}

	if contactInfo == "" {
		return fmt.Errorf("user %s has no contact info for channel %s: %w", payload.UserID, processor.name, asynq.SkipRetry)
	}

	// Fetch the strictly versioned template that was frozen onto the payload
	// This completely prevents mid-flight crashes!
	dbTemplate, err := processor.repo.GetTemplate(ctx, payload.TemplateName, payload.TemplateVersion)
	if err != nil {
		return fmt.Errorf("failed to fetch template '%s' %s: %w", payload.TemplateName, payload.TemplateVersion, asynq.SkipRetry)
	}

	// =========================================================================
	// 4. DYNAMIC TEMPLATE RENDERING
	// =========================================================================
	// We use Go's built-in text/template engine to merge the raw DB string with the JSON payload
	tmpl, err := template.New("notification").Parse(dbTemplate.BodyTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err, asynq.SkipRetry)
	}

	var renderedBody bytes.Buffer
	if err := tmpl.Execute(&renderedBody, payload.Data); err != nil {
		// If the JSON payload doesn't match the template variables, it will error here.
		return fmt.Errorf("failed to execute template: %w", err, asynq.SkipRetry)
	}

	// 5. Call the external provider (Twilio or AWS SES)
	err = processor.sender.Send(ctx, contactInfo, renderedBody.String())
	if err != nil {
		return fmt.Errorf("external sender failed: %w", err)
	}

	// 6. Mark this specific channel as sent!
	if err := processor.repo.UpdateDeliveryStatus(ctx, payload.DeliveryID, "sent"); err != nil {
		return fmt.Errorf("failed to update status to sent: %w", err)
	}

	log.Printf("[%s CONSUMER] ✅ Successfully sent to %s and updated DB!", processor.name, contactInfo)
	return nil
}
