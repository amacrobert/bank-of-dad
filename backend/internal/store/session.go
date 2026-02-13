package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"
)

type Session struct {
	Token     string
	UserType  string
	UserID    int64
	FamilyID  int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
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

	_, err := s.db.Exec(
		`INSERT INTO sessions (token, user_type, user_id, family_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		token, userType, userID, familyID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}

	return token, nil
}

// GetByToken looks up a session by its token using the read pool.
// Returns nil if the token does not exist or the session has expired.
func (s *SessionStore) GetByToken(token string) (*Session, error) {
	row := s.db.QueryRow(
		`SELECT token, user_type, user_id, family_id, created_at, expires_at
		 FROM sessions
		 WHERE token = $1 AND expires_at > $2`,
		token, time.Now(),
	)

	var sess Session
	err := row.Scan(&sess.Token, &sess.UserType, &sess.UserID, &sess.FamilyID, &sess.CreatedAt, &sess.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
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
	_, err := s.db.Exec(
		`UPDATE sessions SET family_id = $1 WHERE token = $2`,
		familyID, token,
	)
	if err != nil {
		return fmt.Errorf("update session family_id: %w", err)
	}
	return nil
}

// DeleteByToken removes a single session row by token.
func (s *SessionStore) DeleteByToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpired removes all sessions whose expires_at is in the past and
// returns the number of rows deleted.
func (s *SessionStore) DeleteExpired() (int64, error) {
	res, err := s.db.Exec(
		`DELETE FROM sessions WHERE expires_at < $1`,
		time.Now(),
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
