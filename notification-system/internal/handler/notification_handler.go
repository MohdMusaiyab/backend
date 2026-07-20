package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohdMusaiyab/notification-system/internal/service"
)

// NotificationHandler handles incoming HTTP requests
type NotificationHandler struct {
	service service.NotificationService
}

// NewNotificationHandler creates a new handler instance
func NewNotificationHandler(service service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// SendNotificationRequest defines the expected JSON payload
type SendNotificationRequest struct {
	Recipient string `json:"recipient" binding:"required,email"`
	Message   string `json:"message" binding:"required"`
}

// HandleSendNotification is the Gin controller for POST /notification
func (h *NotificationHandler) HandleSendNotification(c *gin.Context) {
	// 1. Enforce Idempotency at the front door!
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is strictly required to prevent duplicates"})
		return
	}

	var req SendNotificationRequest
	// 2. Validate the incoming JSON structure
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// 3. Pass the validated data AND the idempotency key down to the Brain (Service Layer)
	err := h.service.ProcessNotification(c.Request.Context(), req.Recipient, req.Message, idempotencyKey)
	if err != nil {
		// If the Brain tells us it's a duplicate, we calmly return a 200 OK
		if err == service.ErrDuplicateRequest {
			c.JSON(http.StatusOK, gin.H{"status": "Duplicate request ignored, already processing"})
			return
		}
		
		// If the Brain tells us the queues are full, we trigger LOAD SHEDDING
		if err == service.ErrSystemOverloaded {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "System is currently experiencing extreme traffic. Please try again later.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process notification"})
		return
	}

	// 4. Respond instantly with a 202 Accepted
	c.JSON(http.StatusAccepted, gin.H{"status": "Notification enqueued for delivery"})
}
