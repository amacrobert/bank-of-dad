package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
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

	sess.CreatedAt, err = time.Parse(time.DateTime, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	sess.ExpiresAt, err = time.Parse(time.DateTime, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	return &sess, nil
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
