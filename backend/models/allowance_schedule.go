package models

import "time"

// Frequency represents how often a schedule executes.
type Frequency string

const (
	FrequencyWeekly   Frequency = "weekly"
	FrequencyBiweekly Frequency = "biweekly"
	FrequencyMonthly  Frequency = "monthly"
)

// ScheduleStatus represents the current state of a schedule.
type ScheduleStatus string

const (
	ScheduleStatusActive ScheduleStatus = "active"
	ScheduleStatusPaused ScheduleStatus = "paused"
)

// AllowanceSchedule represents a recurring deposit configuration.
type AllowanceSchedule struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	ChildID     int64          `gorm:"not null" json:"child_id"`
	ParentID    int64          `gorm:"not null" json:"parent_id"`
	AmountCents int64          `gorm:"not null" json:"amount_cents"`
	Frequency   Frequency      `gorm:"not null" json:"frequency"`
	DayOfWeek   *int           `json:"day_of_week,omitempty"`
	DayOfMonth  *int           `json:"day_of_month,omitempty"`
	Note        *string        `json:"note,omitempty"`
	Status      ScheduleStatus `gorm:"not null;default:active" json:"status"`
	NextRunAt   *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Child        Child         `gorm:"foreignKey:ChildID" json:"-"`
	Parent       Parent        `gorm:"foreignKey:ParentID" json:"-"`
	Transactions []Transaction `gorm:"foreignKey:ScheduleID" json:"-"`
}
