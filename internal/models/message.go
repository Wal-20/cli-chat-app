package models

import (
	"time"
)

// this is used for sending the message to the backend
type Message struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatroomID uint      `gorm:"foreignKey:ID" json:"chatroomId"`
	UserId     uint      `gorm:"foreignKey:ID" json:"userID"`
	Content    string    `gorm:"type(text)" json:"content"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// this is used for retrieving the messages on the client
type MessageWithUser struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
}
