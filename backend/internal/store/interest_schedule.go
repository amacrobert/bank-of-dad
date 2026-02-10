package store

import (
	"database/sql"
	"fmt"
	"time"
)

// InterestSchedule represents an interest accrual schedule for a child.
type InterestSchedule struct {
	ID         int64          `json:"id"`
	ChildID    int64          `json:"child_id"`
	ParentID   int64          `json:"parent_id"`
	Frequency  Frequency      `json:"frequency"`
	DayOfWeek  *int           `json:"day_of_week,omitempty"`
	DayOfMonth *int           `json:"day_of_month,omitempty"`
	Status     ScheduleStatus `json:"status"`
	NextRunAt  *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// InterestScheduleStore handles database operations for interest accrual schedules.
type InterestScheduleStore struct {
	db *DB
}

// NewInterestScheduleStore creates a new InterestScheduleStore.
func NewInterestScheduleStore(db *DB) *InterestScheduleStore {
	return &InterestScheduleStore{db: db}
}

// Create inserts a new interest schedule.
func (s *InterestScheduleStore) Create(sched *InterestSchedule) (*InterestSchedule, error) {
	var dayOfWeek, dayOfMonth interface{}
	if sched.DayOfWeek != nil {
		dayOfWeek = *sched.DayOfWeek
	}
	if sched.DayOfMonth != nil {
		dayOfMonth = *sched.DayOfMonth
	}
	var nextRunAt interface{}
	if sched.NextRunAt != nil {
		nextRunAt = sched.NextRunAt.UTC().Format(time.RFC3339)
	}

	result, err := s.db.Write.Exec(
		`INSERT INTO interest_schedules
		(child_id, parent_id, frequency, day_of_week, day_of_month, status, next_run_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sched.ChildID, sched.ParentID, sched.Frequency,
		dayOfWeek, dayOfMonth, sched.Status, nextRunAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert interest schedule: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

// GetByID retrieves an interest schedule by its ID.
func (s *InterestScheduleStore) GetByID(id int64) (*InterestSchedule, error) {
	return s.scanOne(
		`SELECT id, child_id, parent_id, frequency, day_of_week, day_of_month,
		        status, next_run_at, created_at, updated_at
		 FROM interest_schedules WHERE id = ?`, id,
	)
}

// GetByChildID retrieves the interest schedule for a child (any status), or nil if none exists.
func (s *InterestScheduleStore) GetByChildID(childID int64) (*InterestSchedule, error) {
	return s.scanOne(
		`SELECT id, child_id, parent_id, frequency, day_of_week, day_of_month,
		        status, next_run_at, created_at, updated_at
		 FROM interest_schedules WHERE child_id = ?`, childID,
	)
}

// Update updates a schedule's frequency, day, and next_run_at fields.
func (s *InterestScheduleStore) Update(sched *InterestSchedule) (*InterestSchedule, error) {
	var dayOfWeek, dayOfMonth interface{}
	if sched.DayOfWeek != nil {
		dayOfWeek = *sched.DayOfWeek
	}
	if sched.DayOfMonth != nil {
		dayOfMonth = *sched.DayOfMonth
	}
	var nextRunAt interface{}
	if sched.NextRunAt != nil {
		nextRunAt = sched.NextRunAt.UTC().Format(time.RFC3339)
	}

	_, err := s.db.Write.Exec(
		`UPDATE interest_schedules
		 SET frequency = ?, day_of_week = ?, day_of_month = ?,
		     next_run_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		sched.Frequency, dayOfWeek, dayOfMonth, nextRunAt, sched.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update interest schedule: %w", err)
	}

	return s.GetByID(sched.ID)
}

// Delete removes an interest schedule by its ID.
func (s *InterestScheduleStore) Delete(id int64) error {
	_, err := s.db.Write.Exec(`DELETE FROM interest_schedules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete interest schedule: %w", err)
	}
	return nil
}

// ListDue returns all active interest schedules whose next_run_at is at or before the given time.
func (s *InterestScheduleStore) ListDue(now time.Time) ([]InterestSchedule, error) {
	rows, err := s.db.Read.Query(
		`SELECT id, child_id, parent_id, frequency, day_of_week, day_of_month,
		        status, next_run_at, created_at, updated_at
		 FROM interest_schedules
		 WHERE status = 'active' AND next_run_at <= ?
		 ORDER BY next_run_at ASC`,
		now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("list due interest schedules: %w", err)
	}
	defer rows.Close()

	var schedules []InterestSchedule
	for rows.Next() {
		sched, err := s.scanRow(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *sched)
	}

	return schedules, rows.Err()
}

// UpdateNextRunAt sets the next_run_at for a schedule.
func (s *InterestScheduleStore) UpdateNextRunAt(id int64, nextRunAt time.Time) error {
	_, err := s.db.Write.Exec(
		`UPDATE interest_schedules SET next_run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		nextRunAt.UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		return fmt.Errorf("update next_run_at: %w", err)
	}
	return nil
}

// UpdateStatus sets the status of an interest schedule.
func (s *InterestScheduleStore) UpdateStatus(id int64, status ScheduleStatus) error {
	_, err := s.db.Write.Exec(
		`UPDATE interest_schedules SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// scanOne scans a single row from a query.
func (s *InterestScheduleStore) scanOne(query string, args ...interface{}) (*InterestSchedule, error) {
	var sched InterestSchedule
	var dayOfWeek, dayOfMonth sql.NullInt64
	var nextRunAt sql.NullString
	var createdAt, updatedAt string

	err := s.db.Read.QueryRow(query, args...).Scan(
		&sched.ID, &sched.ChildID, &sched.ParentID, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &sched.Status, &nextRunAt,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan interest schedule: %w", err)
	}

	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int64)
		sched.DayOfWeek = &v
	}
	if dayOfMonth.Valid {
		v := int(dayOfMonth.Int64)
		sched.DayOfMonth = &v
	}
	if nextRunAt.Valid {
		t, _ := parseTime(nextRunAt.String)
		sched.NextRunAt = &t
	}
	sched.CreatedAt, _ = parseTime(createdAt)
	sched.UpdatedAt, _ = parseTime(updatedAt)

	return &sched, nil
}

// scanRow scans a single row from rows.
func (s *InterestScheduleStore) scanRow(rows *sql.Rows) (*InterestSchedule, error) {
	var sched InterestSchedule
	var dayOfWeek, dayOfMonth sql.NullInt64
	var nextRunAt sql.NullString
	var createdAt, updatedAt string

	if err := rows.Scan(
		&sched.ID, &sched.ChildID, &sched.ParentID, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &sched.Status, &nextRunAt,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan interest schedule row: %w", err)
	}

	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int64)
		sched.DayOfWeek = &v
	}
	if dayOfMonth.Valid {
		v := int(dayOfMonth.Int64)
		sched.DayOfMonth = &v
	}
	if nextRunAt.Valid {
		t, _ := parseTime(nextRunAt.String)
		sched.NextRunAt = &t
	}
	sched.CreatedAt, _ = parseTime(createdAt)
	sched.UpdatedAt, _ = parseTime(updatedAt)

	return &sched, nil
}
