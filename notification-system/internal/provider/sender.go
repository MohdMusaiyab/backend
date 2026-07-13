package provider

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

// NotificationSender defines the interface for external notification services
type NotificationSender interface {
	Send(ctx context.Context, recipient, message string) error
}

// mockSender simulates a network delay AND chaotic external failures
type mockSender struct {
	rng *rand.Rand
}

// NewMockSender creates a new instance of mockSender with a random number generator
func NewMockSender() NotificationSender {
	// We seed the random number generator so the failures are truly random on every run
	source := rand.NewSource(time.Now().UnixNano())
	return &mockSender{
		rng: rand.New(source),
	}
}

// Send simulates an external API call (like Twilio or AWS SES)
func (m *mockSender) Send(ctx context.Context, recipient, message string) error {
	// 1. Always simulate the 500ms network latency first
	time.Sleep(500 * time.Millisecond)

	// 2. Introduce 40% "Chaotic" Failure
	// Float32 returns a decimal between 0.0 and 1.0
	if m.rng.Float32() < 0.40 {
		log.Printf("[MOCK SENDER ❌] Simulating catastrophic network failure for %s", recipient)
		// Returning this error tells the Asynq consumer (worker) that the job FAILED.
		// Asynq will capture this exact error string and NACK the job, triggering a retry.
		return errors.New("503 Service Unavailable: Twilio API Timeout")
	}

	// 3. The 60% Success Case
	log.Printf("[MOCK SENDER ✅] Successfully dispatched message to %s", recipient)
	return nil
}
