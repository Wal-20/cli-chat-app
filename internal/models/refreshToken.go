package models

import (
	"time"
)

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"` // Primary key
	Token     string    `gorm:"not null" json:"token"`              // The refresh token (store hashed value)
	UserID    uint      `gorm:"not null;index" json:"user_id"`      // Foreign key to the Users table
	User      User      `gorm:"constraint:OnDelete:CASCADE;"`       // Association with Users table
	IssuedAt  time.Time `gorm:"not null" json:"issued_at"`          // Issuance time
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`         // Expiration time
	// IsRevoked bool      `gorm:"default:false" json:"is_revoked"`    // Revocation status
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`   // Creation timestamp
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`   // Update timestamp
}
