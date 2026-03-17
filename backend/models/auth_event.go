package models

import "time"

// AuthEvent represents a security audit log entry.
type AuthEvent struct {
	ID        int64     `gorm:"primaryKey"`
	EventType string    `gorm:"not null"`
	UserType  string    `gorm:"not null"`
	UserID    int64     `gorm:"not null;default:0"`
	FamilyID  int64     `gorm:"not null;default:0"`
	IPAddress string    `gorm:"not null"`
	Details   string    `gorm:"not null;default:''"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
