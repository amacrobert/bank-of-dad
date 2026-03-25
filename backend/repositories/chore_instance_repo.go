package repositories

import (
	"errors"
	"fmt"
	"time"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInstanceNotFound        = errors.New("chore instance not found")
)

// ChoreInstanceWithDetails extends ChoreInstance with joined chore fields.
type ChoreInstanceWithDetails struct {
	models.ChoreInstance
	ChoreName        string  `json:"chore_name" gorm:"column:chore_name"`
	ChoreDescription *string `json:"chore_description" gorm:"column:chore_description"`
}

// PendingChoreInstance extends ChoreInstance with joined chore and child name fields.
type PendingChoreInstance struct {
	models.ChoreInstance
	ChoreName string `json:"chore_name" gorm:"column:chore_name"`
	ChildName string `json:"child_name" gorm:"column:child_name"`
}

// ChoreInstanceRepo handles database operations for chore instances using GORM.
type ChoreInstanceRepo struct {
	db *gorm.DB
}

// NewChoreInstanceRepo creates a new ChoreInstanceRepo.
func NewChoreInstanceRepo(db *gorm.DB) *ChoreInstanceRepo {
	return &ChoreInstanceRepo{db: db}
}

// CreateInstance inserts a new chore instance and returns the created entity.
func (r *ChoreInstanceRepo) CreateInstance(instance *models.ChoreInstance) (*models.ChoreInstance, error) {
	if err := r.db.Create(instance).Error; err != nil {
		return nil, fmt.Errorf("insert chore instance: %w", err)
	}
	return instance, nil
}

// GetByID retrieves a chore instance by its ID. Returns (nil, nil) if not found.
func (r *ChoreInstanceRepo) GetByID(id int64) (*models.ChoreInstance, error) {
	var instance models.ChoreInstance
	err := r.db.First(&instance, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chore instance by id: %w", err)
	}
	return &instance, nil
}

// ListByChild returns non-expired instances for a child, grouped by status.
// available = status 'available', pending = status 'pending_approval', completed = status 'approved'.
// Each group is ordered by created_at DESC. Includes chore name and description.
func (r *ChoreInstanceRepo) ListByChild(childID int64) (available []ChoreInstanceWithDetails, pending []ChoreInstanceWithDetails, completed []ChoreInstanceWithDetails, err error) {
	queryByStatus := func(status models.ChoreInstanceStatus) ([]ChoreInstanceWithDetails, error) {
		var results []ChoreInstanceWithDetails
		err := r.db.Table("chore_instances").
			Select("chore_instances.*, chores.name as chore_name, chores.description as chore_description").
			Joins("JOIN chores ON chores.id = chore_instances.chore_id").
			Where("chore_instances.child_id = ? AND chore_instances.status = ?", childID, status).
			Order("chore_instances.created_at DESC").
			Find(&results).Error
		if err != nil {
			return nil, err
		}
		return results, nil
	}

	available, err = queryByStatus(models.ChoreInstanceStatusAvailable)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list available instances: %w", err)
	}
	pending, err = queryByStatus(models.ChoreInstanceStatusPendingApproval)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list pending instances: %w", err)
	}
	completed, err = queryByStatus(models.ChoreInstanceStatusApproved)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("list completed instances: %w", err)
	}

	return available, pending, completed, nil
}

// ListPendingByFamily returns all pending_approval instances for children in a given family.
// Includes chore name and child name. Ordered by completed_at ASC (oldest first).
func (r *ChoreInstanceRepo) ListPendingByFamily(familyID int64) ([]PendingChoreInstance, error) {
	var results []PendingChoreInstance
	err := r.db.Table("chore_instances").
		Select("chore_instances.*, chores.name as chore_name, children.first_name as child_name").
		Joins("JOIN chores ON chores.id = chore_instances.chore_id").
		Joins("JOIN children ON children.id = chore_instances.child_id").
		Where("children.family_id = ? AND chore_instances.status = ?", familyID, "pending_approval").
		Order("chore_instances.completed_at ASC").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list pending instances by family: %w", err)
	}
	return results, nil
}

// MarkComplete transitions an instance from 'available' to 'pending_approval'.
// Sets completed_at to NOW() and clears rejection_reason.
// Verifies the instance belongs to the specified child and is currently 'available'.
func (r *ChoreInstanceRepo) MarkComplete(instanceID, childID int64) error {
	result := r.db.Model(&models.ChoreInstance{}).
		Where("id = ? AND child_id = ? AND status = ?", instanceID, childID, "available").
		Updates(map[string]interface{}{
			"status":           models.ChoreInstanceStatusPendingApproval,
			"completed_at":     gorm.Expr("NOW()"),
			"rejection_reason": nil,
			"updated_at":       gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return fmt.Errorf("mark chore instance complete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvalidStatusTransition
	}
	return nil
}

// Approve transitions an instance from 'pending_approval' to 'approved'.
// Sets reviewed_at to NOW(), reviewed_by_parent_id, and transaction_id.
// Verifies the instance is currently 'pending_approval'.
func (r *ChoreInstanceRepo) Approve(instanceID int64, parentID int64, transactionID *int64) error {
	result := r.db.Model(&models.ChoreInstance{}).
		Where("id = ? AND status = ?", instanceID, "pending_approval").
		Updates(map[string]interface{}{
			"status":               models.ChoreInstanceStatusApproved,
			"reviewed_at":          gorm.Expr("NOW()"),
			"reviewed_by_parent_id": parentID,
			"transaction_id":       transactionID,
			"updated_at":           gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return fmt.Errorf("approve chore instance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvalidStatusTransition
	}
	return nil
}

// Reject transitions an instance from 'pending_approval' back to 'available'.
// Sets reviewed_at to NOW(), reviewed_by_parent_id, rejection_reason, and clears completed_at.
// Verifies the instance is currently 'pending_approval'.
func (r *ChoreInstanceRepo) Reject(instanceID int64, parentID int64, reason string) error {
	result := r.db.Model(&models.ChoreInstance{}).
		Where("id = ? AND status = ?", instanceID, "pending_approval").
		Updates(map[string]interface{}{
			"status":               models.ChoreInstanceStatusAvailable,
			"reviewed_at":          gorm.Expr("NOW()"),
			"reviewed_by_parent_id": parentID,
			"rejection_reason":     reason,
			"completed_at":         nil,
			"updated_at":           gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return fmt.Errorf("reject chore instance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrInvalidStatusTransition
	}
	return nil
}

// ExistsForPeriod checks if an instance already exists for the given chore, child, and period_start.
func (r *ChoreInstanceRepo) ExistsForPeriod(choreID, childID int64, periodStart time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.ChoreInstance{}).
		Where("chore_id = ? AND child_id = ? AND period_start = ?", choreID, childID, periodStart).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check instance exists for period: %w", err)
	}
	return count > 0, nil
}

// ChoreEarning represents a single earning from a completed chore.
type ChoreEarning struct {
	ChoreName  string    `json:"chore_name" gorm:"column:chore_name"`
	RewardCents int      `json:"reward_cents" gorm:"column:reward_cents"`
	ApprovedAt time.Time `json:"approved_at" gorm:"column:approved_at"`
}

// GetEarnings returns chore earnings for a child: total earned, count completed, and recent approved instances.
func (r *ChoreInstanceRepo) GetEarnings(childID int64, recentLimit int) (totalCents int64, completedCount int64, recent []ChoreEarning, err error) {
	// Aggregate totals
	type aggregateResult struct {
		TotalCents     int64 `gorm:"column:total_cents"`
		CompletedCount int64 `gorm:"column:completed_count"`
	}
	var agg aggregateResult
	err = r.db.Table("chore_instances").
		Select("COALESCE(SUM(reward_cents), 0) as total_cents, COUNT(*) as completed_count").
		Where("child_id = ? AND status = ?", childID, "approved").
		Scan(&agg).Error
	if err != nil {
		return 0, 0, nil, fmt.Errorf("get chore earnings aggregates: %w", err)
	}

	// Recent approved instances
	err = r.db.Table("chore_instances").
		Select("chores.name as chore_name, chore_instances.reward_cents, chore_instances.reviewed_at as approved_at").
		Joins("JOIN chores ON chores.id = chore_instances.chore_id").
		Where("chore_instances.child_id = ? AND chore_instances.status = ?", childID, "approved").
		Order("chore_instances.reviewed_at DESC").
		Limit(recentLimit).
		Scan(&recent).Error
	if err != nil {
		return 0, 0, nil, fmt.Errorf("get recent chore earnings: %w", err)
	}

	return agg.TotalCents, agg.CompletedCount, recent, nil
}

// ListCompletedByFamily returns approved chore instances for a family with pagination.
// Includes chore name and child name. Ordered by reviewed_at DESC (most recent first).
// Returns instances for the current page and the total count of completed instances.
func (r *ChoreInstanceRepo) ListCompletedByFamily(familyID int64, limit, offset int) ([]PendingChoreInstance, int64, error) {
	base := r.db.Table("chore_instances").
		Joins("JOIN chores ON chores.id = chore_instances.chore_id").
		Joins("JOIN children ON children.id = chore_instances.child_id").
		Where("children.family_id = ? AND chore_instances.status = ?", familyID, "approved")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count completed instances by family: %w", err)
	}

	var results []PendingChoreInstance
	err := r.db.Table("chore_instances").
		Select("chore_instances.*, chores.name as chore_name, children.first_name as child_name").
		Joins("JOIN chores ON chores.id = chore_instances.chore_id").
		Joins("JOIN children ON children.id = chore_instances.child_id").
		Where("children.family_id = ? AND chore_instances.status = ?", familyID, "approved").
		Order("chore_instances.reviewed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&results).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list completed instances by family: %w", err)
	}
	return results, total, nil
}

// DeleteByChoreID deletes all instances for a specific chore (used before chore deletion).
func (r *ChoreInstanceRepo) DeleteByChoreID(choreID int64) error {
	if err := r.db.Where("chore_id = ?", choreID).Delete(&models.ChoreInstance{}).Error; err != nil {
		return fmt.Errorf("delete instances by chore id: %w", err)
	}
	return nil
}

// ExpireByPeriod updates all 'available' instances where period_end < before to 'expired'.
// Returns the count of expired rows.
func (r *ChoreInstanceRepo) ExpireByPeriod(before time.Time) (int64, error) {
	result := r.db.Model(&models.ChoreInstance{}).
		Where("status = ? AND period_end < ?", "available", before).
		Updates(map[string]interface{}{
			"status":     models.ChoreInstanceStatusExpired,
			"updated_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return 0, fmt.Errorf("expire chore instances: %w", result.Error)
	}
	return result.RowsAffected, nil
}
