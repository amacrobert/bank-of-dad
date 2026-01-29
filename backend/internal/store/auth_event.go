package store

import (
	"fmt"
	"time"
)

type AuthEvent struct {
	ID        int64
	EventType string // login_success, login_failure, logout, account_created, account_locked, password_reset, name_updated
	UserType  string // "parent" or "child"
	UserID    int64  // 0 for unknown users
	FamilyID  int64  // 0 for unknown families
	IPAddress string
	Details   string // non-sensitive context
	CreatedAt time.Time
}

type AuthEventStore struct {
	db *DB
}

func NewAuthEventStore(db *DB) *AuthEventStore {
	return &AuthEventStore{db: db}
}

func (s *AuthEventStore) LogEvent(event AuthEvent) error {
	_, err := s.db.Write.Exec(
		`INSERT INTO auth_events (event_type, user_type, user_id, family_id, ip_address, details, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.EventType,
		event.UserType,
		event.UserID,
		event.FamilyID,
		event.IPAddress,
		event.Details,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert auth event: %w", err)
	}
	return nil
}

func (s *AuthEventStore) GetEventsByFamily(familyID int64) ([]AuthEvent, error) {
	rows, err := s.db.Read.Query(
		`SELECT id, event_type, user_type, user_id, family_id, ip_address, details, created_at
		 FROM auth_events
		 WHERE family_id = ?
		 ORDER BY created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("query auth events: %w", err)
	}
	defer rows.Close()

	var events []AuthEvent
	for rows.Next() {
		var e AuthEvent
		if err := rows.Scan(
			&e.ID,
			&e.EventType,
			&e.UserType,
			&e.UserID,
			&e.FamilyID,
			&e.IPAddress,
			&e.Details,
			&e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan auth event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate auth events: %w", err)
	}

	return events, nil
}
