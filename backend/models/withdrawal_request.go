package models

import "time"

// WithdrawalRequestStatus represents the current state of a withdrawal request.
type WithdrawalRequestStatus string

const (
	WithdrawalRequestStatusPending   WithdrawalRequestStatus = "pending"
	WithdrawalRequestStatusApproved  WithdrawalRequestStatus = "approved"
	WithdrawalRequestStatusDenied    WithdrawalRequestStatus = "denied"
	WithdrawalRequestStatusCancelled WithdrawalRequestStatus = "cancelled"
)

// WithdrawalRequest represents a child's request to withdraw funds, subject to parent approval.
type WithdrawalRequest struct {
	ID                 int64                   `gorm:"primaryKey" json:"id"`
	ChildID            int64                   `gorm:"not null" json:"child_id"`
	FamilyID           int64                   `gorm:"not null" json:"family_id"`
	AmountCents        int                     `gorm:"not null" json:"amount_cents"`
	Reason             string                  `gorm:"not null;size:500" json:"reason"`
	Status             WithdrawalRequestStatus `gorm:"not null;default:pending" json:"status"`
	DenialReason       *string                 `gorm:"size:500" json:"denial_reason,omitempty"`
	ReviewedByParentID *int64                  `json:"reviewed_by_parent_id,omitempty"`
	ReviewedAt         *time.Time              `json:"reviewed_at,omitempty"`
	TransactionID      *int64                  `json:"transaction_id,omitempty"`
	CreatedAt          time.Time               `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time               `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	Child       Child        `gorm:"foreignKey:ChildID" json:"-"`
	Family      Family       `gorm:"foreignKey:FamilyID" json:"-"`
	Transaction *Transaction `gorm:"foreignKey:TransactionID" json:"-"`
}
