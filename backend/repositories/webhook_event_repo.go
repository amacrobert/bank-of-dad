package repositories

import (
	"fmt"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// WebhookEventRepo provides GORM-based access to the stripe_webhook_events table.
type WebhookEventRepo struct {
	db *gorm.DB
}

// NewWebhookEventRepo creates a new WebhookEventRepo.
func NewWebhookEventRepo(db *gorm.DB) *WebhookEventRepo {
	return &WebhookEventRepo{db: db}
}

// RecordEvent inserts a new webhook event record. Uses ON CONFLICT DO NOTHING
// for idempotency (duplicate stripe_event_id is silently ignored).
func (r *WebhookEventRepo) RecordEvent(eventID, eventType string) error {
	evt := models.StripeWebhookEvent{
		StripeEventID: eventID,
		EventType:     eventType,
	}
	result := r.db.Where("stripe_event_id = ?", eventID).FirstOrCreate(&evt)
	if result.Error != nil {
		return fmt.Errorf("record webhook event: %w", result.Error)
	}
	return nil
}

// HasBeenProcessed checks if a Stripe webhook event has already been processed.
func (r *WebhookEventRepo) HasBeenProcessed(eventID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.StripeWebhookEvent{}).Where("stripe_event_id = ?", eventID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check webhook event processed: %w", err)
	}
	return count > 0, nil
}
