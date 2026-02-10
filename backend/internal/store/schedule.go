package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Frequency represents how often an allowance schedule executes.
type Frequency string

const (
	FrequencyWeekly   Frequency = "weekly"
	FrequencyBiweekly Frequency = "biweekly"
	FrequencyMonthly  Frequency = "monthly"
)

// ScheduleStatus represents the current state of an allowance schedule.
type ScheduleStatus string

const (
	ScheduleStatusActive ScheduleStatus = "active"
	ScheduleStatusPaused ScheduleStatus = "paused"
)

// AllowanceSchedule represents a recurring deposit configuration.
type AllowanceSchedule struct {
	ID          int64          `json:"id"`
	ChildID     int64          `json:"child_id"`
	ParentID    int64          `json:"parent_id"`
	AmountCents int64          `json:"amount_cents"`
	Frequency   Frequency      `json:"frequency"`
	DayOfWeek   *int           `json:"day_of_week,omitempty"`
	DayOfMonth  *int           `json:"day_of_month,omitempty"`
	Note        *string        `json:"note,omitempty"`
	Status      ScheduleStatus `json:"status"`
	NextRunAt   *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ScheduleWithChild extends AllowanceSchedule with the child's first name for list views.
type ScheduleWithChild struct {
	AllowanceSchedule
	ChildFirstName string `json:"child_first_name"`
}

// UpcomingAllowance represents a child's next scheduled allowance deposit.
type UpcomingAllowance struct {
	AmountCents int64     `json:"amount_cents"`
	NextDate    time.Time `json:"next_date"`
	Note        *string   `json:"note,omitempty"`
}

// ScheduleStore handles database operations for allowance schedules.
type ScheduleStore struct {
	db *DB
}

// NewScheduleStore creates a new ScheduleStore.
func NewScheduleStore(db *DB) *ScheduleStore {
	return &ScheduleStore{db: db}
}

// Create inserts a new allowance schedule into the database.
func (s *ScheduleStore) Create(sched *AllowanceSchedule) (*AllowanceSchedule, error) {
	var dayOfWeek, dayOfMonth interface{}
	if sched.DayOfWeek != nil {
		dayOfWeek = *sched.DayOfWeek
	}
	if sched.DayOfMonth != nil {
		dayOfMonth = *sched.DayOfMonth
	}
	var note interface{}
	if sched.Note != nil {
		note = *sched.Note
	}
	var nextRunAt interface{}
	if sched.NextRunAt != nil {
		nextRunAt = sched.NextRunAt.UTC().Format(time.RFC3339)
	}

	result, err := s.db.Write.Exec(
		`INSERT INTO allowance_schedules
		(child_id, parent_id, amount_cents, frequency, day_of_week, day_of_month, note, status, next_run_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sched.ChildID, sched.ParentID, sched.AmountCents, sched.Frequency,
		dayOfWeek, dayOfMonth, note, sched.Status, nextRunAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert schedule: %w", err)
	}

	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

// GetByID retrieves a schedule by its ID.
func (s *ScheduleStore) GetByID(id int64) (*AllowanceSchedule, error) {
	var sched AllowanceSchedule
	var dayOfWeek, dayOfMonth sql.NullInt64
	var note sql.NullString
	var nextRunAt sql.NullString
	var createdAt, updatedAt string

	err := s.db.Read.QueryRow(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules WHERE id = ?`, id,
	).Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule by id: %w", err)
	}

	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int64)
		sched.DayOfWeek = &v
	}
	if dayOfMonth.Valid {
		v := int(dayOfMonth.Int64)
		sched.DayOfMonth = &v
	}
	if note.Valid {
		sched.Note = &note.String
	}
	if nextRunAt.Valid {
		t, _ := parseTime(nextRunAt.String)
		sched.NextRunAt = &t
	}
	sched.CreatedAt, _ = parseTime(createdAt)
	sched.UpdatedAt, _ = parseTime(updatedAt)

	return &sched, nil
}

// GetByChildID returns the single allowance schedule for a child (any status), or nil if none exists.
func (s *ScheduleStore) GetByChildID(childID int64) (*AllowanceSchedule, error) {
	var sched AllowanceSchedule
	var dayOfWeek, dayOfMonth sql.NullInt64
	var note sql.NullString
	var nextRunAt sql.NullString
	var createdAt, updatedAt string

	err := s.db.Read.QueryRow(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules WHERE child_id = ?`, childID,
	).Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule by child id: %w", err)
	}

	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int64)
		sched.DayOfWeek = &v
	}
	if dayOfMonth.Valid {
		v := int(dayOfMonth.Int64)
		sched.DayOfMonth = &v
	}
	if note.Valid {
		sched.Note = &note.String
	}
	if nextRunAt.Valid {
		t, _ := parseTime(nextRunAt.String)
		sched.NextRunAt = &t
	}
	sched.CreatedAt, _ = parseTime(createdAt)
	sched.UpdatedAt, _ = parseTime(updatedAt)

	return &sched, nil
}

// ListByParentFamily returns all schedules for a family, with child names joined.
func (s *ScheduleStore) ListByParentFamily(familyID int64) ([]ScheduleWithChild, error) {
	rows, err := s.db.Read.Query(
		`SELECT s.id, s.child_id, s.parent_id, s.amount_cents, s.frequency,
		        s.day_of_week, s.day_of_month, s.note, s.status, s.next_run_at,
		        s.created_at, s.updated_at, c.first_name
		 FROM allowance_schedules s
		 JOIN children c ON c.id = s.child_id
		 WHERE c.family_id = ?
		 ORDER BY s.created_at DESC`, familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []ScheduleWithChild
	for rows.Next() {
		var sc ScheduleWithChild
		var dayOfWeek, dayOfMonth sql.NullInt64
		var note sql.NullString
		var nextRunAt sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(&sc.ID, &sc.ChildID, &sc.ParentID, &sc.AmountCents, &sc.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &sc.Status, &nextRunAt,
			&createdAt, &updatedAt, &sc.ChildFirstName); err != nil {
			return nil, fmt.Errorf("scan schedule: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			sc.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			sc.DayOfMonth = &v
		}
		if note.Valid {
			sc.Note = &note.String
		}
		if nextRunAt.Valid {
			t, _ := parseTime(nextRunAt.String)
			sc.NextRunAt = &t
		}
		sc.CreatedAt, _ = parseTime(createdAt)
		sc.UpdatedAt, _ = parseTime(updatedAt)

		schedules = append(schedules, sc)
	}

	return schedules, rows.Err()
}

// Update updates a schedule's amount, frequency, day, and note fields.
func (s *ScheduleStore) Update(sched *AllowanceSchedule) (*AllowanceSchedule, error) {
	var dayOfWeek, dayOfMonth interface{}
	if sched.DayOfWeek != nil {
		dayOfWeek = *sched.DayOfWeek
	}
	if sched.DayOfMonth != nil {
		dayOfMonth = *sched.DayOfMonth
	}
	var note interface{}
	if sched.Note != nil {
		note = *sched.Note
	}
	var nextRunAt interface{}
	if sched.NextRunAt != nil {
		nextRunAt = sched.NextRunAt.UTC().Format(time.RFC3339)
	}

	_, err := s.db.Write.Exec(
		`UPDATE allowance_schedules
		 SET amount_cents = ?, frequency = ?, day_of_week = ?, day_of_month = ?,
		     note = ?, next_run_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		sched.AmountCents, sched.Frequency, dayOfWeek, dayOfMonth,
		note, nextRunAt, sched.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update schedule: %w", err)
	}

	return s.GetByID(sched.ID)
}

// Delete removes a schedule by its ID.
func (s *ScheduleStore) Delete(id int64) error {
	_, err := s.db.Write.Exec(`DELETE FROM allowance_schedules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

// ListDue returns all active schedules whose next_run_at is at or before the given time.
func (s *ScheduleStore) ListDue(now time.Time) ([]AllowanceSchedule, error) {
	rows, err := s.db.Read.Query(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules
		 WHERE status = 'active' AND next_run_at <= ?
		 ORDER BY next_run_at ASC`,
		now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("list due schedules: %w", err)
	}
	defer rows.Close()

	var schedules []AllowanceSchedule
	for rows.Next() {
		var sched AllowanceSchedule
		var dayOfWeek, dayOfMonth sql.NullInt64
		var note sql.NullString
		var nextRunAt sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
			&createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan due schedule: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			sched.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			sched.DayOfMonth = &v
		}
		if note.Valid {
			sched.Note = &note.String
		}
		if nextRunAt.Valid {
			t, _ := parseTime(nextRunAt.String)
			sched.NextRunAt = &t
		}
		sched.CreatedAt, _ = parseTime(createdAt)
		sched.UpdatedAt, _ = parseTime(updatedAt)

		schedules = append(schedules, sched)
	}

	return schedules, rows.Err()
}

// UpdateNextRunAt sets the next_run_at for a schedule.
func (s *ScheduleStore) UpdateNextRunAt(id int64, nextRunAt time.Time) error {
	_, err := s.db.Write.Exec(
		`UPDATE allowance_schedules SET next_run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		nextRunAt.UTC().Format(time.RFC3339), id,
	)
	if err != nil {
		return fmt.Errorf("update next_run_at: %w", err)
	}
	return nil
}

// UpdateStatus sets the status of a schedule.
func (s *ScheduleStore) UpdateStatus(id int64, status ScheduleStatus) error {
	_, err := s.db.Write.Exec(
		`UPDATE allowance_schedules SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// ListActiveByChild returns all active schedules for a child, sorted by next_run_at.
func (s *ScheduleStore) ListActiveByChild(childID int64) ([]AllowanceSchedule, error) {
	rows, err := s.db.Read.Query(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules
		 WHERE child_id = ? AND status = 'active'
		 ORDER BY next_run_at ASC`, childID,
	)
	if err != nil {
		return nil, fmt.Errorf("list active schedules by child: %w", err)
	}
	defer rows.Close()

	var schedules []AllowanceSchedule
	for rows.Next() {
		var sched AllowanceSchedule
		var dayOfWeek, dayOfMonth sql.NullInt64
		var note sql.NullString
		var nextRunAt sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
			&createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan active schedule: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			sched.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			sched.DayOfMonth = &v
		}
		if note.Valid {
			sched.Note = &note.String
		}
		if nextRunAt.Valid {
			t, _ := parseTime(nextRunAt.String)
			sched.NextRunAt = &t
		}
		sched.CreatedAt, _ = parseTime(createdAt)
		sched.UpdatedAt, _ = parseTime(updatedAt)

		schedules = append(schedules, sched)
	}

	return schedules, rows.Err()
}
