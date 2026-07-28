package model

import (
	"time"

	"github.com/google/uuid"
)

// UserPreferences maps directly to our JSONB column in Postgres.
// I'm using a dedicated struct here so GORM's JSON serializer handles the raw byte parsing for me,
// meaning I don't have to manually unmarshal ugly JSON strings into Go objects.
type UserPreferences struct {
	Channels struct {
		Email bool `json:"email"`
		SMS   bool `json:"sms"`
	} `json:"channels"`
}

type User struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email       string          `gorm:"type:varchar(255);not null;unique"`
	Phone       string          `gorm:"type:varchar(50)"`
	Preferences UserPreferences `gorm:"type:jsonb;serializer:json;not null;default:'{}'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
