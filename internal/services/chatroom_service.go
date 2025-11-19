package services

import (
	"fmt"
	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/repositories"

	"gorm.io/gorm"
	"time"
)

type ChatroomService struct {
	repo repositories.ChatroomRepository
}

func NewChatroomService(r repositories.ChatroomRepository) *ChatroomService {
	return &ChatroomService{repo: r}
}

func (s *ChatroomService) GetPublicChatrooms(userID uint) ([]models.Chatroom, error) {
	return s.repo.GetPublicChatroomsNotJoined(userID)
}

func (s *ChatroomService) JoinChatroom(userID uint, username string, chatroomID string) (*models.Chatroom, error) {
	chatroom, err := s.repo.FindByID(chatroomID)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.FindUserChatroom(userID, chatroom.Id); err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a placeholder membership without joining yet; handler validates and flips joined.
			now := time.Now()
			uc := &models.UserChatroom{UserID: userID, Name: username, ChatroomID: chatroom.Id, LastJoinTime: &now, IsJoined: false}
			if err := s.repo.CreateUserChatroom(uc); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return chatroom, nil
}

func (s *ChatroomService) LeaveChatroom(userID uint, chatroomID string, wasOwner bool) (map[string]any, error) {
	uc, err := s.repo.FindUserChatroom(userID, chatroomID)
	if err != nil {
		return nil, err
	}
	if !uc.IsJoined {
		return nil, fmt.Errorf("User already not part of chatroom")
	}
	uc.IsJoined = false
	uc.IsInvited = false
	if wasOwner || uc.IsOwner {
		uc.IsOwner = false
		uc.IsAdmin = false
	}
	if err := s.repo.SaveUserChatroom(uc); err != nil {
		return nil, err
	}

	// count remaining
	remaining, err := s.repo.CountJoinedUsers(chatroomID)
	if err != nil {
		return nil, err
	}
	if remaining < 1 {
		if err := s.repo.DeleteChatroomByID(chatroomID); err != nil {
			return nil, err
		}
		return map[string]any{"Status": "Chatroom deleted as last user left"}, nil
	}
	return map[string]any{"Status": "Left chatroom successfully", "data": uc}, nil
}

func RemoveOldUserChatrooms() (int64, error) {
	threshold := time.Now().AddDate(0, 0, -60) // older than 2 months

	result := config.DB.
		Where("created_at < ? AND (is_banned = true OR is_joined = false)", threshold).
		Delete(&models.UserChatroom{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

