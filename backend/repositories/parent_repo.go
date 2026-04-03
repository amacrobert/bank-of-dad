package repositories

import (
	"errors"
	"fmt"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// ParentRepo provides GORM-based access to the parents table.
type ParentRepo struct {
	db *gorm.DB
}

// NewParentRepo creates a new ParentRepo.
func NewParentRepo(db *gorm.DB) *ParentRepo {
	return &ParentRepo{db: db}
}

// Create inserts a new parent with the given Google ID, email, and display name.
func (r *ParentRepo) Create(googleID, email, displayName string) (*models.Parent, error) {
	p := models.Parent{
		GoogleID:    googleID,
		Email:       email,
		DisplayName: displayName,
	}
	if err := r.db.Create(&p).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, fmt.Errorf("google account already registered")
		}
		return nil, fmt.Errorf("insert parent: %w", err)
	}
	return &p, nil
}

// GetByGoogleID retrieves a parent by Google ID. Returns (nil, nil) if not found.
func (r *ParentRepo) GetByGoogleID(googleID string) (*models.Parent, error) {
	var p models.Parent
	err := r.db.Where("google_id = ?", googleID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get parent by google id: %w", err)
	}
	return &p, nil
}

// GetByID retrieves a parent by ID. Returns (nil, nil) if not found.
func (r *ParentRepo) GetByID(id int64) (*models.Parent, error) {
	var p models.Parent
	err := r.db.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get parent by id: %w", err)
	}
	return &p, nil
}

// GetByFamilyID retrieves all parents belonging to a family.
func (r *ParentRepo) GetByFamilyID(familyID int64) ([]models.Parent, error) {
	var parents []models.Parent
	if err := r.db.Where("family_id = ?", familyID).Find(&parents).Error; err != nil {
		return nil, fmt.Errorf("get parents by family id: %w", err)
	}
	return parents, nil
}

// UpdateNotificationPrefs updates only the provided notification preference fields for a parent.
func (r *ParentRepo) UpdateNotificationPrefs(parentID int64, prefs map[string]bool) error {
	updates := make(map[string]interface{})
	for k, v := range prefs {
		updates[k] = v
	}
	if err := r.db.Model(&models.Parent{}).Where("id = ?", parentID).Updates(updates).Error; err != nil {
		return fmt.Errorf("update notification prefs: %w", err)
	}
	return nil
}

// SetFamilyID updates the family ID for a parent.
func (r *ParentRepo) SetFamilyID(parentID, familyID int64) error {
	if err := r.db.Model(&models.Parent{}).Where("id = ?", parentID).Update("family_id", familyID).Error; err != nil {
		return fmt.Errorf("set family id: %w", err)
	}
	return nil
}
