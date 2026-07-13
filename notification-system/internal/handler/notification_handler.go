package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdMusaiyab/notification-system/internal/service"
)

// SendNotificationRequest defines the strict JSON payload we expect from the client.
// We use Gin's built-in validator (go-playground/validator) to enforce rules.
type SendNotificationRequest struct {
	Recipient string `json:"recipient" binding:"required,email"` // Must exist and be a valid email format
	Message   string `json:"message" binding:"required,min=5"`   // Must exist and be at least 5 characters
}

// NotificationHandler acts as our HTTP Transport layer
type NotificationHandler struct {
	service service.NotificationService
}

// NewNotificationHandler creates a new handler injecting the core service layer
func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// HandleSendNotification is the Gin endpoint handler for POST /notification
func (h *NotificationHandler) HandleSendNotification(c *gin.Context) {
	var req SendNotificationRequest

	// 1. Validate the incoming JSON
	// ShouldBindJSON will check if the JSON matches our struct AND run our validation rules (like checking if it's a real email).
	if err := c.ShouldBindJSON(&req); err != nil {
		// If validation fails, immediately return a 400 Bad Request to the client
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 2. Pass the validated data down to our Core Service (The Brain)
	// This keeps the service layer completely unaware of the HTTP framework!
	err := h.service.ProcessNotification(c.Request.Context(), req.Recipient, req.Message)
	if err != nil {
		// If something went wrong in the DB or Provider, return a 500 error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process notification"})
		return
	}

	// 3. Return a successful 201 Created response
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Notification dispatched to internal processing pipeline",
	})
}
