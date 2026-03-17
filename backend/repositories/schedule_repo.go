package repositories

import (
	"errors"
	"fmt"
	"time"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// ScheduleWithChild extends AllowanceSchedule with the child's first name for list views.
type ScheduleWithChild struct {
	models.AllowanceSchedule
	ChildFirstName string `json:"child_first_name"`
}

// DueAllowanceSchedule extends AllowanceSchedule with the family's timezone for timezone-aware scheduling.
type DueAllowanceSchedule struct {
	models.AllowanceSchedule
	FamilyTimezone string `json:"family_timezone"`
}

// ScheduleRepo handles database operations for allowance schedules using GORM.
type ScheduleRepo struct {
	db *gorm.DB
}

// NewScheduleRepo creates a new ScheduleRepo.
func NewScheduleRepo(db *gorm.DB) *ScheduleRepo {
	return &ScheduleRepo{db: db}
}

// Create inserts a new allowance schedule into the database and returns the created entity.
func (r *ScheduleRepo) Create(sched *models.AllowanceSchedule) (*models.AllowanceSchedule, error) {
	if err := r.db.Create(sched).Error; err != nil {
		return nil, fmt.Errorf("insert schedule: %w", err)
	}
	return r.GetByID(sched.ID)
}

// GetByID retrieves a schedule by its ID. Returns (nil, nil) if not found.
func (r *ScheduleRepo) GetByID(id int64) (*models.AllowanceSchedule, error) {
	var sched models.AllowanceSchedule
	err := r.db.First(&sched, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule by id: %w", err)
	}
	return &sched, nil
}

// GetByChildID returns the single allowance schedule for a child (any status), or nil if none exists.
func (r *ScheduleRepo) GetByChildID(childID int64) (*models.AllowanceSchedule, error) {
	var sched models.AllowanceSchedule
	err := r.db.Where("child_id = ?", childID).First(&sched).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule by child id: %w", err)
	}
	return &sched, nil
}

// ListByParentFamily returns all schedules for a family, with child names joined.
func (r *ScheduleRepo) ListByParentFamily(familyID int64) ([]ScheduleWithChild, error) {
	var results []ScheduleWithChild
	err := r.db.
		Table("allowance_schedules s").
		Select("s.*, c.first_name as child_first_name").
		Joins("JOIN children c ON c.id = s.child_id").
		Where("c.family_id = ?", familyID).
		Order("s.created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	return results, nil
}

// Update updates a schedule's amount, frequency, day, note, and next_run_at fields.
func (r *ScheduleRepo) Update(sched *models.AllowanceSchedule) (*models.AllowanceSchedule, error) {
	err := r.db.Model(&models.AllowanceSchedule{}).
		Where("id = ?", sched.ID).
		Updates(map[string]interface{}{
			"amount_cents": sched.AmountCents,
			"frequency":    sched.Frequency,
			"day_of_week":  sched.DayOfWeek,
			"day_of_month": sched.DayOfMonth,
			"note":         sched.Note,
			"next_run_at":  sched.NextRunAt,
			"updated_at":   gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return nil, fmt.Errorf("update schedule: %w", err)
	}
	return r.GetByID(sched.ID)
}

// Delete removes a schedule by its ID.
func (r *ScheduleRepo) Delete(id int64) error {
	if err := r.db.Delete(&models.AllowanceSchedule{}, id).Error; err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

// ListDue returns all active schedules whose next_run_at is at or before the given time,
// including the family's timezone for timezone-aware next-run calculation.
func (r *ScheduleRepo) ListDue(now time.Time) ([]DueAllowanceSchedule, error) {
	var results []DueAllowanceSchedule
	err := r.db.
		Table("allowance_schedules s").
		Select("s.*, COALESCE(f.timezone, '') as family_timezone").
		Joins("JOIN children c ON c.id = s.child_id").
		Joins("JOIN families f ON f.id = c.family_id").
		Where("s.status = ? AND s.next_run_at <= ? AND c.is_disabled = ?", "active", now, false).
		Order("s.next_run_at ASC").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list due schedules: %w", err)
	}
	return results, nil
}

// ListAllActiveWithTimezone returns all active schedules with their family's timezone.
// Used for startup recalculation of next_run_at values.
func (r *ScheduleRepo) ListAllActiveWithTimezone() ([]DueAllowanceSchedule, error) {
	var results []DueAllowanceSchedule
	err := r.db.
		Table("allowance_schedules s").
		Select("s.*, COALESCE(f.timezone, '') as family_timezone").
		Joins("JOIN children c ON c.id = s.child_id").
		Joins("JOIN families f ON f.id = c.family_id").
		Where("s.status = ? AND c.is_disabled = ?", "active", false).
		Order("s.id ASC").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list all active schedules with timezone: %w", err)
	}
	return results, nil
}

// UpdateNextRunAt sets the next_run_at for a schedule.
func (r *ScheduleRepo) UpdateNextRunAt(id int64, nextRunAt time.Time) error {
	err := r.db.Model(&models.AllowanceSchedule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"next_run_at": nextRunAt,
			"updated_at":  gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return fmt.Errorf("update next_run_at: %w", err)
	}
	return nil
}

// UpdateStatus sets the status of a schedule.
func (r *ScheduleRepo) UpdateStatus(id int64, status models.ScheduleStatus) error {
	err := r.db.Model(&models.AllowanceSchedule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// ListActiveByChild returns all active schedules for a child, sorted by next_run_at.
func (r *ScheduleRepo) ListActiveByChild(childID int64) ([]models.AllowanceSchedule, error) {
	var schedules []models.AllowanceSchedule
	err := r.db.
		Where("child_id = ? AND status = ?", childID, "active").
		Order("next_run_at ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, fmt.Errorf("list active schedules by child: %w", err)
	}
	return schedules, nil
}
