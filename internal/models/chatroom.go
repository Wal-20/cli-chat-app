package models


type Chatroom struct {
	Id uint `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminId int `gorm:"foreignKey:ID" json:"adminId"`
	MaxUserCount uint  `gorm:"type(int)" json:"maxUserCount"`
	Users []User  `gorm:"many2many:user_chatrooms;" json:"users"`
	IsPublic bool `gorm:"type(bool);default:false" json:"isPublic"`
	CreatedAt string `gorm:"autoCreateTime" json:"created_at"`
}

