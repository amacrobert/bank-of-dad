package repositories

import (
	"errors"
	"fmt"
	"time"

	"bank-of-dad/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrGoalNotFound             = errors.New("goal not found or not active")
	ErrInsufficientAvailable    = errors.New("amount exceeds available balance")
	ErrDeallocationExceedsSaved = errors.New("de-allocation exceeds saved amount")
	ErrZeroAllocation           = errors.New("allocation amount must be non-zero")
)

// UpdateGoalParams contains the optional fields for updating a savings goal.
type UpdateGoalParams struct {
	Name        *string
	TargetCents *int64
	Emoji       *string
	EmojiSet    bool // if true and Emoji is nil, clears the emoji
}

// AffectedGoalInfo represents a goal affected by a withdrawal.
type AffectedGoalInfo struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	CurrentSavedCents int64  `json:"current_saved_cents"`
	NewSavedCents     int64  `json:"new_saved_cents"`
}

// SavingsGoalRepo provides GORM-based access to the savings_goals table.
type SavingsGoalRepo struct {
	db *gorm.DB
}

// NewSavingsGoalRepo creates a new SavingsGoalRepo.
func NewSavingsGoalRepo(db *gorm.DB) *SavingsGoalRepo {
	return &SavingsGoalRepo{db: db}
}

// Create inserts a new savings goal and returns it.
func (r *SavingsGoalRepo) Create(childID int64, name string, targetCents int64, emoji *string) (*models.SavingsGoal, error) {
	goal := models.SavingsGoal{
		ChildID:     childID,
		Name:        name,
		TargetCents: targetCents,
		Emoji:       emoji,
	}
	if err := r.db.Create(&goal).Error; err != nil {
		return nil, fmt.Errorf("insert savings goal: %w", err)
	}
	return r.GetByID(goal.ID)
}

// GetByID retrieves a savings goal by its ID. Returns (nil, nil) if not found.
func (r *SavingsGoalRepo) GetByID(id int64) (*models.SavingsGoal, error) {
	var g models.SavingsGoal
	err := r.db.First(&g, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get savings goal by id: %w", err)
	}
	return &g, nil
}

// ListByChild returns all savings goals for a child, active first then completed, ordered by created_at ASC within each group.
func (r *SavingsGoalRepo) ListByChild(childID int64) ([]*models.SavingsGoal, error) {
	var goals []*models.SavingsGoal
	err := r.db.Where("child_id = ?", childID).
		Order("CASE WHEN status = 'active' THEN 0 ELSE 1 END").
		Order("created_at ASC").
		Find(&goals).Error
	if err != nil {
		return nil, fmt.Errorf("list savings goals: %w", err)
	}
	return goals, nil
}

// CountActiveByChild returns the count of active goals for a child.
func (r *SavingsGoalRepo) CountActiveByChild(childID int64) (int, error) {
	var count int64
	err := r.db.Model(&models.SavingsGoal{}).
		Where("child_id = ? AND status = 'active'", childID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count active goals: %w", err)
	}
	return int(count), nil
}

// Allocate atomically allocates (positive) or de-allocates (negative) funds to/from a goal.
// Returns the updated goal. If saved_cents >= target_cents after allocation, marks the goal completed.
func (r *SavingsGoalRepo) Allocate(goalID, childID, amountCents int64) (*models.SavingsGoal, error) {
	if amountCents == 0 {
		return nil, ErrZeroAllocation
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock and fetch the goal
		var goal models.SavingsGoal
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", goalID).
			First(&goal).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGoalNotFound
		}
		if err != nil {
			return fmt.Errorf("lock goal: %w", err)
		}
		if goal.Status != "active" {
			return ErrGoalNotFound
		}
		if goal.ChildID != childID {
			return ErrGoalNotFound
		}

		if amountCents > 0 {
			// Positive allocation: check available balance
			var availableBalance int64
			err = tx.Raw(
				`SELECT c.balance_cents - COALESCE(SUM(sg.saved_cents), 0)
				 FROM children c
				 LEFT JOIN savings_goals sg ON sg.child_id = c.id AND sg.status = 'active'
				 WHERE c.id = ?
				 GROUP BY c.balance_cents`, childID,
			).Scan(&availableBalance).Error
			if err != nil {
				return fmt.Errorf("check available balance: %w", err)
			}
			if amountCents > availableBalance {
				return ErrInsufficientAvailable
			}
		} else if -amountCents > goal.SavedCents {
			return ErrDeallocationExceedsSaved
		}

		// Update saved_cents
		newSavedCents := goal.SavedCents + amountCents
		if err := tx.Model(&models.SavingsGoal{}).Where("id = ?", goalID).
			Updates(map[string]interface{}{"saved_cents": newSavedCents, "updated_at": time.Now()}).Error; err != nil {
			return fmt.Errorf("update saved_cents: %w", err)
		}

		// Check for goal completion
		if newSavedCents >= goal.TargetCents {
			now := time.Now()
			if err := tx.Model(&models.SavingsGoal{}).Where("id = ?", goalID).
				Updates(map[string]interface{}{"status": "completed", "completed_at": now}).Error; err != nil {
				return fmt.Errorf("complete goal: %w", err)
			}
		}

		// Insert allocation record
		alloc := models.GoalAllocation{
			GoalID:      goalID,
			ChildID:     childID,
			AmountCents: amountCents,
		}
		if err := tx.Create(&alloc).Error; err != nil {
			return fmt.Errorf("insert allocation: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(goalID)
}

// Update partially updates an active savings goal.
// If target_cents is reduced to <= saved_cents, auto-completes the goal.
func (r *SavingsGoalRepo) Update(goalID, childID int64, params *UpdateGoalParams) (*models.SavingsGoal, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock and verify goal
		var goal models.SavingsGoal
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", goalID).
			First(&goal).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGoalNotFound
		}
		if err != nil {
			return fmt.Errorf("lock goal for update: %w", err)
		}
		if goal.Status != "active" || goal.ChildID != childID {
			return ErrGoalNotFound
		}

		// Build update map
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}
		if params.Name != nil {
			updates["name"] = *params.Name
		}
		if params.TargetCents != nil {
			updates["target_cents"] = *params.TargetCents
		}
		if params.EmojiSet {
			updates["emoji"] = params.Emoji
		}

		if err := tx.Model(&models.SavingsGoal{}).Where("id = ?", goalID).Updates(updates).Error; err != nil {
			return fmt.Errorf("update goal: %w", err)
		}

		// Check for auto-completion after target reduction
		if params.TargetCents != nil && goal.SavedCents >= *params.TargetCents {
			now := time.Now()
			if err := tx.Model(&models.SavingsGoal{}).Where("id = ?", goalID).
				Updates(map[string]interface{}{"status": "completed", "completed_at": now}).Error; err != nil {
				return fmt.Errorf("auto-complete goal: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByID(goalID)
}

// Delete removes an active or completed savings goal and returns the released saved_cents.
func (r *SavingsGoalRepo) Delete(goalID, childID int64) (int64, error) {
	var savedCents int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock and verify
		var goal models.SavingsGoal
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", goalID).
			First(&goal).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGoalNotFound
		}
		if err != nil {
			return fmt.Errorf("lock goal for delete: %w", err)
		}
		if (goal.Status != "active" && goal.Status != "completed") || goal.ChildID != childID {
			return ErrGoalNotFound
		}

		savedCents = goal.SavedCents

		// Delete the goal (cascades to goal_allocations via DB constraint)
		if err := tx.Delete(&models.SavingsGoal{}, goalID).Error; err != nil {
			return fmt.Errorf("delete goal: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return savedCents, nil
}

// ReduceGoalsProportionally reduces active goals' saved_cents proportionally to release totalToRelease cents.
// Records de-allocation entries for each affected goal. All within a single DB transaction.
func (r *SavingsGoalRepo) ReduceGoalsProportionally(childID, totalToRelease int64) error {
	if totalToRelease <= 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get all active goals with saved_cents > 0
		var goals []models.SavingsGoal
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("child_id = ? AND status = 'active' AND saved_cents > 0", childID).
			Find(&goals).Error
		if err != nil {
			return fmt.Errorf("query active goals: %w", err)
		}

		if len(goals) == 0 {
			return nil
		}

		var totalSaved int64
		for _, g := range goals {
			totalSaved += g.SavedCents
		}
		if totalSaved == 0 {
			return nil
		}

		// Cap the release to total saved
		release := totalToRelease
		if release > totalSaved {
			release = totalSaved
		}

		// Proportionally reduce each goal
		var released int64
		for i, g := range goals {
			var reduction int64
			if i == len(goals)-1 {
				reduction = release - released
			} else {
				reduction = g.SavedCents * release / totalSaved
			}

			if reduction <= 0 {
				continue
			}
			if reduction > g.SavedCents {
				reduction = g.SavedCents
			}

			newSaved := g.SavedCents - reduction
			if err := tx.Model(&models.SavingsGoal{}).Where("id = ?", g.ID).
				Updates(map[string]interface{}{"saved_cents": newSaved, "updated_at": time.Now()}).Error; err != nil {
				return fmt.Errorf("update goal saved_cents: %w", err)
			}

			// Record de-allocation
			alloc := models.GoalAllocation{
				GoalID:      g.ID,
				ChildID:     childID,
				AmountCents: -reduction,
			}
			if err := tx.Create(&alloc).Error; err != nil {
				return fmt.Errorf("insert de-allocation: %w", err)
			}

			released += reduction
		}

		return nil
	})
}

// GetTotalSavedByChild returns the sum of saved_cents across all active goals for a child.
func (r *SavingsGoalRepo) GetTotalSavedByChild(childID int64) (int64, error) {
	var total *int64
	err := r.db.Model(&models.SavingsGoal{}).
		Where("child_id = ? AND status = 'active'", childID).
		Select("COALESCE(SUM(saved_cents), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("get total saved: %w", err)
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}

// GetAffectedGoals returns active goals that would be impacted by reducing totalToRelease cents.
func (r *SavingsGoalRepo) GetAffectedGoals(childID, totalToRelease int64) ([]AffectedGoalInfo, error) {
	var goals []models.SavingsGoal
	err := r.db.Where("child_id = ? AND status = 'active' AND saved_cents > 0", childID).
		Order("id").
		Find(&goals).Error
	if err != nil {
		return nil, fmt.Errorf("query affected goals: %w", err)
	}

	var totalSaved int64
	for _, g := range goals {
		totalSaved += g.SavedCents
	}

	release := totalToRelease
	if release > totalSaved {
		release = totalSaved
	}

	var affected []AffectedGoalInfo
	var released int64
	for i, g := range goals {
		var reduction int64
		if i == len(goals)-1 {
			reduction = release - released
		} else {
			reduction = g.SavedCents * release / totalSaved
		}
		if reduction <= 0 {
			continue
		}
		if reduction > g.SavedCents {
			reduction = g.SavedCents
		}
		affected = append(affected, AffectedGoalInfo{
			ID:                g.ID,
			Name:              g.Name,
			CurrentSavedCents: g.SavedCents,
			NewSavedCents:     g.SavedCents - reduction,
		})
		released += reduction
	}

	return affected, nil
}

// GetAvailableBalance returns the child's balance minus the sum of active goals' saved_cents.
func (r *SavingsGoalRepo) GetAvailableBalance(childID int64) (int64, error) {
	var available int64
	err := r.db.Raw(
		`SELECT c.balance_cents - COALESCE(SUM(sg.saved_cents), 0)
		 FROM children c
		 LEFT JOIN savings_goals sg ON sg.child_id = c.id AND sg.status = 'active'
		 WHERE c.id = ?
		 GROUP BY c.balance_cents`, childID,
	).Scan(&available).Error
	if err != nil {
		return 0, fmt.Errorf("get available balance: %w", err)
	}
	return available, nil
}
