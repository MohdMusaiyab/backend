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
	
	// Methods for managing individual channel deliveries (Stage 5)
	SaveDeliveries(ctx context.Context, deliveries []model.NotificationDelivery) error
	GetDeliveryByID(ctx context.Context, id string) (*model.NotificationDelivery, error)
	UpdateDeliveryStatus(ctx context.Context, deliveryID string, status string) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Save inserts the notification. If the struct contains Deliveries, GORM saves them all in a single transaction!
func (r *notificationRepository) Save(ctx context.Context, n *model.Notification) error {
	// GORM automatically wraps .Create() in an ACID Transaction if it detects children (Deliveries)
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
	// We use Preload to eagerly fetch the child Deliveries associated with this ID
	err := r.db.WithContext(ctx).Preload("Deliveries").First(&notif, "id = ?", id).Error
	return &notif, err
}

func (r *notificationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ?", id).Update("status", status).Error
}

// SaveDeliveries allows the Router to save a batch of new channel deliveries
func (r *notificationRepository) SaveDeliveries(ctx context.Context, deliveries []model.NotificationDelivery) error {
	return r.db.WithContext(ctx).Create(&deliveries).Error
}

// GetDeliveryByID allows specific channel workers (like Email) to perform Idempotency Checks
func (r *notificationRepository) GetDeliveryByID(ctx context.Context, id string) (*model.NotificationDelivery, error) {
	var delivery model.NotificationDelivery
	err := r.db.WithContext(ctx).First(&delivery, "id = ?", id).Error
	return &delivery, err
}

// UpdateDeliveryStatus allows specific channel workers to mark themselves as Sent or Failed
func (r *notificationRepository) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status string) error {
	return r.db.WithContext(ctx).Model(&model.NotificationDelivery{}).Where("id = ?", deliveryID).Update("status", status).Error
}
