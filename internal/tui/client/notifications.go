package client

import (
	"encoding/json"
	"fmt"

	"github.com/Wal-20/cli-chat-app/internal/models"
)

func (c *APIClient) GetNotifications() (models.NotificationsResponse, error) {
	var result models.NotificationsResponse
	body, err := c.get("/users/notifications")
	if err != nil {
		return result, err
	}
	if len(body) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return models.NotificationsResponse{}, err
	}
	return result, nil
}

func (c *APIClient) DeleteNotification(id uint) error {
	_, err := c.delete(fmt.Sprintf("/notifications/%v", id), nil)
	if err == nil && c.cache != nil {
		c.cache.Delete("notifications")
	}
	return err
}
