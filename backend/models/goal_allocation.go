package models

import "time"

// GoalAllocation represents an audit trail entry for goal fund movements.
type GoalAllocation struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	GoalID      int64     `gorm:"not null" json:"goal_id"`
	ChildID     int64     `gorm:"not null" json:"child_id"`
	AmountCents int64     `gorm:"not null" json:"amount_cents"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Associations
	SavingsGoal SavingsGoal `gorm:"foreignKey:GoalID" json:"-"`
	Child       Child       `gorm:"foreignKey:ChildID" json:"-"`
}
