package store

import (
	"database/sql"
	"fmt"
	"time"
)

// SavingsGoal represents a child's savings target.
type SavingsGoal struct {
	ID          int64      `json:"id"`
	ChildID     int64      `json:"child_id"`
	Name        string     `json:"name"`
	TargetCents int64      `json:"target_cents"`
	SavedCents  int64      `json:"saved_cents"`
	Emoji       *string    `json:"emoji,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
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
	var targetDate sql.NullTime
	var completedAt sql.NullTime

	err := scanner.Scan(
		&g.ID, &g.ChildID, &g.Name, &g.TargetCents, &g.SavedCents,
		&emoji, &targetDate, &g.Status, &completedAt,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if emoji.Valid {
		g.Emoji = &emoji.String
	}
	if targetDate.Valid {
		t := targetDate.Time
		g.TargetDate = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		g.CompletedAt = &t
	}

	return &g, nil
}

const goalColumns = `id, child_id, name, target_cents, saved_cents, emoji, target_date, status, completed_at, created_at, updated_at`

// Create inserts a new savings goal and returns it.
func (s *SavingsGoalStore) Create(childID int64, name string, targetCents int64, emoji *string, targetDate *time.Time) (*SavingsGoal, error) {
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO savings_goals (child_id, name, target_cents, emoji, target_date)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		childID, name, targetCents, emoji, targetDate,
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
