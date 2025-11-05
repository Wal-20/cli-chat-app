package models

import (
	"time"
)

type Notification struct {
	Id         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId     uint      `gorm:"foreignKey:ID" json:"userID"`
	ChatroomId uint      `json:"chatroom"`
	Type       string    `json:"type"`
	SenderId   uint      `json:"senderId"`
	Content    string    `gorm:"type(text)" json:"content"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	ExpiresAt  time.Time `gorm:"index" json:"expires_at"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}
