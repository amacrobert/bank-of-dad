package store

import (
	"database/sql"
	"fmt"
)

type WebhookEventStore struct {
	db *sql.DB
}

func NewWebhookEventStore(db *sql.DB) *WebhookEventStore {
	return &WebhookEventStore{db: db}
}

// IsProcessed checks if a Stripe webhook event has already been processed.
func (s *WebhookEventStore) IsProcessed(eventID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM stripe_webhook_events WHERE stripe_event_id = $1)`, eventID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check webhook event processed: %w", err)
	}
	return exists, nil
}

// MarkProcessed records that a Stripe webhook event has been processed.
// Uses INSERT ON CONFLICT DO NOTHING for idempotency.
func (s *WebhookEventStore) MarkProcessed(eventID, eventType string) error {
	_, err := s.db.Exec(
		`INSERT INTO stripe_webhook_events (stripe_event_id, event_type) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		eventID, eventType,
	)
	if err != nil {
		return fmt.Errorf("mark webhook event processed: %w", err)
	}
	return nil
}
