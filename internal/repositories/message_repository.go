package repositories

import (
    "github.com/Wal-20/cli-chat-app/internal/config"
    "github.com/Wal-20/cli-chat-app/internal/models"
    "gorm.io/gorm"
)

type MessageRepository interface {
    Create(message *models.Message) error
}

type GormMessageRepository struct { db *gorm.DB }

func NewMessageRepository(db *gorm.DB) *GormMessageRepository { return &GormMessageRepository{db: db} }

func (r *GormMessageRepository) Create(message *models.Message) error { return r.db.Create(message).Error }

func DefaultMessageRepository() MessageRepository { return NewMessageRepository(config.DB) }

