package models

import "time"

// RefreshToken represents an opaque refresh token stored in the database.
type RefreshToken struct {
	ID        int64     `gorm:"primaryKey"`
	TokenHash string    `gorm:"uniqueIndex;not null"`
	UserType  string    `gorm:"not null"`
	UserID    int64     `gorm:"not null"`
	FamilyID  int64     `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
