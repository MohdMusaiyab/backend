package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	// The generic fan-out event (Pushed by the HTTP API)
	TypeEventNotificationRequested = "event:notification_requested"

	// The specific isolated channel tasks (Pushed by the Router Worker)
	TypeSendEmail = "notification:send:email"
	TypeSendSMS   = "notification:send:sms"
)

// EventNotificationRequestedPayload is used by the API to just say "Hey, an event happened!"
type EventNotificationRequestedPayload struct {
	NotificationID string
	Recipient      string
	Message        string
}

// ChannelDeliveryPayload is used by the specific Email/SMS workers
type ChannelDeliveryPayload struct {
	DeliveryID string // Notice we pass the DeliveryID, not the parent NotificationID!
	Recipient  string
	Message    string
}

// NewEventNotificationRequestedTask is created by the HTTP API Producer
func NewEventNotificationRequestedTask(notificationID, recipient, message string) (*asynq.Task, error) {
	payload, err := json.Marshal(EventNotificationRequestedPayload{
		NotificationID: notificationID,
		Recipient:      recipient,
		Message:        message,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventNotificationRequested, payload), nil
}

// NewSendEmailTask is created dynamically by the Router Worker
func NewSendEmailTask(deliveryID, recipient, message string) (*asynq.Task, error) {
	payload, err := json.Marshal(ChannelDeliveryPayload{
		DeliveryID: deliveryID,
		Recipient:  recipient,
		Message:    message,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, payload), nil
}

// NewSendSMSTask is created dynamically by the Router Worker
func NewSendSMSTask(deliveryID, recipient, message string) (*asynq.Task, error) {
	payload, err := json.Marshal(ChannelDeliveryPayload{
		DeliveryID: deliveryID,
		Recipient:  recipient,
		Message:    message,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendSMS, payload), nil
}
