package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeEventNotificationRequested = "event:notification_requested"
	TypeSendEmail                  = "notification:send:email"
	TypeSendSMS                    = "notification:send:sms"
)

// EventNotificationRequestedPayload is used by the API to just say "Hey, an event happened!"
type EventNotificationRequestedPayload struct {
	NotificationID  string
	UserID          string
	TemplateName    string
	TemplateVersion string
	Data            map[string]interface{}
	RequestID       string // STAGE 9: The Tracing Baton!
}

// ChannelDeliveryPayload is used by the specific Email/SMS workers
type ChannelDeliveryPayload struct {
	DeliveryID      string 
	UserID          string
	TemplateName    string
	TemplateVersion string
	Data            map[string]interface{}
	RequestID       string // STAGE 9: The Tracing Baton!
}

// NewEventNotificationRequestedTask is created by the HTTP API Producer
func NewEventNotificationRequestedTask(notificationID, userID, templateName, templateVersion string, data map[string]interface{}, requestID string) (*asynq.Task, error) {
	payload, err := json.Marshal(EventNotificationRequestedPayload{
		NotificationID:  notificationID,
		UserID:          userID,
		TemplateName:    templateName,
		TemplateVersion: templateVersion,
		Data:            data,
		RequestID:       requestID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventNotificationRequested, payload), nil
}

// NewSendEmailTask is created dynamically by the Router Worker
func NewSendEmailTask(deliveryID, userID, templateName, templateVersion string, data map[string]interface{}, requestID string) (*asynq.Task, error) {
	payload, err := json.Marshal(ChannelDeliveryPayload{
		DeliveryID:      deliveryID,
		UserID:          userID,
		TemplateName:    templateName,
		TemplateVersion: templateVersion,
		Data:            data,
		RequestID:       requestID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, payload), nil
}

// NewSendSMSTask is created dynamically by the Router Worker
func NewSendSMSTask(deliveryID, userID, templateName, templateVersion string, data map[string]interface{}, requestID string) (*asynq.Task, error) {
	payload, err := json.Marshal(ChannelDeliveryPayload{
		DeliveryID:      deliveryID,
		UserID:          userID,
		TemplateName:    templateName,
		TemplateVersion: templateVersion,
		Data:            data,
		RequestID:       requestID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendSMS, payload), nil
}
