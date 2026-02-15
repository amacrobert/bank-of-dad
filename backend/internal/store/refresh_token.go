package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

type RefreshToken struct {
	ID        int64
	TokenHash string
	UserType  string
	UserID    int64
	FamilyID  int64
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RefreshTokenStore struct {
	db *sql.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{db: db}
}

// HashToken computes the SHA-256 hex digest of a raw token string.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// Create generates a cryptographically random refresh token, stores its
// SHA-256 hash in the database, and returns the raw token to send to the client.
func (s *RefreshTokenStore) Create(userType string, userID int64, familyID int64, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	rawToken := base64.URLEncoding.EncodeToString(b)
	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(ttl)

	_, err := s.db.Exec(
		`INSERT INTO refresh_tokens (token_hash, user_type, user_id, family_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		tokenHash, userType, userID, familyID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("insert refresh token: %w", err)
	}

	return rawToken, nil
}

// Validate hashes the raw token, looks it up in the database, and returns the
// record if it exists and has not expired. Returns nil, nil if not found.
func (s *RefreshTokenStore) Validate(rawToken string) (*RefreshToken, error) {
	tokenHash := HashToken(rawToken)

	row := s.db.QueryRow(
		`SELECT id, token_hash, user_type, user_id, family_id, expires_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	)

	var rt RefreshToken
	err := row.Scan(&rt.ID, &rt.TokenHash, &rt.UserType, &rt.UserID, &rt.FamilyID, &rt.ExpiresAt, &rt.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan refresh token: %w", err)
	}

	return &rt, nil
}

// DeleteByHash removes a refresh token by its hash.
func (s *RefreshTokenStore) DeleteByHash(tokenHash string) error {
	_, err := s.db.Exec(`DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteByUser removes all refresh tokens for a given user.
func (s *RefreshTokenStore) DeleteByUser(userType string, userID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM refresh_tokens WHERE user_type = $1 AND user_id = $2`,
		userType, userID,
	)
	if err != nil {
		return fmt.Errorf("delete refresh tokens by user: %w", err)
	}
	return nil
}

// DeleteExpired removes all refresh tokens whose expires_at is in the past.
func (s *RefreshTokenStore) DeleteExpired() (int64, error) {
	res, err := s.db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return res.RowsAffected()
}

// StartCleanupLoop runs DeleteExpired periodically in a goroutine.
func (s *RefreshTokenStore) StartCleanupLoop(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			n, err := s.DeleteExpired()
			if err != nil {
				log.Printf("Refresh token cleanup error: %v", err)
			} else if n > 0 {
				log.Printf("Cleaned up %d expired refresh tokens", n)
			}
			select {
			case <-ticker.C:
			case <-stop:
				return
			}
		}
	}()
}
