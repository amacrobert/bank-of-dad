package models

import "time"

// StripeWebhookEvent tracks processed Stripe webhook events for idempotency.
type StripeWebhookEvent struct {
	StripeEventID string    `gorm:"primaryKey;column:stripe_event_id"`
	EventType     string    `gorm:"not null"`
	ProcessedAt   time.Time `gorm:"autoCreateTime;column:processed_at"`
}

func (StripeWebhookEvent) TableName() string {
	return "stripe_webhook_events"
}
