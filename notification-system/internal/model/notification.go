package model

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a single message sent to a user
type Notification struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Recipient      string    `gorm:"not null"`
	Message        string    `gorm:"not null"`
	Status         string    `gorm:"not null;default:'pending'"`
	
	// IdempotencyKey prevents duplicate requests from being processed multiple times
	IdempotencyKey string    `gorm:"unique;column:idempotency_key"`
	
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
