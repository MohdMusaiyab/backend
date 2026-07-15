package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/mohdMusaiyab/notification-system/internal/model"
	"gorm.io/gorm"
)

// ErrDuplicateIdempotencyKey lets the service layer cleanly detect a duplicate request
var ErrDuplicateIdempotencyKey = errors.New("idempotency key already exists")

type NotificationRepository interface {
	Save(ctx context.Context, n *model.Notification) error
	GetByID(ctx context.Context, id string) (*model.Notification, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Save(ctx context.Context, n *model.Notification) error {
	err := r.db.WithContext(ctx).Create(n).Error
	if err != nil {
		// Catching the Postgres unique constraint violation so the API layer knows it's a duplicate request
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrDuplicateIdempotencyKey
		}
		return err
	}
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id string) (*model.Notification, error) {
	var notif model.Notification
	// Fetching the record by ID so the worker can check if it was already sent
	err := r.db.WithContext(ctx).First(&notif, "id = ?", id).Error
	return &notif, err
}

func (r *notificationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	// Updating only the status column to keep the DB operation fast and efficient
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ?", id).Update("status", status).Error
}
