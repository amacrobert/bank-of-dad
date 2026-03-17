package repositories

import (
	"fmt"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// AuthEventRepo provides GORM-based access to the auth_events table.
type AuthEventRepo struct {
	db *gorm.DB
}

// NewAuthEventRepo creates a new AuthEventRepo.
func NewAuthEventRepo(db *gorm.DB) *AuthEventRepo {
	return &AuthEventRepo{db: db}
}

// Log inserts a new auth event.
func (r *AuthEventRepo) Log(event models.AuthEvent) error {
	if err := r.db.Create(&event).Error; err != nil {
		return fmt.Errorf("insert auth event: %w", err)
	}
	return nil
}

// ListByUser returns auth events for a given family, ordered by created_at DESC,
// limited to the specified count.
func (r *AuthEventRepo) ListByUser(familyID int64, limit int) ([]models.AuthEvent, error) {
	var events []models.AuthEvent
	query := r.db.Where("family_id = ?", familyID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("query auth events: %w", err)
	}
	return events, nil
}

// GetEventsByFamily returns all auth events for a family, ordered by created_at DESC.
func (r *AuthEventRepo) GetEventsByFamily(familyID int64) ([]models.AuthEvent, error) {
	var events []models.AuthEvent
	if err := r.db.Where("family_id = ?", familyID).Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("query auth events: %w", err)
	}
	return events, nil
}
