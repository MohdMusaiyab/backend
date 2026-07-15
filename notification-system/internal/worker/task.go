package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeSendNotification = "notification:send"
)

// SendNotificationPayload defines the data stored in the Redis queue
type SendNotificationPayload struct {
	// We added the Database ID so the worker can query the database for idempotency!
	NotificationID string 
	Recipient      string
	Message        string
}

// NewSendNotificationTask packages the data for the background worker
func NewSendNotificationTask(notificationID, recipient, message string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendNotificationPayload{
		NotificationID: notificationID,
		Recipient:      recipient,
		Message:        message,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeSendNotification, payload), nil
}
