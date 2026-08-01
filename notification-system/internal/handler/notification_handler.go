package handler

import (
	"net/http"
	"time"

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

// SendNotificationRequest defines the rich JSON payload (Updated for Stage 8)
type SendNotificationRequest struct {
	UserID       string                 `json:"user_id" binding:"required,uuid"`
	TemplateName string                 `json:"template_name" binding:"required"`
	Data         map[string]interface{} `json:"data"`
	
	// Optional scheduling parameter (RFC3339 format)
	SendAt       *time.Time             `json:"send_at,omitempty"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload. Ensure user_id is a valid UUID and send_at is a valid RFC3339 timestamp."})
		return
	}

	// 2.5 STAGE 8 VALIDATION: Ensure Time-Travel is strictly in the future!
	if req.SendAt != nil {
		if req.SendAt.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The 'send_at' timestamp cannot be in the past."})
			return
		}
	}

	// 3. Pass the validated data down to the Brain (Service Layer)
	err := h.service.ProcessNotification(c.Request.Context(), req.UserID, req.TemplateName, req.Data, idempotencyKey, req.SendAt)
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

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process notification: " + err.Error()})
		return
	}

	// 4. Respond instantly with a 202 Accepted
	c.JSON(http.StatusAccepted, gin.H{"status": "Notification event enqueued for dynamic routing"})
}
