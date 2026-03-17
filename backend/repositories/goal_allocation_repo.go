package repositories

import (
	"fmt"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// GoalAllocationRepo provides GORM-based access to the goal_allocations table.
type GoalAllocationRepo struct {
	db *gorm.DB
}

// NewGoalAllocationRepo creates a new GoalAllocationRepo.
func NewGoalAllocationRepo(db *gorm.DB) *GoalAllocationRepo {
	return &GoalAllocationRepo{db: db}
}

// ListByGoal returns all allocations for a goal, newest first.
func (r *GoalAllocationRepo) ListByGoal(goalID int64) ([]*models.GoalAllocation, error) {
	var allocations []*models.GoalAllocation
	err := r.db.Where("goal_id = ?", goalID).
		Order("created_at DESC").
		Find(&allocations).Error
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	return allocations, nil
}
