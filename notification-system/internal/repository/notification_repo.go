package repository

import (
	"context"

	"github.com/mohdMusaiyab/notification-system/internal/model"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for database operations.
// Using an interface makes mocking for unit tests much easier later!
type NotificationRepository interface {
	Save(ctx context.Context, notification *model.Notification) error
}

// notificationRepository is the actual GORM implementation
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new instance of the repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Save inserts the notification into the database using GORM.
// WithContext ensures that if the user cancels the HTTP request early, the DB query also aborts to save resources.
func (r *notificationRepository) Save(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}
