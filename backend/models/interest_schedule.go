package models

import "time"

// InterestSchedule represents a recurring interest accrual configuration.
type InterestSchedule struct {
	ID         int64          `gorm:"primaryKey" json:"id"`
	ChildID    int64          `gorm:"uniqueIndex;not null" json:"child_id"`
	ParentID   int64          `gorm:"not null" json:"parent_id"`
	Frequency  Frequency      `gorm:"not null" json:"frequency"`
	DayOfWeek  *int           `json:"day_of_week,omitempty"`
	DayOfMonth *int           `json:"day_of_month,omitempty"`
	Status     ScheduleStatus `gorm:"not null;default:active" json:"status"`
	NextRunAt  *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Child  Child  `gorm:"foreignKey:ChildID" json:"-"`
	Parent Parent `gorm:"foreignKey:ParentID" json:"-"`
}
