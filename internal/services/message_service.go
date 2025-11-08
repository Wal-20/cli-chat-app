package services

import (
	"github.com/Wal-20/cli-chat-app/internal/api/ws"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/repositories"
)

type MessageService struct {
	messages repositories.MessageRepository
	users    repositories.UserRepository
}

func NewMessageService(m repositories.MessageRepository, u repositories.UserRepository) *MessageService {
	return &MessageService{messages: m, users: u}
}

func (s *MessageService) SendMessage(senderID, chatroomID uint, content string) (models.Message, string, error) {
	msg := models.Message{UserId: senderID, ChatroomID: chatroomID, Content: content}
	if err := s.messages.Create(&msg); err != nil {
		return models.Message{}, "", err
	}

	user, err := s.users.FindByID(senderID)
	if err == nil {
		ws.BroadcastMessage(chatroomID, models.MessageWithUser{Content: msg.Content, CreatedAt: msg.CreatedAt, Username: user.Name})
	}
	return msg, user.Name, nil
}
