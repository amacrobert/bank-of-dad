package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"
)

// parseTime tries multiple formats that SQLite might use for datetime values.
func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.DateTime,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

type Session struct {
	Token     string
	UserType  string
	UserID    int64
	FamilyID  int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	db *DB
}

func NewSessionStore(db *DB) *SessionStore {
	return &SessionStore{db: db}
}

// Create generates a cryptographically random session token, inserts it into
// the sessions table with the given metadata and TTL, and returns the token.
func (s *SessionStore) Create(userType string, userID int64, familyID int64, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(b)

	expiresAt := time.Now().Add(ttl)

	_, err := s.db.Write.Exec(
		`INSERT INTO sessions (token, user_type, user_id, family_id, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		token, userType, userID, familyID, expiresAt.UTC().Format(time.DateTime),
	)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}

	return token, nil
}

// GetByToken looks up a session by its token using the read pool.
// Returns nil if the token does not exist or the session has expired.
func (s *SessionStore) GetByToken(token string) (*Session, error) {
	row := s.db.Read.QueryRow(
		`SELECT token, user_type, user_id, family_id, created_at, expires_at
		 FROM sessions
		 WHERE token = ? AND expires_at > ?`,
		token, time.Now().UTC().Format(time.DateTime),
	)

	var sess Session
	var createdAt, expiresAt string
	err := row.Scan(&sess.Token, &sess.UserType, &sess.UserID, &sess.FamilyID, &createdAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}

	sess.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	sess.ExpiresAt, err = parseTime(expiresAt)
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	return &sess, nil
}

// ValidateSession looks up a session by token and returns the user info if valid.
// If the session exists but is expired, it lazily deletes the row from the DB.
// Implements middleware.SessionValidator interface.
func (s *SessionStore) ValidateSession(token string) (string, int64, int64, error) {
	sess, err := s.GetByToken(token)
	if err != nil {
		return "", 0, 0, err
	}
	if sess == nil {
		// Lazy cleanup: delete the expired row if it exists
		s.DeleteByToken(token) //nolint:errcheck // best-effort cleanup of expired session
		return "", 0, 0, fmt.Errorf("session not found or expired")
	}
	return sess.UserType, sess.UserID, sess.FamilyID, nil
}

// UpdateFamilyID updates the family_id for a session identified by token.
func (s *SessionStore) UpdateFamilyID(token string, familyID int64) error {
	_, err := s.db.Write.Exec(
		`UPDATE sessions SET family_id = ? WHERE token = ?`,
		familyID, token,
	)
	if err != nil {
		return fmt.Errorf("update session family_id: %w", err)
	}
	return nil
}

// DeleteByToken removes a single session row by token.
func (s *SessionStore) DeleteByToken(token string) error {
	_, err := s.db.Write.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpired removes all sessions whose expires_at is in the past and
// returns the number of rows deleted.
func (s *SessionStore) DeleteExpired() (int64, error) {
	res, err := s.db.Write.Exec(
		`DELETE FROM sessions WHERE expires_at < ?`,
		time.Now().UTC().Format(time.DateTime),
	)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return res.RowsAffected()
}

// StartCleanupLoop runs DeleteExpired periodically in a goroutine.
// It runs immediately on start, then every interval. Stops when stop is closed.
func (s *SessionStore) StartCleanupLoop(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			n, err := s.DeleteExpired()
			if err != nil {
				log.Printf("Session cleanup error: %v", err)
			} else if n > 0 {
				log.Printf("Cleaned up %d expired sessions", n)
			}
			select {
			case <-ticker.C:
			case <-stop:
				return
			}
		}
	}()
}
