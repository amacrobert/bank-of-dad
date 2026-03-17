package repositories

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"bank-of-dad/models"

	"errors"

	"gorm.io/gorm"
)

// HashToken computes the SHA-256 hex digest of a raw token string.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// RefreshTokenRepo provides GORM-based access to the refresh_tokens table.
type RefreshTokenRepo struct {
	db *gorm.DB
}

// NewRefreshTokenRepo creates a new RefreshTokenRepo.
func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// Create generates a cryptographically random refresh token, stores its
// SHA-256 hash in the database, and returns the raw token to send to the client.
func (r *RefreshTokenRepo) Create(userType string, userID int64, familyID int64, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	rawToken := base64.URLEncoding.EncodeToString(b)
	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(ttl)

	rt := models.RefreshToken{
		TokenHash: tokenHash,
		UserType:  userType,
		UserID:    userID,
		FamilyID:  familyID,
		ExpiresAt: expiresAt,
	}
	if err := r.db.Create(&rt).Error; err != nil {
		return "", fmt.Errorf("insert refresh token: %w", err)
	}

	return rawToken, nil
}

// Validate hashes the raw token, looks it up in the database, and returns the
// record if it exists and has not expired. Returns nil, nil if not found.
func (r *RefreshTokenRepo) Validate(rawToken string) (*models.RefreshToken, error) {
	tokenHash := HashToken(rawToken)

	var rt models.RefreshToken
	err := r.db.Where("token_hash = ? AND expires_at > NOW()", tokenHash).First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}
	return &rt, nil
}

// GetByHash retrieves a refresh token by its hash. Returns (nil, nil) if not found.
func (r *RefreshTokenRepo) GetByHash(tokenHash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token_hash = ?", tokenHash).First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}
	return &rt, nil
}

// DeleteByHash removes a refresh token by its hash.
func (r *RefreshTokenRepo) DeleteByHash(tokenHash string) error {
	if err := r.db.Where("token_hash = ?", tokenHash).Delete(&models.RefreshToken{}).Error; err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteAllForUser removes all refresh tokens for a given user.
func (r *RefreshTokenRepo) DeleteAllForUser(userType string, userID int64) error {
	if err := r.db.Where("user_type = ? AND user_id = ?", userType, userID).Delete(&models.RefreshToken{}).Error; err != nil {
		return fmt.Errorf("delete refresh tokens by user: %w", err)
	}
	return nil
}

// DeleteExpired removes all refresh tokens whose expires_at is in the past.
// Returns the number of rows deleted.
func (r *RefreshTokenRepo) DeleteExpired() (int64, error) {
	result := r.db.Where("expires_at < NOW()").Delete(&models.RefreshToken{})
	if result.Error != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// StartCleanupLoop runs DeleteExpired periodically in a goroutine.
func (r *RefreshTokenRepo) StartCleanupLoop(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			n, err := r.DeleteExpired()
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
