package models

import (
	"time"
)

type UserChatroom struct {
	UserID       uint      `gorm:"primaryKey" json:"user_id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	ChatroomID   uint      `gorm:"primaryKey" json:"chatroom_id"`
	IsAdmin      bool      `gorm:"default:false" json:"is_admin"`
	IsOwner      bool      `gorm:"default:false" json:"is_owner"`
	IsJoined     bool `gorm:"default:true" json:"is_joined"`
	IsBanned     bool      `gorm:"default:false" json:"is_banned"`
	LastJoinTime *time.Time `gorm:"autoUpdateTime" json:"last_join_time"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsInvited    bool      `gorm:"default:false" json:"is_invited"`
	InviteExpires *time.Time `gorm:"default:null" json:"invite_expires_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}


