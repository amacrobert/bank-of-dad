package repositories

import (
	"errors"
	"fmt"
	"time"

	"bank-of-dad/models"

	"gorm.io/gorm"
)

// DueInterestSchedule extends InterestSchedule with the family's timezone for timezone-aware scheduling.
type DueInterestSchedule struct {
	models.InterestSchedule
	FamilyTimezone string `json:"family_timezone"`
}

// InterestScheduleRepo handles database operations for interest accrual schedules using GORM.
type InterestScheduleRepo struct {
	db *gorm.DB
}

// NewInterestScheduleRepo creates a new InterestScheduleRepo.
func NewInterestScheduleRepo(db *gorm.DB) *InterestScheduleRepo {
	return &InterestScheduleRepo{db: db}
}

// Create inserts a new interest schedule and returns the created entity.
func (r *InterestScheduleRepo) Create(sched *models.InterestSchedule) (*models.InterestSchedule, error) {
	if err := r.db.Create(sched).Error; err != nil {
		return nil, fmt.Errorf("insert interest schedule: %w", err)
	}
	return r.GetByID(sched.ID)
}

// GetByID retrieves an interest schedule by its ID. Returns (nil, nil) if not found.
func (r *InterestScheduleRepo) GetByID(id int64) (*models.InterestSchedule, error) {
	var sched models.InterestSchedule
	err := r.db.First(&sched, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get interest schedule by id: %w", err)
	}
	return &sched, nil
}

// GetByChildID retrieves the interest schedule for a child (any status), or nil if none exists.
func (r *InterestScheduleRepo) GetByChildID(childID int64) (*models.InterestSchedule, error) {
	var sched models.InterestSchedule
	err := r.db.Where("child_id = ?", childID).First(&sched).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get interest schedule by child id: %w", err)
	}
	return &sched, nil
}

// Update updates a schedule's frequency, day, and next_run_at fields.
func (r *InterestScheduleRepo) Update(sched *models.InterestSchedule) (*models.InterestSchedule, error) {
	err := r.db.Model(&models.InterestSchedule{}).
		Where("id = ?", sched.ID).
		Updates(map[string]interface{}{
			"frequency":    sched.Frequency,
			"day_of_week":  sched.DayOfWeek,
			"day_of_month": sched.DayOfMonth,
			"next_run_at":  sched.NextRunAt,
			"updated_at":   gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return nil, fmt.Errorf("update interest schedule: %w", err)
	}
	return r.GetByID(sched.ID)
}

// Delete removes an interest schedule by its ID.
func (r *InterestScheduleRepo) Delete(id int64) error {
	if err := r.db.Delete(&models.InterestSchedule{}, id).Error; err != nil {
		return fmt.Errorf("delete interest schedule: %w", err)
	}
	return nil
}

// ListDue returns all active interest schedules whose next_run_at is at or before the given time,
// including the family's timezone for timezone-aware next-run calculation.
func (r *InterestScheduleRepo) ListDue(now time.Time) ([]DueInterestSchedule, error) {
	var results []DueInterestSchedule
	err := r.db.
		Table("interest_schedules s").
		Select("s.*, COALESCE(f.timezone, '') as family_timezone").
		Joins("JOIN children c ON c.id = s.child_id").
		Joins("JOIN families f ON f.id = c.family_id").
		Where("s.status = ? AND s.next_run_at <= ? AND c.is_disabled = ?", "active", now, false).
		Order("s.next_run_at ASC").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list due interest schedules: %w", err)
	}
	return results, nil
}

// ListAllActiveWithTimezone returns all active interest schedules with their family's timezone.
// Used for startup recalculation of next_run_at values.
func (r *InterestScheduleRepo) ListAllActiveWithTimezone() ([]DueInterestSchedule, error) {
	var results []DueInterestSchedule
	err := r.db.
		Table("interest_schedules s").
		Select("s.*, COALESCE(f.timezone, '') as family_timezone").
		Joins("JOIN children c ON c.id = s.child_id").
		Joins("JOIN families f ON f.id = c.family_id").
		Where("s.status = ? AND c.is_disabled = ?", "active", false).
		Order("s.id ASC").
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("list all active interest schedules with timezone: %w", err)
	}
	return results, nil
}

// UpdateNextRunAt sets the next_run_at for an interest schedule.
func (r *InterestScheduleRepo) UpdateNextRunAt(id int64, nextRunAt time.Time) error {
	err := r.db.Model(&models.InterestSchedule{}).
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

// UpdateStatus sets the status of an interest schedule.
func (r *InterestScheduleRepo) UpdateStatus(id int64, status models.ScheduleStatus) error {
	err := r.db.Model(&models.InterestSchedule{}).
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
