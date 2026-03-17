package models

import "time"

// SavingsGoal represents a child's savings target.
type SavingsGoal struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	ChildID     int64      `gorm:"not null" json:"child_id"`
	Name        string     `gorm:"not null" json:"name"`
	TargetCents int64      `gorm:"not null" json:"target_cents"`
	SavedCents  int64      `gorm:"not null;default:0" json:"saved_cents"`
	Emoji       *string    `json:"emoji,omitempty"`
	Status      string     `gorm:"not null;default:active" json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Child           Child            `gorm:"foreignKey:ChildID" json:"-"`
	GoalAllocations []GoalAllocation `gorm:"foreignKey:GoalID" json:"-"`
}
