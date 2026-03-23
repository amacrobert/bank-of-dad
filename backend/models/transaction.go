package models

import (
	"errors"
	"time"
)

// TransactionType represents the type of a balance transaction.
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeAllowance  TransactionType = "allowance"
	TransactionTypeInterest   TransactionType = "interest"
	TransactionTypeChore      TransactionType = "chore"
)

// ErrInsufficientFunds is returned when a withdrawal exceeds the available balance.
var ErrInsufficientFunds = errors.New("insufficient funds")

// Transaction represents a record of money added or removed from a child's account.
type Transaction struct {
	ID              int64           `gorm:"primaryKey" json:"id"`
	ChildID         int64           `gorm:"not null" json:"child_id"`
	ParentID        int64           `gorm:"not null" json:"parent_id"`
	AmountCents     int64           `gorm:"not null" json:"amount_cents"`
	TransactionType TransactionType `gorm:"column:transaction_type;not null" json:"type"`
	Note            *string         `json:"note,omitempty"`
	ScheduleID      *int64          `json:"schedule_id,omitempty"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"created_at"`

	// Associations
	Child    Child             `gorm:"foreignKey:ChildID" json:"-"`
	Parent   Parent            `gorm:"foreignKey:ParentID" json:"-"`
	Schedule *AllowanceSchedule `gorm:"foreignKey:ScheduleID" json:"-"`
}
