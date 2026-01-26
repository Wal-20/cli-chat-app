package models

import (
	"time"
)

type User struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"`
	Password  string     `gorm:"type:varchar(100);not null" json:"password"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `gorm:"type:datetime" json:"last_login"`
	Chatrooms []Chatroom `gorm:"many2many:user_chatrooms;" json:"chatrooms"`
}
