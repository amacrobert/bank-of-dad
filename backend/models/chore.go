package models

import "time"

// ChoreRecurrence represents how often a chore repeats.
type ChoreRecurrence string

const (
	ChoreRecurrenceOneTime ChoreRecurrence = "one_time"
	ChoreRecurrenceDaily   ChoreRecurrence = "daily"
	ChoreRecurrenceWeekly  ChoreRecurrence = "weekly"
	ChoreRecurrenceMonthly ChoreRecurrence = "monthly"
)

// ChoreInstanceStatus represents the current state of a chore instance.
type ChoreInstanceStatus string

const (
	ChoreInstanceStatusAvailable       ChoreInstanceStatus = "available"
	ChoreInstanceStatusPendingApproval ChoreInstanceStatus = "pending_approval"
	ChoreInstanceStatusApproved        ChoreInstanceStatus = "approved"
	ChoreInstanceStatusExpired         ChoreInstanceStatus = "expired"
)

// Chore is a task template defined by a parent, belonging to a family.
type Chore struct {
	ID                int64           `gorm:"primaryKey" json:"id"`
	FamilyID          int64           `gorm:"not null" json:"family_id"`
	CreatedByParentID int64           `gorm:"not null" json:"created_by_parent_id"`
	Name              string          `gorm:"not null" json:"name"`
	Description       *string         `json:"description,omitempty"`
	RewardCents       int             `gorm:"not null;default:0" json:"reward_cents"`
	Recurrence        ChoreRecurrence `gorm:"not null;default:one_time" json:"recurrence"`
	DayOfWeek         *int            `json:"day_of_week,omitempty"`
	DayOfMonth        *int            `json:"day_of_month,omitempty"`
	IsActive          bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt         time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Family      Family            `gorm:"foreignKey:FamilyID" json:"-"`
	Parent      Parent            `gorm:"foreignKey:CreatedByParentID" json:"-"`
	Assignments []ChoreAssignment `gorm:"foreignKey:ChoreID" json:"-"`
	Instances   []ChoreInstance   `gorm:"foreignKey:ChoreID" json:"-"`
}

// ChoreAssignment links a chore to a child.
type ChoreAssignment struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ChoreID   int64     `gorm:"not null" json:"chore_id"`
	ChildID   int64     `gorm:"not null" json:"child_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Associations
	Chore Chore `gorm:"foreignKey:ChoreID" json:"-"`
	Child Child `gorm:"foreignKey:ChildID" json:"-"`
}

// ChoreInstance is a specific occurrence of a chore for a specific child.
type ChoreInstance struct {
	ID                  int64               `gorm:"primaryKey" json:"id"`
	ChoreID             int64               `gorm:"not null" json:"chore_id"`
	ChildID             int64               `gorm:"not null" json:"child_id"`
	RewardCents         int                 `gorm:"not null" json:"reward_cents"`
	Status              ChoreInstanceStatus `gorm:"not null;default:available" json:"status"`
	PeriodStart         *time.Time          `gorm:"type:date" json:"period_start,omitempty"`
	PeriodEnd           *time.Time          `gorm:"type:date" json:"period_end,omitempty"`
	CompletedAt         *time.Time          `json:"completed_at,omitempty"`
	ReviewedAt          *time.Time          `json:"reviewed_at,omitempty"`
	ReviewedByParentID  *int64              `json:"reviewed_by_parent_id,omitempty"`
	RejectionReason     *string             `json:"rejection_reason,omitempty"`
	TransactionID       *int64              `json:"transaction_id,omitempty"`
	CreatedAt           time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time           `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Chore       Chore        `gorm:"foreignKey:ChoreID" json:"-"`
	Child       Child        `gorm:"foreignKey:ChildID" json:"-"`
	Transaction *Transaction `gorm:"foreignKey:TransactionID" json:"-"`
}
