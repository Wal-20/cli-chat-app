package models

import (
	"time"
)

type Notification struct {
	Id uint `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId     uint   `gorm:"foreignKey:ID" json:"userID"`
	Content    string `gorm:"type(text)" json:"content"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"` // Add expiration time
}

