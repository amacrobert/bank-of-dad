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

// DueInterestSchedule extends InterestSchedule with the family's timezone for timezone-aware scheduling.
type DueInterestSchedule struct {
	InterestSchedule
	FamilyTimezone string `json:"family_timezone"`
}

// InterestScheduleStore handles database operations for interest accrual schedules.
type InterestScheduleStore struct {
	db *sql.DB
}

// NewInterestScheduleStore creates a new InterestScheduleStore.
func NewInterestScheduleStore(db *sql.DB) *InterestScheduleStore {
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

	var id int64
	err := s.db.QueryRow(
		`INSERT INTO interest_schedules
		(child_id, parent_id, frequency, day_of_week, day_of_month, status, next_run_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		sched.ChildID, sched.ParentID, sched.Frequency,
		dayOfWeek, dayOfMonth, sched.Status, sched.NextRunAt,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert interest schedule: %w", err)
	}

	return s.GetByID(id)
}

// GetByID retrieves an interest schedule by its ID.
func (s *InterestScheduleStore) GetByID(id int64) (*InterestSchedule, error) {
	return s.scanOne(
		`SELECT id, child_id, parent_id, frequency, day_of_week, day_of_month,
		        status, next_run_at, created_at, updated_at
		 FROM interest_schedules WHERE id = $1`, id,
	)
}

// GetByChildID retrieves the interest schedule for a child (any status), or nil if none exists.
func (s *InterestScheduleStore) GetByChildID(childID int64) (*InterestSchedule, error) {
	return s.scanOne(
		`SELECT id, child_id, parent_id, frequency, day_of_week, day_of_month,
		        status, next_run_at, created_at, updated_at
		 FROM interest_schedules WHERE child_id = $1`, childID,
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

	_, err := s.db.Exec(
		`UPDATE interest_schedules
		 SET frequency = $1, day_of_week = $2, day_of_month = $3,
		     next_run_at = $4, updated_at = NOW()
		 WHERE id = $5`,
		sched.Frequency, dayOfWeek, dayOfMonth, sched.NextRunAt, sched.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update interest schedule: %w", err)
	}

	return s.GetByID(sched.ID)
}

// Delete removes an interest schedule by its ID.
func (s *InterestScheduleStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM interest_schedules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete interest schedule: %w", err)
	}
	return nil
}

// ListDue returns all active interest schedules whose next_run_at is at or before the given time,
// including the family's timezone for timezone-aware next-run calculation.
func (s *InterestScheduleStore) ListDue(now time.Time) ([]DueInterestSchedule, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.child_id, s.parent_id, s.frequency, s.day_of_week, s.day_of_month,
		        s.status, s.next_run_at, s.created_at, s.updated_at, COALESCE(f.timezone, '')
		 FROM interest_schedules s
		 JOIN children c ON c.id = s.child_id
		 JOIN families f ON f.id = c.family_id
		 WHERE s.status = 'active' AND s.next_run_at <= $1
		 ORDER BY s.next_run_at ASC`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list due interest schedules: %w", err)
	}
	defer rows.Close()

	var schedules []DueInterestSchedule
	for rows.Next() {
		var ds DueInterestSchedule
		var dayOfWeek, dayOfMonth sql.NullInt64
		var nextRunAt sql.NullTime

		if err := rows.Scan(&ds.ID, &ds.ChildID, &ds.ParentID, &ds.Frequency,
			&dayOfWeek, &dayOfMonth, &ds.Status, &nextRunAt,
			&ds.CreatedAt, &ds.UpdatedAt, &ds.FamilyTimezone); err != nil {
			return nil, fmt.Errorf("scan due interest schedule: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			ds.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			ds.DayOfMonth = &v
		}
		if nextRunAt.Valid {
			ds.NextRunAt = &nextRunAt.Time
		}

		schedules = append(schedules, ds)
	}

	return schedules, rows.Err()
}

// ListAllActiveWithTimezone returns all active interest schedules with their family's timezone.
// Used for startup recalculation of next_run_at values.
func (s *InterestScheduleStore) ListAllActiveWithTimezone() ([]DueInterestSchedule, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.child_id, s.parent_id, s.frequency, s.day_of_week, s.day_of_month,
		        s.status, s.next_run_at, s.created_at, s.updated_at, COALESCE(f.timezone, '')
		 FROM interest_schedules s
		 JOIN children c ON c.id = s.child_id
		 JOIN families f ON f.id = c.family_id
		 WHERE s.status = 'active'
		 ORDER BY s.id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all active interest schedules with timezone: %w", err)
	}
	defer rows.Close()

	var schedules []DueInterestSchedule
	for rows.Next() {
		var ds DueInterestSchedule
		var dayOfWeek, dayOfMonth sql.NullInt64
		var nextRunAt sql.NullTime

		if err := rows.Scan(&ds.ID, &ds.ChildID, &ds.ParentID, &ds.Frequency,
			&dayOfWeek, &dayOfMonth, &ds.Status, &nextRunAt,
			&ds.CreatedAt, &ds.UpdatedAt, &ds.FamilyTimezone); err != nil {
			return nil, fmt.Errorf("scan active interest schedule with timezone: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			ds.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			ds.DayOfMonth = &v
		}
		if nextRunAt.Valid {
			ds.NextRunAt = &nextRunAt.Time
		}

		schedules = append(schedules, ds)
	}

	return schedules, rows.Err()
}

// UpdateNextRunAt sets the next_run_at for a schedule.
func (s *InterestScheduleStore) UpdateNextRunAt(id int64, nextRunAt time.Time) error {
	_, err := s.db.Exec(
		`UPDATE interest_schedules SET next_run_at = $1, updated_at = NOW() WHERE id = $2`,
		nextRunAt, id,
	)
	if err != nil {
		return fmt.Errorf("update next_run_at: %w", err)
	}
	return nil
}

// UpdateStatus sets the status of an interest schedule.
func (s *InterestScheduleStore) UpdateStatus(id int64, status ScheduleStatus) error {
	_, err := s.db.Exec(
		`UPDATE interest_schedules SET status = $1, updated_at = NOW() WHERE id = $2`,
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
	var nextRunAt sql.NullTime

	err := s.db.QueryRow(query, args...).Scan(
		&sched.ID, &sched.ChildID, &sched.ParentID, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &sched.Status, &nextRunAt,
		&sched.CreatedAt, &sched.UpdatedAt,
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
		sched.NextRunAt = &nextRunAt.Time
	}

	return &sched, nil
}
