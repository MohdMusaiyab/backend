package worker

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

// A list of all task types our system supports.
// Using a constant prevents typos when pushing (Producer) or pulling (Consumer) from the queue.
const (
	TypeSendNotification = "notification:send"
)

// SendNotificationPayload is the actual data we serialize into JSON and push into Redis.
// Notice it doesn't have GORM or Gin tags, because it only cares about the Queue.
type SendNotificationPayload struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

// NewSendNotificationTask is a helper function to create the asynq.Task.
// We keep this here so the Service layer doesn't have to worry about JSON serialization.
func NewSendNotificationTask(recipient, message string) (*asynq.Task, error) {
	payload := SendNotificationPayload{
		Recipient: recipient,
		Message:   message,
	}

	// Serialize the struct into a JSON byte array
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create and return the asynq task with our predefined Type and serialized Payload
	return asynq.NewTask(TypeSendNotification, payloadBytes), nil
}
