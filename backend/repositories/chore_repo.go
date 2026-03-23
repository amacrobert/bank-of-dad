package repositories

import (
	"errors"
	"fmt"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// ChoreAssignmentInfo holds the child_id and child name for a chore assignment.
type ChoreAssignmentInfo struct {
	ChildID   int64  `json:"child_id"`
	ChildName string `json:"child_name"`
}

// ChoreWithAssignments extends Chore with assignment info and a pending instance count.
type ChoreWithAssignments struct {
	models.Chore
	Assignments  []ChoreAssignmentInfo `json:"assignments" gorm:"-"`
	PendingCount int                   `json:"pending_count" gorm:"-"`
}

// ChoreRepo handles database operations for chores using GORM.
type ChoreRepo struct {
	db *gorm.DB
}

// NewChoreRepo creates a new ChoreRepo.
func NewChoreRepo(db *gorm.DB) *ChoreRepo {
	return &ChoreRepo{db: db}
}

// Create inserts a new chore into the database and returns the created entity.
func (r *ChoreRepo) Create(chore *models.Chore) (*models.Chore, error) {
	if err := r.db.Create(chore).Error; err != nil {
		return nil, fmt.Errorf("insert chore: %w", err)
	}
	return chore, nil
}

// GetByID retrieves a chore by its ID. Returns (nil, nil) if not found.
func (r *ChoreRepo) GetByID(id int64) (*models.Chore, error) {
	var chore models.Chore
	err := r.db.First(&chore, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get chore by id: %w", err)
	}
	return &chore, nil
}

// ListByFamily returns all chores for a family with their assignments and pending instance counts.
func (r *ChoreRepo) ListByFamily(familyID int64) ([]ChoreWithAssignments, error) {
	// 1. Fetch all chores for the family.
	var chores []models.Chore
	if err := r.db.Where("family_id = ?", familyID).Order("created_at DESC").Find(&chores).Error; err != nil {
		return nil, fmt.Errorf("list chores: %w", err)
	}
	if len(chores) == 0 {
		return []ChoreWithAssignments{}, nil
	}

	choreIDs := make([]int64, len(chores))
	for i, c := range chores {
		choreIDs[i] = c.ID
	}

	// 2. Batch-load assignments joined with children for names.
	type assignmentRow struct {
		ChoreID   int64  `gorm:"column:chore_id"`
		ChildID   int64  `gorm:"column:child_id"`
		ChildName string `gorm:"column:child_name"`
	}
	var assignmentRows []assignmentRow
	err := r.db.
		Table("chore_assignments a").
		Select("a.chore_id, a.child_id, c.first_name as child_name").
		Joins("JOIN children c ON c.id = a.child_id").
		Where("a.chore_id IN ?", choreIDs).
		Find(&assignmentRows).Error
	if err != nil {
		return nil, fmt.Errorf("list chore assignments: %w", err)
	}

	// 3. Batch-load pending instance counts grouped by chore_id.
	type pendingRow struct {
		ChoreID      int64 `gorm:"column:chore_id"`
		PendingCount int   `gorm:"column:pending_count"`
	}
	var pendingRows []pendingRow
	err = r.db.
		Table("chore_instances").
		Select("chore_id, COUNT(*) as pending_count").
		Where("chore_id IN ? AND status = ?", choreIDs, "pending_approval").
		Group("chore_id").
		Find(&pendingRows).Error
	if err != nil {
		return nil, fmt.Errorf("count pending instances: %w", err)
	}

	// 4. Assemble results.
	assignmentsByChore := make(map[int64][]ChoreAssignmentInfo)
	for _, row := range assignmentRows {
		assignmentsByChore[row.ChoreID] = append(assignmentsByChore[row.ChoreID], ChoreAssignmentInfo{
			ChildID:   row.ChildID,
			ChildName: row.ChildName,
		})
	}

	pendingByChore := make(map[int64]int)
	for _, row := range pendingRows {
		pendingByChore[row.ChoreID] = row.PendingCount
	}

	results := make([]ChoreWithAssignments, len(chores))
	for i, c := range chores {
		assignments := assignmentsByChore[c.ID]
		if assignments == nil {
			assignments = []ChoreAssignmentInfo{}
		}
		results[i] = ChoreWithAssignments{
			Chore:        c,
			Assignments:  assignments,
			PendingCount: pendingByChore[c.ID],
		}
	}

	return results, nil
}

// CreateAssignment inserts a new chore assignment and returns the created entity.
func (r *ChoreRepo) CreateAssignment(assignment *models.ChoreAssignment) (*models.ChoreAssignment, error) {
	if err := r.db.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("insert chore assignment: %w", err)
	}
	return assignment, nil
}

// DeleteAssignment removes a chore assignment by chore_id and child_id.
func (r *ChoreRepo) DeleteAssignment(choreID, childID int64) error {
	if err := r.db.Where("chore_id = ? AND child_id = ?", choreID, childID).Delete(&models.ChoreAssignment{}).Error; err != nil {
		return fmt.Errorf("delete chore assignment: %w", err)
	}
	return nil
}

// Delete removes a chore by its ID. Assignments cascade via FK.
func (r *ChoreRepo) Delete(id int64) error {
	if err := r.db.Delete(&models.Chore{}, id).Error; err != nil {
		return fmt.Errorf("delete chore: %w", err)
	}
	return nil
}

// Update updates a chore's name, description, reward_cents, recurrence, day_of_week, day_of_month, and is_active.
func (r *ChoreRepo) Update(chore *models.Chore) (*models.Chore, error) {
	err := r.db.Model(&models.Chore{}).
		Where("id = ?", chore.ID).
		Updates(map[string]interface{}{
			"name":         chore.Name,
			"description":  chore.Description,
			"reward_cents": chore.RewardCents,
			"recurrence":   chore.Recurrence,
			"day_of_week":  chore.DayOfWeek,
			"day_of_month": chore.DayOfMonth,
			"is_active":    chore.IsActive,
			"updated_at":   gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return nil, fmt.Errorf("update chore: %w", err)
	}
	return r.GetByID(chore.ID)
}

// ChoreWithAssignmentIDs holds a chore and its assigned child IDs for scheduler use.
type ChoreWithAssignmentIDs struct {
	models.Chore
	ChildIDs []int64
}

// ListAllActiveRecurring returns all active recurring chores with their assigned child IDs.
func (r *ChoreRepo) ListAllActiveRecurring() ([]ChoreWithAssignmentIDs, error) {
	var chores []models.Chore
	if err := r.db.Where("is_active = ? AND recurrence != ?", true, "one_time").Find(&chores).Error; err != nil {
		return nil, fmt.Errorf("list active recurring chores: %w", err)
	}
	if len(chores) == 0 {
		return nil, nil
	}

	choreIDs := make([]int64, len(chores))
	for i, c := range chores {
		choreIDs[i] = c.ID
	}

	var assignments []models.ChoreAssignment
	if err := r.db.Where("chore_id IN ?", choreIDs).Find(&assignments).Error; err != nil {
		return nil, fmt.Errorf("list assignments for recurring: %w", err)
	}

	assignmentsByChore := make(map[int64][]int64)
	for _, a := range assignments {
		assignmentsByChore[a.ChoreID] = append(assignmentsByChore[a.ChoreID], a.ChildID)
	}

	results := make([]ChoreWithAssignmentIDs, 0, len(chores))
	for _, c := range chores {
		childIDs := assignmentsByChore[c.ID]
		if len(childIDs) > 0 {
			results = append(results, ChoreWithAssignmentIDs{
				Chore:    c,
				ChildIDs: childIDs,
			})
		}
	}
	return results, nil
}

// SetActive toggles the is_active flag on a chore.
func (r *ChoreRepo) SetActive(id int64, active bool) error {
	err := r.db.Model(&models.Chore{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  active,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return fmt.Errorf("set chore active: %w", err)
	}
	return nil
}
