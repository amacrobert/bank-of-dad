package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrGoalNotFound             = errors.New("goal not found or not active")
	ErrInsufficientAvailable    = errors.New("amount exceeds available balance")
	ErrDeallocationExceedsSaved = errors.New("de-allocation exceeds saved amount")
	ErrZeroAllocation           = errors.New("allocation amount must be non-zero")
)

// SavingsGoal represents a child's savings target.
type SavingsGoal struct {
	ID          int64      `json:"id"`
	ChildID     int64      `json:"child_id"`
	Name        string     `json:"name"`
	TargetCents int64      `json:"target_cents"`
	SavedCents  int64      `json:"saved_cents"`
	Emoji       *string    `json:"emoji,omitempty"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GoalAllocation represents a transfer of funds to/from a savings goal.
type GoalAllocation struct {
	ID          int64     `json:"id"`
	GoalID      int64     `json:"goal_id"`
	ChildID     int64     `json:"child_id"`
	AmountCents int64     `json:"amount_cents"`
	CreatedAt   time.Time `json:"created_at"`
}

// SavingsGoalStore handles database operations for savings goals.
type SavingsGoalStore struct {
	db *sql.DB
}

// NewSavingsGoalStore creates a new SavingsGoalStore.
func NewSavingsGoalStore(db *sql.DB) *SavingsGoalStore {
	return &SavingsGoalStore{db: db}
}

// scanGoal scans a single savings goal row, handling nullable fields.
func scanGoal(scanner interface{ Scan(dest ...interface{}) error }) (*SavingsGoal, error) {
	var g SavingsGoal
	var emoji sql.NullString
	var completedAt sql.NullTime

	err := scanner.Scan(
		&g.ID, &g.ChildID, &g.Name, &g.TargetCents, &g.SavedCents,
		&emoji, &g.Status, &completedAt,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if emoji.Valid {
		g.Emoji = &emoji.String
	}
	if completedAt.Valid {
		t := completedAt.Time
		g.CompletedAt = &t
	}

	return &g, nil
}

const goalColumns = `id, child_id, name, target_cents, saved_cents, emoji, status, completed_at, created_at, updated_at`

// Create inserts a new savings goal and returns it.
func (s *SavingsGoalStore) Create(childID int64, name string, targetCents int64, emoji *string) (*SavingsGoal, error) {
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO savings_goals (child_id, name, target_cents, emoji)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		childID, name, targetCents, emoji,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert savings goal: %w", err)
	}

	return s.GetByID(id)
}

// GetByID retrieves a savings goal by its ID.
func (s *SavingsGoalStore) GetByID(id int64) (*SavingsGoal, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`SELECT %s FROM savings_goals WHERE id = $1`, goalColumns),
		id,
	)
	g, err := scanGoal(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get savings goal by id: %w", err)
	}
	return g, nil
}

// ListByChild returns all savings goals for a child, active first then completed.
func (s *SavingsGoalStore) ListByChild(childID int64) ([]*SavingsGoal, error) {
	rows, err := s.db.Query(
		fmt.Sprintf(`SELECT %s FROM savings_goals WHERE child_id = $1
		ORDER BY CASE WHEN status = 'active' THEN 0 ELSE 1 END, created_at DESC`, goalColumns),
		childID,
	)
	if err != nil {
		return nil, fmt.Errorf("list savings goals: %w", err)
	}
	defer rows.Close()

	var goals []*SavingsGoal
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, fmt.Errorf("scan savings goal: %w", err)
		}
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

// CountActiveByChild returns the count of active goals for a child.
func (s *SavingsGoalStore) CountActiveByChild(childID int64) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM savings_goals WHERE child_id = $1 AND status = 'active'`,
		childID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active goals: %w", err)
	}
	return count, nil
}

// Allocate atomically allocates (positive) or de-allocates (negative) funds to/from a goal.
// Returns the updated goal. If saved_cents >= target_cents after allocation, marks the goal completed.
func (s *SavingsGoalStore) Allocate(goalID, childID, amountCents int64) (*SavingsGoal, error) {
	if amountCents == 0 {
		return nil, ErrZeroAllocation
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock and fetch the goal
	var goalChildID int64
	var savedCents, targetCents int64
	var status string
	err = tx.QueryRow(
		`SELECT child_id, saved_cents, target_cents, status FROM savings_goals WHERE id = $1 FOR UPDATE`,
		goalID,
	).Scan(&goalChildID, &savedCents, &targetCents, &status)
	if err == sql.ErrNoRows {
		return nil, ErrGoalNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock goal: %w", err)
	}
	if status != "active" {
		return nil, ErrGoalNotFound
	}
	if goalChildID != childID {
		return nil, ErrGoalNotFound
	}

	if amountCents > 0 {
		// Positive allocation: check available balance
		var availableBalance int64
		err = tx.QueryRow(
			`SELECT c.balance_cents - COALESCE(SUM(sg.saved_cents), 0)
			 FROM children c
			 LEFT JOIN savings_goals sg ON sg.child_id = c.id AND sg.status = 'active'
			 WHERE c.id = $1
			 GROUP BY c.balance_cents`,
			childID,
		).Scan(&availableBalance)
		if err != nil {
			return nil, fmt.Errorf("check available balance: %w", err)
		}
		if amountCents > availableBalance {
			return nil, ErrInsufficientAvailable
		}
	// Negative (de-allocation): check saved_cents >= abs(amount)
	} else if -amountCents > savedCents {
		return nil, ErrDeallocationExceedsSaved
	}

	// Update saved_cents
	newSavedCents := savedCents + amountCents
	_, err = tx.Exec(
		`UPDATE savings_goals SET saved_cents = $1, updated_at = NOW() WHERE id = $2`,
		newSavedCents, goalID,
	)
	if err != nil {
		return nil, fmt.Errorf("update saved_cents: %w", err)
	}

	// Check for goal completion
	if newSavedCents >= targetCents {
		_, err = tx.Exec(
			`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = $1`,
			goalID,
		)
		if err != nil {
			return nil, fmt.Errorf("complete goal: %w", err)
		}
	}

	// Insert allocation record
	_, err = tx.Exec(
		`INSERT INTO goal_allocations (goal_id, child_id, amount_cents) VALUES ($1, $2, $3)`,
		goalID, childID, amountCents,
	)
	if err != nil {
		return nil, fmt.Errorf("insert allocation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit allocation: %w", err)
	}

	return s.GetByID(goalID)
}

// UpdateGoalParams contains the optional fields for updating a savings goal.
type UpdateGoalParams struct {
	Name          *string
	TargetCents   *int64
	Emoji         *string
	EmojiSet      bool // if true and Emoji is nil, clears the emoji
}

// Update partially updates an active savings goal.
// If target_cents is reduced to <= saved_cents, auto-completes the goal.
func (s *SavingsGoalStore) Update(goalID, childID int64, params *UpdateGoalParams) (*SavingsGoal, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock and verify goal
	var currentChildID int64
	var savedCents int64
	var status string
	err = tx.QueryRow(
		`SELECT child_id, saved_cents, status FROM savings_goals WHERE id = $1 FOR UPDATE`,
		goalID,
	).Scan(&currentChildID, &savedCents, &status)
	if err == sql.ErrNoRows {
		return nil, ErrGoalNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock goal for update: %w", err)
	}
	if status != "active" || currentChildID != childID {
		return nil, ErrGoalNotFound
	}

	// Build dynamic update
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if params.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *params.Name)
		argIdx++
	}
	if params.TargetCents != nil {
		setClauses = append(setClauses, fmt.Sprintf("target_cents = $%d", argIdx))
		args = append(args, *params.TargetCents)
		argIdx++
	}
	if params.EmojiSet {
		setClauses = append(setClauses, fmt.Sprintf("emoji = $%d", argIdx))
		args = append(args, params.Emoji)
		argIdx++
	}
	// Execute update
	query := fmt.Sprintf("UPDATE savings_goals SET %s WHERE id = $%d",
		joinStrings(setClauses, ", "), argIdx)
	args = append(args, goalID)

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("update goal: %w", err)
	}

	// Check for auto-completion after target reduction
	if params.TargetCents != nil && savedCents >= *params.TargetCents {
		_, err = tx.Exec(
			`UPDATE savings_goals SET status = 'completed', completed_at = NOW() WHERE id = $1`,
			goalID,
		)
		if err != nil {
			return nil, fmt.Errorf("auto-complete goal: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update: %w", err)
	}

	return s.GetByID(goalID)
}

// Delete removes an active savings goal and returns the released saved_cents.
func (s *SavingsGoalStore) Delete(goalID, childID int64) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock and verify
	var currentChildID int64
	var savedCents int64
	var status string
	err = tx.QueryRow(
		`SELECT child_id, saved_cents, status FROM savings_goals WHERE id = $1 FOR UPDATE`,
		goalID,
	).Scan(&currentChildID, &savedCents, &status)
	if err == sql.ErrNoRows {
		return 0, ErrGoalNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("lock goal for delete: %w", err)
	}
	if (status != "active" && status != "completed") || currentChildID != childID {
		return 0, ErrGoalNotFound
	}

	// Delete the goal (cascades to goal_allocations)
	_, err = tx.Exec(`DELETE FROM savings_goals WHERE id = $1`, goalID)
	if err != nil {
		return 0, fmt.Errorf("delete goal: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit delete: %w", err)
	}

	return savedCents, nil
}

// ReduceGoalsProportionally reduces active goals' saved_cents proportionally to release totalToRelease cents.
// Records de-allocation entries for each affected goal. All within a single DB transaction.
func (s *SavingsGoalStore) ReduceGoalsProportionally(childID, totalToRelease int64) error {
	if totalToRelease <= 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get all active goals with saved_cents > 0
	rows, err := tx.Query(
		`SELECT id, saved_cents FROM savings_goals WHERE child_id = $1 AND status = 'active' AND saved_cents > 0 FOR UPDATE`,
		childID,
	)
	if err != nil {
		return fmt.Errorf("query active goals: %w", err)
	}

	type goalInfo struct {
		id         int64
		savedCents int64
	}
	var goals []goalInfo
	var totalSaved int64

	for rows.Next() {
		var g goalInfo
		if err := rows.Scan(&g.id, &g.savedCents); err != nil {
			rows.Close()
			return fmt.Errorf("scan goal: %w", err)
		}
		goals = append(goals, g)
		totalSaved += g.savedCents
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	if totalSaved == 0 || len(goals) == 0 {
		return nil
	}

	// Cap the release to total saved
	if totalToRelease > totalSaved {
		totalToRelease = totalSaved
	}

	// Proportionally reduce each goal
	var released int64
	for i, g := range goals {
		var reduction int64
		if i == len(goals)-1 {
			// Last goal gets the remainder to avoid rounding errors
			reduction = totalToRelease - released
		} else {
			reduction = g.savedCents * totalToRelease / totalSaved
		}

		if reduction <= 0 {
			continue
		}
		if reduction > g.savedCents {
			reduction = g.savedCents
		}

		newSaved := g.savedCents - reduction
		_, err = tx.Exec(
			`UPDATE savings_goals SET saved_cents = $1, updated_at = NOW() WHERE id = $2`,
			newSaved, g.id,
		)
		if err != nil {
			return fmt.Errorf("update goal saved_cents: %w", err)
		}

		// Record de-allocation
		_, err = tx.Exec(
			`INSERT INTO goal_allocations (goal_id, child_id, amount_cents) VALUES ($1, $2, $3)`,
			g.id, childID, -reduction,
		)
		if err != nil {
			return fmt.Errorf("insert de-allocation: %w", err)
		}

		released += reduction
	}

	return tx.Commit()
}

// GetTotalSavedByChild returns the sum of saved_cents across all active goals for a child.
func (s *SavingsGoalStore) GetTotalSavedByChild(childID int64) (int64, error) {
	var total int64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(saved_cents), 0) FROM savings_goals WHERE child_id = $1 AND status = 'active'`,
		childID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get total saved: %w", err)
	}
	return total, nil
}

// AffectedGoalInfo represents a goal affected by a withdrawal.
type AffectedGoalInfo struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	CurrentSavedCents int64  `json:"current_saved_cents"`
	NewSavedCents     int64  `json:"new_saved_cents"`
}

// GetAffectedGoals returns active goals that would be impacted by reducing totalToRelease cents.
func (s *SavingsGoalStore) GetAffectedGoals(childID, totalToRelease int64) ([]AffectedGoalInfo, error) {
	rows, err := s.db.Query(
		`SELECT id, name, saved_cents FROM savings_goals WHERE child_id = $1 AND status = 'active' AND saved_cents > 0 ORDER BY id`,
		childID,
	)
	if err != nil {
		return nil, fmt.Errorf("query affected goals: %w", err)
	}
	defer rows.Close()

	type goalRaw struct {
		id         int64
		name       string
		savedCents int64
	}
	var goals []goalRaw
	var totalSaved int64

	for rows.Next() {
		var g goalRaw
		if err := rows.Scan(&g.id, &g.name, &g.savedCents); err != nil {
			return nil, fmt.Errorf("scan goal: %w", err)
		}
		goals = append(goals, g)
		totalSaved += g.savedCents
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if totalToRelease > totalSaved {
		totalToRelease = totalSaved
	}

	var affected []AffectedGoalInfo
	var released int64
	for i, g := range goals {
		var reduction int64
		if i == len(goals)-1 {
			reduction = totalToRelease - released
		} else {
			reduction = g.savedCents * totalToRelease / totalSaved
		}
		if reduction <= 0 {
			continue
		}
		if reduction > g.savedCents {
			reduction = g.savedCents
		}
		affected = append(affected, AffectedGoalInfo{
			ID:               g.id,
			Name:             g.name,
			CurrentSavedCents: g.savedCents,
			NewSavedCents:     g.savedCents - reduction,
		})
		released += reduction
	}

	return affected, nil
}

// joinStrings joins strings with a separator (avoids importing strings package).
func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// GetAvailableBalance returns the child's balance minus the sum of active goals' saved_cents.
func (s *SavingsGoalStore) GetAvailableBalance(childID int64) (int64, error) {
	var available int64
	err := s.db.QueryRow(
		`SELECT c.balance_cents - COALESCE(SUM(sg.saved_cents), 0)
		 FROM children c
		 LEFT JOIN savings_goals sg ON sg.child_id = c.id AND sg.status = 'active'
		 WHERE c.id = $1
		 GROUP BY c.balance_cents`,
		childID,
	).Scan(&available)
	if err != nil {
		return 0, fmt.Errorf("get available balance: %w", err)
	}
	return available, nil
}

// ListAllocationsByGoal returns all allocations for a goal, newest first.
func (s *SavingsGoalStore) ListAllocationsByGoal(goalID int64) ([]*GoalAllocation, error) {
	rows, err := s.db.Query(
		`SELECT id, goal_id, child_id, amount_cents, created_at
		 FROM goal_allocations WHERE goal_id = $1 ORDER BY created_at DESC`,
		goalID,
	)
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	defer rows.Close()

	var allocations []*GoalAllocation
	for rows.Next() {
		var a GoalAllocation
		if err := rows.Scan(&a.ID, &a.GoalID, &a.ChildID, &a.AmountCents, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan allocation: %w", err)
		}
		allocations = append(allocations, &a)
	}
	return allocations, rows.Err()
}
