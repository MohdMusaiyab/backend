package provider

import (
	"context"
	"fmt"
	"time"
)

// NotificationSender defines the interface for external APIs (like AWS SES, Twilio, etc.)
type NotificationSender interface {
	Send(ctx context.Context, recipient, message string) error
}

// mockSender is a fake implementation we use for local development (Stage 0)
type mockSender struct{}

// NewMockSender creates and returns our fake sender
func NewMockSender() NotificationSender {
	return &mockSender{}
}

// Send simulates a network request by waiting 500ms, then printing to the console.
func (s *mockSender) Send(ctx context.Context, recipient, message string) error {
	// Simulating external network latency
	time.Sleep(500 * time.Millisecond)
	
	fmt.Printf("[MOCK SENDER] Successfully dispatched message to %s: %s\n", recipient, message)
	
	// In a real scenario, we might return an error if Twilio is down. 
	// For now, it always succeeds.
	return nil
}
