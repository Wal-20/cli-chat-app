package models

import (
	"time"
)

type Chatroom struct {
	Id uint `gorm:"primaryKey;autoIncrement" json:"id"`
	Title string `gorm:"type:varchar(100);not null" json:"title"`
	AdminId uint `gorm:"foreignKey:ID" json:"adminId"`
	MaxUserCount uint  `gorm:"type(int);default:10" json:"maxUserCount"`
	Users []User  `gorm:"many2many:user_chatrooms;" json:"users"`
	IsPublic bool `gorm:"type(bool);default:false" json:"isPublic"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

