package models

import (
	"fmt"
	"regexp"
	"time"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// ValidateSlug checks that a slug meets length and character requirements.
func ValidateSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 30 {
		return fmt.Errorf("slug must be between 3 and 30 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}
	return nil
}

// Family represents a family group identified by a unique slug.
type Family struct {
	ID                            int64      `gorm:"primaryKey" json:"id"`
	Slug                          string     `gorm:"uniqueIndex;not null" json:"slug"`
	Timezone                      string     `gorm:"not null;default:America/New_York" json:"timezone"`
	BankName                      string     `gorm:"not null;default:Dad" json:"bank_name"`
	AccountType                   string     `gorm:"not null;default:free" json:"account_type"`
	StripeCustomerID              *string    `gorm:"uniqueIndex" json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID          *string    `gorm:"uniqueIndex" json:"stripe_subscription_id,omitempty"`
	SubscriptionStatus            *string    `json:"subscription_status,omitempty"`
	SubscriptionCurrentPeriodEnd  *time.Time `json:"subscription_current_period_end,omitempty"`
	SubscriptionCancelAtPeriodEnd bool       `gorm:"not null;default:false" json:"subscription_cancel_at_period_end"`
	CreatedAt                     time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Associations
	Parents  []Parent `gorm:"foreignKey:FamilyID" json:"-"`
	Children []Child  `gorm:"foreignKey:FamilyID" json:"-"`
}

func (Family) TableName() string {
	return "families"
}
