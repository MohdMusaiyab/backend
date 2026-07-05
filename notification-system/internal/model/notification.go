package model

import (
	"time"
)

// Notification maps directly to our notifications SQL table
type Notification struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Recipient string    `gorm:"type:varchar(255);not null" json:"recipient"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Status    string    `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
