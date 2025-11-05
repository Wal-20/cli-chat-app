package services

import (
	"github.com/Wal-20/cli-chat-app/internal/config"
	"github.com/Wal-20/cli-chat-app/internal/models"
)

// NotificationService coordinates retrieval of user notifications and related invite metadata.
type NotificationService struct{}

func NewNotificationService() *NotificationService { return &NotificationService{} }

// GetUserNotifications fetches active notifications and pending chatroom invites for the given user.
func (s *NotificationService) GetUserNotifications(userID uint) (models.NotificationsResponse, error) {
	var resp models.NotificationsResponse
	if err := config.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&resp.Notifications).Error; err != nil {
		return resp, err
	}
	return resp, nil
}
