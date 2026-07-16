package model

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents the master "Broadcast Event" requested by the API
type Notification struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Recipient      string    `gorm:"not null"`
	Message        string    `gorm:"not null"`
	Status         string    `gorm:"not null;default:'pending'"`
	
	// IdempotencyKey prevents duplicate requests from being processed multiple times
	IdempotencyKey string    `gorm:"unique;column:idempotency_key"`
	
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// 1-to-Many Relationship: A single broadcast can result in multiple channel deliveries
	Deliveries     []NotificationDelivery `gorm:"foreignKey:NotificationID"`
}

// NotificationDelivery physically tracks the status of a specific channel (Email vs SMS)
type NotificationDelivery struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NotificationID uuid.UUID `gorm:"type:uuid;not null"`
	Channel        string    `gorm:"not null"` // e.g., 'email', 'sms'
	Status         string    `gorm:"not null;default:'pending'"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
