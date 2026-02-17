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

// DueAllowanceSchedule extends AllowanceSchedule with the family's timezone for timezone-aware scheduling.
type DueAllowanceSchedule struct {
	AllowanceSchedule
	FamilyTimezone string `json:"family_timezone"`
}

// UpcomingAllowance represents a child's next scheduled allowance deposit.
type UpcomingAllowance struct {
	AmountCents int64     `json:"amount_cents"`
	NextDate    time.Time `json:"next_date"`
	Note        *string   `json:"note,omitempty"`
}

// ScheduleStore handles database operations for allowance schedules.
type ScheduleStore struct {
	db *sql.DB
}

// NewScheduleStore creates a new ScheduleStore.
func NewScheduleStore(db *sql.DB) *ScheduleStore {
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

	var id int64
	err := s.db.QueryRow(
		`INSERT INTO allowance_schedules
		(child_id, parent_id, amount_cents, frequency, day_of_week, day_of_month, note, status, next_run_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		sched.ChildID, sched.ParentID, sched.AmountCents, sched.Frequency,
		dayOfWeek, dayOfMonth, note, sched.Status, sched.NextRunAt,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert schedule: %w", err)
	}

	return s.GetByID(id)
}

// GetByID retrieves a schedule by its ID.
func (s *ScheduleStore) GetByID(id int64) (*AllowanceSchedule, error) {
	var sched AllowanceSchedule
	var dayOfWeek, dayOfMonth sql.NullInt64
	var note sql.NullString
	var nextRunAt sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules WHERE id = $1`, id,
	).Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
		&sched.CreatedAt, &sched.UpdatedAt)

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
		sched.NextRunAt = &nextRunAt.Time
	}

	return &sched, nil
}

// GetByChildID returns the single allowance schedule for a child (any status), or nil if none exists.
func (s *ScheduleStore) GetByChildID(childID int64) (*AllowanceSchedule, error) {
	var sched AllowanceSchedule
	var dayOfWeek, dayOfMonth sql.NullInt64
	var note sql.NullString
	var nextRunAt sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules WHERE child_id = $1`, childID,
	).Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
		&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
		&sched.CreatedAt, &sched.UpdatedAt)

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
		sched.NextRunAt = &nextRunAt.Time
	}

	return &sched, nil
}

// ListByParentFamily returns all schedules for a family, with child names joined.
func (s *ScheduleStore) ListByParentFamily(familyID int64) ([]ScheduleWithChild, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.child_id, s.parent_id, s.amount_cents, s.frequency,
		        s.day_of_week, s.day_of_month, s.note, s.status, s.next_run_at,
		        s.created_at, s.updated_at, c.first_name
		 FROM allowance_schedules s
		 JOIN children c ON c.id = s.child_id
		 WHERE c.family_id = $1
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
		var nextRunAt sql.NullTime

		if err := rows.Scan(&sc.ID, &sc.ChildID, &sc.ParentID, &sc.AmountCents, &sc.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &sc.Status, &nextRunAt,
			&sc.CreatedAt, &sc.UpdatedAt, &sc.ChildFirstName); err != nil {
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
			sc.NextRunAt = &nextRunAt.Time
		}

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

	_, err := s.db.Exec(
		`UPDATE allowance_schedules
		 SET amount_cents = $1, frequency = $2, day_of_week = $3, day_of_month = $4,
		     note = $5, next_run_at = $6, updated_at = NOW()
		 WHERE id = $7`,
		sched.AmountCents, sched.Frequency, dayOfWeek, dayOfMonth,
		note, sched.NextRunAt, sched.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update schedule: %w", err)
	}

	return s.GetByID(sched.ID)
}

// Delete removes a schedule by its ID.
func (s *ScheduleStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM allowance_schedules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

// ListDue returns all active schedules whose next_run_at is at or before the given time,
// including the family's timezone for timezone-aware next-run calculation.
func (s *ScheduleStore) ListDue(now time.Time) ([]DueAllowanceSchedule, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.child_id, s.parent_id, s.amount_cents, s.frequency,
		        s.day_of_week, s.day_of_month, s.note, s.status, s.next_run_at,
		        s.created_at, s.updated_at, COALESCE(f.timezone, '')
		 FROM allowance_schedules s
		 JOIN children c ON c.id = s.child_id
		 JOIN families f ON f.id = c.family_id
		 WHERE s.status = 'active' AND s.next_run_at <= $1
		 ORDER BY s.next_run_at ASC`,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("list due schedules: %w", err)
	}
	defer rows.Close()

	var schedules []DueAllowanceSchedule
	for rows.Next() {
		var ds DueAllowanceSchedule
		var dayOfWeek, dayOfMonth sql.NullInt64
		var note sql.NullString
		var nextRunAt sql.NullTime

		if err := rows.Scan(&ds.ID, &ds.ChildID, &ds.ParentID, &ds.AmountCents, &ds.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &ds.Status, &nextRunAt,
			&ds.CreatedAt, &ds.UpdatedAt, &ds.FamilyTimezone); err != nil {
			return nil, fmt.Errorf("scan due schedule: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			ds.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			ds.DayOfMonth = &v
		}
		if note.Valid {
			ds.Note = &note.String
		}
		if nextRunAt.Valid {
			ds.NextRunAt = &nextRunAt.Time
		}

		schedules = append(schedules, ds)
	}

	return schedules, rows.Err()
}

// ListAllActiveWithTimezone returns all active schedules with their family's timezone.
// Used for startup recalculation of next_run_at values.
func (s *ScheduleStore) ListAllActiveWithTimezone() ([]DueAllowanceSchedule, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.child_id, s.parent_id, s.amount_cents, s.frequency,
		        s.day_of_week, s.day_of_month, s.note, s.status, s.next_run_at,
		        s.created_at, s.updated_at, COALESCE(f.timezone, '')
		 FROM allowance_schedules s
		 JOIN children c ON c.id = s.child_id
		 JOIN families f ON f.id = c.family_id
		 WHERE s.status = 'active'
		 ORDER BY s.id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list all active schedules with timezone: %w", err)
	}
	defer rows.Close()

	var schedules []DueAllowanceSchedule
	for rows.Next() {
		var ds DueAllowanceSchedule
		var dayOfWeek, dayOfMonth sql.NullInt64
		var note sql.NullString
		var nextRunAt sql.NullTime

		if err := rows.Scan(&ds.ID, &ds.ChildID, &ds.ParentID, &ds.AmountCents, &ds.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &ds.Status, &nextRunAt,
			&ds.CreatedAt, &ds.UpdatedAt, &ds.FamilyTimezone); err != nil {
			return nil, fmt.Errorf("scan active schedule with timezone: %w", err)
		}

		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			ds.DayOfWeek = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			ds.DayOfMonth = &v
		}
		if note.Valid {
			ds.Note = &note.String
		}
		if nextRunAt.Valid {
			ds.NextRunAt = &nextRunAt.Time
		}

		schedules = append(schedules, ds)
	}

	return schedules, rows.Err()
}

// UpdateNextRunAt sets the next_run_at for a schedule.
func (s *ScheduleStore) UpdateNextRunAt(id int64, nextRunAt time.Time) error {
	_, err := s.db.Exec(
		`UPDATE allowance_schedules SET next_run_at = $1, updated_at = NOW() WHERE id = $2`,
		nextRunAt, id,
	)
	if err != nil {
		return fmt.Errorf("update next_run_at: %w", err)
	}
	return nil
}

// UpdateStatus sets the status of a schedule.
func (s *ScheduleStore) UpdateStatus(id int64, status ScheduleStatus) error {
	_, err := s.db.Exec(
		`UPDATE allowance_schedules SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// ListActiveByChild returns all active schedules for a child, sorted by next_run_at.
func (s *ScheduleStore) ListActiveByChild(childID int64) ([]AllowanceSchedule, error) {
	rows, err := s.db.Query(
		`SELECT id, child_id, parent_id, amount_cents, frequency,
		        day_of_week, day_of_month, note, status, next_run_at,
		        created_at, updated_at
		 FROM allowance_schedules
		 WHERE child_id = $1 AND status = 'active'
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
		var nextRunAt sql.NullTime

		if err := rows.Scan(&sched.ID, &sched.ChildID, &sched.ParentID, &sched.AmountCents, &sched.Frequency,
			&dayOfWeek, &dayOfMonth, &note, &sched.Status, &nextRunAt,
			&sched.CreatedAt, &sched.UpdatedAt); err != nil {
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
			sched.NextRunAt = &nextRunAt.Time
		}

		schedules = append(schedules, sched)
	}

	return schedules, rows.Err()
}
