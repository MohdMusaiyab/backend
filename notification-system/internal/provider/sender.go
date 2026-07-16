package provider

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// NotificationSender represents a generic external provider
type NotificationSender interface {
	Send(ctx context.Context, recipient, message string) error
}

// MockEmailSender simulates an external Email provider like AWS SES (Fast, Highly Reliable)
type MockEmailSender struct{}

func NewMockEmailSender() NotificationSender {
	return &MockEmailSender{}
}

func (s *MockEmailSender) Send(ctx context.Context, recipient, message string) error {
	// Emails are fast (100ms)
	time.Sleep(100 * time.Millisecond)

	// High reliability (5% failure rate)
	if rand.Intn(100) < 5 {
		return fmt.Errorf("aws ses temporary timeout")
	}

	return nil
}

// MockSMSSender simulates an external SMS provider like Twilio (Slow, Rate Limited, Unreliable)
type MockSMSSender struct{}

func NewMockSMSSender() NotificationSender {
	return &MockSMSSender{}
}

func (s *MockSMSSender) Send(ctx context.Context, recipient, message string) error {
	// SMS is inherently slower (500ms)
	time.Sleep(500 * time.Millisecond)

	// Low reliability (30% failure rate) representing strict rate limits
	if rand.Intn(100) < 30 {
		return fmt.Errorf("twilio rate limit exceeded (429 Too Many Requests)")
	}

	return nil
}
