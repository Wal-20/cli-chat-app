package repositories

import (
    "github.com/Wal-20/cli-chat-app/internal/config"
    "github.com/Wal-20/cli-chat-app/internal/models"
    "gorm.io/gorm"
)

type UserRepository interface {
    FindByID(id uint) (*models.User, error)
    FindByName(name string) (*models.User, error)
    Create(user *models.User) error
    Save(user *models.User) error
    DeleteByID(id uint) error
    GetChatroomsByUserID(userID uint) ([]models.Chatroom, error)
}

type GormUserRepository struct { db *gorm.DB }

func NewUserRepository(db *gorm.DB) *GormUserRepository { return &GormUserRepository{db: db} }

func (r *GormUserRepository) FindByID(id uint) (*models.User, error) {
    var u models.User
    if err := r.db.First(&u, id).Error; err != nil { return nil, err }
    return &u, nil
}

func (r *GormUserRepository) FindByName(name string) (*models.User, error) {
    var u models.User
    if err := r.db.Where("name = ?", name).First(&u).Error; err != nil { return nil, err }
    return &u, nil
}

func (r *GormUserRepository) Create(user *models.User) error { return r.db.Create(user).Error }

func (r *GormUserRepository) Save(user *models.User) error { return r.db.Save(user).Error }

func (r *GormUserRepository) DeleteByID(id uint) error { return r.db.Delete(&models.User{}, id).Error }

func (r *GormUserRepository) GetChatroomsByUserID(userID uint) ([]models.Chatroom, error) {
    var user models.User
    if err := r.db.Preload("Chatrooms").First(&user, userID).Error; err != nil { return nil, err }
    return user.Chatrooms, nil
}

// Default global constructor using app DB
func DefaultUserRepository() UserRepository { return NewUserRepository(config.DB) }

