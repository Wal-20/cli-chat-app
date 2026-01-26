package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/models"
)

// Message endpoints
func (c *APIClient) GetMessages(chatroomID uint) ([]models.MessageWithUser, error) {
	resp, err := c.get(fmt.Sprintf("/chatrooms/%v/messages", chatroomID))
	if err != nil {
		return nil, err
	}
	var result struct {
		Messages []models.MessageWithUser `json:"Messages"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Messages, nil
}

// GetMessagesWithSearch fetches messages with an optional search query.
func (c *APIClient) GetMessagesWithSearch(chatroomID uint, search string) ([]models.MessageWithUser, error) {
	path := fmt.Sprintf("/chatrooms/%v/messages", chatroomID)
	if strings.TrimSpace(search) != "" {
		path = path + "?search=" + url.QueryEscape(search)
	}
	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Messages []models.MessageWithUser `json:"Messages"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Messages, nil
}

func (c *APIClient) SendMessage(chatroomID, content string) (map[string]any, error) {
	data := map[string]any{
		"content": content,
	}
	res, err := c.post(fmt.Sprintf("/chatrooms/%s/messages", chatroomID), data)
	if err != nil {
		return nil, err
	}
	return res, nil
}
