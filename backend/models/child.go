package models

import "time"

// Child represents a child account within a family.
type Child struct {
	ID                  int64      `gorm:"primaryKey" json:"id"`
	FamilyID            int64      `gorm:"not null;uniqueIndex:idx_family_child" json:"family_id"`
	FirstName           string     `gorm:"not null;uniqueIndex:idx_family_child" json:"first_name"`
	PasswordHash        string     `gorm:"not null" json:"-"`
	IsLocked            bool       `gorm:"not null;default:false" json:"is_locked"`
	IsDisabled          bool       `gorm:"not null;default:false" json:"is_disabled"`
	FailedLoginAttempts int        `gorm:"not null;default:0" json:"-"`
	BalanceCents        int64      `gorm:"not null;default:0" json:"balance_cents"`
	InterestRateBps     int        `gorm:"not null;default:0" json:"interest_rate_bps"`
	LastInterestAt      *time.Time `json:"last_interest_at,omitempty"`
	Avatar              *string    `json:"avatar,omitempty"`
	Theme               *string    `json:"theme,omitempty"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Family            Family              `gorm:"foreignKey:FamilyID" json:"-"`
	Transactions      []Transaction       `gorm:"foreignKey:ChildID" json:"-"`
	AllowanceSchedule []AllowanceSchedule `gorm:"foreignKey:ChildID" json:"-"`
	SavingsGoals      []SavingsGoal       `gorm:"foreignKey:ChildID" json:"-"`
	GoalAllocations   []GoalAllocation    `gorm:"foreignKey:ChildID" json:"-"`
}
