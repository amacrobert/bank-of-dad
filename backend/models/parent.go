package models

import "time"

// Parent represents a parent user authenticated via Google OAuth.
type Parent struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	GoogleID    string    `gorm:"uniqueIndex;not null" json:"google_id"`
	Email       string    `gorm:"not null" json:"email"`
	DisplayName string    `gorm:"not null" json:"display_name"`
	FamilyID    int64     `gorm:"not null;default:0" json:"family_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Associations
	Family Family `gorm:"foreignKey:FamilyID" json:"-"`
}
