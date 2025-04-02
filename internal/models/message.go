package models

import (
	"time"
)

type Message struct {
	ID         uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatroomID uint   `gorm:"foreignKey:ID" json:"chatroomId"`
	UserId     uint   `gorm:"foreignKey:ID" json:"userID"`
	Content    string `gorm:"type(text)" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
