package service

import (
	"context"

	"github.com/mohdMusaiyab/notification-system/internal/model"
	"github.com/mohdMusaiyab/notification-system/internal/provider"
	"github.com/mohdMusaiyab/notification-system/internal/repository"
)

// NotificationService defines the use-cases for our application
type NotificationService interface {
	ProcessNotification(ctx context.Context, recipient, message string) error
}

// notificationService struct holds our external dependencies
type notificationService struct {
	repo   repository.NotificationRepository
	sender provider.NotificationSender
}

// NewNotificationService is the constructor where we "Inject" the dependencies
func NewNotificationService(repo repository.NotificationRepository, sender provider.NotificationSender) NotificationService {
	return &notificationService{
		repo:   repo,
		sender: sender,
	}
}

// ProcessNotification is our Core Business Logic. It orchestrates the flow.
func (s *notificationService) ProcessNotification(ctx context.Context, recipient, message string) error {
	// Default status
	status := "sent"

	// 1. Ask the Provider to send the message
	err := s.sender.Send(ctx, recipient, message)
	if err != nil {
		// If Twilio/Mock failed, we update our status
		status = "failed"
	}

	// 2. Build our Database Model
	notif := &model.Notification{
		Recipient: recipient,
		Message:   message,
		Status:    status,
	}

	// 3. Ask the Repository to save it to Postgres
	saveErr := s.repo.Save(ctx, notif)
	if saveErr != nil {
		// If the database is down, we return a critical error
		return saveErr
	}

	// Return the original sender error (if it failed to send but saved successfully)
	// Or return nil if everything was perfect.
	return err
}
