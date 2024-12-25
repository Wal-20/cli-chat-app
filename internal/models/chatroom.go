package models


type chatroom struct {
	Id uint `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminId int `gorm:"foreignKey:ID" json:"adminId"`
	MaxUserCount uint  `gorm:"type(int)" json:"maxUserCount"`
	IsPublic bool `gorm:"type(bool);default:false" json:"isPublic"`
	CreatedAt string `gorm:"autoCreateTime" json:"created_at"`
}

