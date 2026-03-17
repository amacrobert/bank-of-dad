package repositories

import (
	"errors"
	"fmt"
	"strings"

	"bank-of-dad/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ChildRepo provides GORM-based data access for child accounts.
type ChildRepo struct {
	db *gorm.DB
}

// NewChildRepo creates a new ChildRepo backed by the given GORM database.
func NewChildRepo(db *gorm.DB) *ChildRepo {
	return &ChildRepo{db: db}
}

// CountByFamily returns the total number of children in a family.
func (r *ChildRepo) CountByFamily(familyID int64) (int, error) {
	var count int64
	err := r.db.Model(&models.Child{}).Where("family_id = ?", familyID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count children by family: %w", err)
	}
	return int(count), nil
}

// Create inserts a new child with a bcrypt-hashed password and returns the full record.
func (r *ChildRepo) Create(familyID int64, firstName, password string, avatar *string) (*models.Child, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	child := models.Child{
		FamilyID:     familyID,
		FirstName:    firstName,
		PasswordHash: string(hash),
		Avatar:       avatar,
	}

	if err := r.db.Create(&child).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("child named %q already exists in this family", firstName)
		}
		return nil, fmt.Errorf("insert child: %w", err)
	}

	return r.GetByID(child.ID)
}

// GetByID returns a child by primary key, or (nil, nil) if not found.
func (r *ChildRepo) GetByID(id int64) (*models.Child, error) {
	var child models.Child
	err := r.db.First(&child, id).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get child by id: %w", err)
	}
	return &child, nil
}

// GetByFamilyAndName returns a child matching the family and first name, or (nil, nil) if not found.
func (r *ChildRepo) GetByFamilyAndName(familyID int64, firstName string) (*models.Child, error) {
	var child models.Child
	err := r.db.Where("family_id = ? AND first_name = ?", familyID, firstName).First(&child).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get child by family and name: %w", err)
	}
	return &child, nil
}

// ListByFamily returns all children in a family, ordered by ID ascending.
func (r *ChildRepo) ListByFamily(familyID int64) ([]models.Child, error) {
	var children []models.Child
	err := r.db.Where("family_id = ?", familyID).Order("id").Find(&children).Error
	if err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}
	return children, nil
}

// CheckPassword compares a plaintext password against the child's stored bcrypt hash.
func (r *ChildRepo) CheckPassword(child *models.Child, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(child.PasswordHash), []byte(password)) == nil
}

// IncrementFailedAttempts atomically increments the failed login attempts counter
// and returns the new count.
func (r *ChildRepo) IncrementFailedAttempts(id int64) (int, error) {
	err := r.db.Exec(
		`UPDATE children SET failed_login_attempts = failed_login_attempts + 1, updated_at = NOW() WHERE id = ?`, id,
	).Error
	if err != nil {
		return 0, fmt.Errorf("increment failed attempts: %w", err)
	}

	var child models.Child
	err = r.db.Select("failed_login_attempts").First(&child, id).Error
	if err != nil {
		return 0, fmt.Errorf("read failed attempts: %w", err)
	}
	return child.FailedLoginAttempts, nil
}

// LockAccount sets is_locked to true for the given child.
func (r *ChildRepo) LockAccount(id int64) error {
	err := r.db.Exec(
		`UPDATE children SET is_locked = TRUE, updated_at = NOW() WHERE id = ?`, id,
	).Error
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}
	return nil
}

// ResetFailedAttempts sets the failed login attempts counter to zero.
func (r *ChildRepo) ResetFailedAttempts(id int64) error {
	err := r.db.Exec(
		`UPDATE children SET failed_login_attempts = 0, updated_at = NOW() WHERE id = ?`, id,
	).Error
	if err != nil {
		return fmt.Errorf("reset failed attempts: %w", err)
	}
	return nil
}

// UpdatePassword hashes the new password and resets is_locked and failed_login_attempts.
func (r *ChildRepo) UpdatePassword(id int64, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	err = r.db.Exec(
		`UPDATE children SET password_hash = ?, is_locked = FALSE, failed_login_attempts = 0, updated_at = NOW() WHERE id = ?`,
		string(hash), id,
	).Error
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// UpdateNameAndAvatar updates the child's name and optionally their avatar.
// When avatarSet is true, avatar is applied (nil clears, non-nil sets).
// When avatarSet is false, the avatar is left unchanged.
func (r *ChildRepo) UpdateNameAndAvatar(id, familyID int64, newName string, avatar *string, avatarSet bool) error {
	var err error
	if avatarSet {
		err = r.db.Exec(
			`UPDATE children SET first_name = ?, avatar = ?, updated_at = NOW() WHERE id = ? AND family_id = ?`,
			newName, avatar, id, familyID,
		).Error
	} else {
		err = r.db.Exec(
			`UPDATE children SET first_name = ?, updated_at = NOW() WHERE id = ? AND family_id = ?`,
			newName, id, familyID,
		).Error
	}
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("child named %q already exists in this family", newName)
		}
		return fmt.Errorf("update name: %w", err)
	}
	return nil
}

// UpdateTheme sets the child's visual theme preference.
func (r *ChildRepo) UpdateTheme(childID int64, theme string) error {
	result := r.db.Exec(
		`UPDATE children SET theme = ?, updated_at = NOW() WHERE id = ?`,
		theme, childID,
	)
	if result.Error != nil {
		return fmt.Errorf("update theme: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("child not found")
	}
	return nil
}

// UpdateAvatar sets the child's avatar emoji (or clears it if avatar is nil).
func (r *ChildRepo) UpdateAvatar(childID int64, avatar *string) error {
	result := r.db.Exec(
		`UPDATE children SET avatar = ?, updated_at = NOW() WHERE id = ?`,
		avatar, childID,
	)
	if result.Error != nil {
		return fmt.Errorf("update avatar: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("child not found")
	}
	return nil
}

// Delete permanently removes a child and all associated data in a single
// atomic transaction.
func (r *ChildRepo) Delete(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM refresh_tokens WHERE user_type = 'child' AND user_id = ?`, id).Error; err != nil {
			return fmt.Errorf("delete child refresh tokens: %w", err)
		}
		if err := tx.Exec(`DELETE FROM auth_events WHERE user_type = 'child' AND user_id = ?`, id).Error; err != nil {
			return fmt.Errorf("delete child auth events: %w", err)
		}
		if err := tx.Exec(`DELETE FROM children WHERE id = ?`, id).Error; err != nil {
			return fmt.Errorf("delete child: %w", err)
		}
		return nil
	})
}

// SetDisabled sets the is_disabled flag on a child.
func (r *ChildRepo) SetDisabled(childID int64, disabled bool) error {
	err := r.db.Exec(
		`UPDATE children SET is_disabled = ?, updated_at = NOW() WHERE id = ?`,
		disabled, childID,
	).Error
	if err != nil {
		return fmt.Errorf("set disabled: %w", err)
	}
	return nil
}

// CountEnabledByFamily returns the number of enabled (non-disabled) children in a family.
func (r *ChildRepo) CountEnabledByFamily(familyID int64) (int, error) {
	var count int64
	err := r.db.Model(&models.Child{}).Where("family_id = ? AND is_disabled = ?", familyID, false).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count enabled children by family: %w", err)
	}
	return int(count), nil
}

// EnableAllChildren sets is_disabled = FALSE for all children in a family.
func (r *ChildRepo) EnableAllChildren(familyID int64) error {
	err := r.db.Exec(
		`UPDATE children SET is_disabled = FALSE, updated_at = NOW() WHERE family_id = ? AND is_disabled = TRUE`,
		familyID,
	).Error
	if err != nil {
		return fmt.Errorf("enable all children: %w", err)
	}
	return nil
}

// DisableExcessChildren disables all children beyond the earliest `limit` (by ID) in a family.
func (r *ChildRepo) DisableExcessChildren(familyID int64, limit int) error {
	err := r.db.Exec(
		`UPDATE children SET is_disabled = TRUE, updated_at = NOW()
		 WHERE family_id = ? AND id NOT IN (
			SELECT id FROM children WHERE family_id = ? ORDER BY id ASC LIMIT ?
		 )`,
		familyID, familyID, limit,
	).Error
	if err != nil {
		return fmt.Errorf("disable excess children: %w", err)
	}
	return nil
}

// ReconcileChildLimits enables the earliest `limit` children (by ID) and disables the rest.
// This is the single source of truth for enforcing the free tier child limit.
func (r *ChildRepo) ReconcileChildLimits(familyID int64, limit int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Enable the earliest `limit` children
		if err := tx.Exec(
			`UPDATE children SET is_disabled = FALSE, updated_at = NOW()
			 WHERE family_id = ? AND is_disabled = TRUE AND id IN (
				SELECT id FROM children WHERE family_id = ? ORDER BY id ASC LIMIT ?
			 )`,
			familyID, familyID, limit,
		).Error; err != nil {
			return fmt.Errorf("enable earliest children: %w", err)
		}

		// Disable the rest
		if err := tx.Exec(
			`UPDATE children SET is_disabled = TRUE, updated_at = NOW()
			 WHERE family_id = ? AND is_disabled = FALSE AND id NOT IN (
				SELECT id FROM children WHERE family_id = ? ORDER BY id ASC LIMIT ?
			 )`,
			familyID, familyID, limit,
		).Error; err != nil {
			return fmt.Errorf("disable excess children: %w", err)
		}

		return nil
	})
}

// GetBalance returns the current balance in cents for a child.
func (r *ChildRepo) GetBalance(id int64) (int64, error) {
	var child models.Child
	err := r.db.Select("balance_cents").First(&child, id).Error
	if err != nil {
		if isNotFound(err) {
			return 0, fmt.Errorf("child not found")
		}
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return child.BalanceCents, nil
}

// isNotFound checks whether the error is a GORM record-not-found error.
func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
