package repositories

import (
    "github.com/Wal-20/cli-chat-app/internal/config"
    "github.com/Wal-20/cli-chat-app/internal/models"
    "gorm.io/gorm"
)

type ChatroomRepository interface {
    FindByID(id any) (*models.Chatroom, error)
    GetPublicChatroomsNotJoined(userID uint) ([]models.Chatroom, error)
    FindUserChatroom(userID any, chatroomID any) (*models.UserChatroom, error)
    CreateUserChatroom(uc *models.UserChatroom) error
    SaveUserChatroom(uc *models.UserChatroom) error
    CountJoinedUsers(chatroomID any) (int64, error)
    DeleteChatroomByID(id any) error
    DeleteUserChatroomsByChatroomID(chatroomID any) error
    SaveNotification(n *models.Notification) error
    SaveChatroom(c *models.Chatroom) error
}

type GormChatroomRepository struct { db *gorm.DB }

func NewChatroomRepository(db *gorm.DB) *GormChatroomRepository { return &GormChatroomRepository{db: db} }

func (r *GormChatroomRepository) FindByID(id any) (*models.Chatroom, error) {
    var c models.Chatroom
    if err := r.db.First(&c, id).Error; err != nil { return nil, err }
    return &c, nil
}

func (r *GormChatroomRepository) GetPublicChatroomsNotJoined(userID uint) ([]models.Chatroom, error) {
    var chatrooms []models.Chatroom
    // Exclude only rooms where the user is currently joined
    sub := r.db.Table("user_chatrooms").
        Select("chatroom_id").
        Where("user_id = ? AND is_joined = ?", userID, true)
    err := r.db.Preload("Users").
        Where("id NOT IN (?) AND is_public = ?", sub, true).
        Find(&chatrooms).Error
    return chatrooms, err
}

func (r *GormChatroomRepository) FindUserChatroom(userID any, chatroomID any) (*models.UserChatroom, error) {
    var uc models.UserChatroom
    if err := r.db.Where("user_id = ? AND chatroom_id = ?", userID, chatroomID).First(&uc).Error; err != nil { return nil, err }
    return &uc, nil
}

func (r *GormChatroomRepository) CreateUserChatroom(uc *models.UserChatroom) error { return r.db.Create(uc).Error }
func (r *GormChatroomRepository) SaveUserChatroom(uc *models.UserChatroom) error   { return r.db.Save(uc).Error }

func (r *GormChatroomRepository) CountJoinedUsers(chatroomID any) (int64, error) {
    var count int64
    err := r.db.Model(&models.UserChatroom{}).Where("chatroom_id = ? AND is_joined = ?", chatroomID, true).Count(&count).Error
    return count, err
}

func (r *GormChatroomRepository) DeleteChatroomByID(id any) error { return r.db.Where("id = ?", id).Delete(&models.Chatroom{}).Error }
func (r *GormChatroomRepository) DeleteUserChatroomsByChatroomID(chatroomID any) error {
    return r.db.Where("chatroom_id = ?", chatroomID).Delete(&models.UserChatroom{}).Error
}
func (r *GormChatroomRepository) SaveNotification(n *models.Notification) error { return r.db.Save(n).Error }
func (r *GormChatroomRepository) SaveChatroom(c *models.Chatroom) error { return r.db.Save(c).Error }

func DefaultChatroomRepository() ChatroomRepository { return NewChatroomRepository(config.DB) }
