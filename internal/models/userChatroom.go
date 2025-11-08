package models

import (
	"time"
)

type UserChatroom struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint       `gorm:"not null;index:idx_user_chatroom,unique" json:"user_id"`
	Name          string     `gorm:"type:varchar(100);not null" json:"name"`
	ChatroomID    uint       `gorm:"not null;index:idx_user_chatroom,unique" json:"chatroom_id"`
	IsAdmin       bool       `gorm:"default:false" json:"is_admin"`
	IsOwner       bool       `gorm:"default:false" json:"is_owner"`
	IsJoined      bool       `gorm:"default:false" json:"is_joined"`
	IsBanned      bool       `gorm:"default:false" json:"is_banned"`
	LastJoinTime  *time.Time `gorm:"autoUpdateTime" json:"last_join_time"`
	IsInvited     bool       `gorm:"default:false" json:"is_invited"`
	InviteExpires *time.Time `gorm:"default:null" json:"invite_expires_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
