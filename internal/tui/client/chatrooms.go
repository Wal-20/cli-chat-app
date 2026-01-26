package client

import (
	"encoding/json"
	"fmt"

	"github.com/patrickmn/go-cache"

	"github.com/Wal-20/cli-chat-app/internal/models"
)

// Chatroom endpoints
func (c *APIClient) GetChatrooms() ([]models.Chatroom, error) {
	resp, err := c.get("/chatrooms/public")
	if err != nil {
		return nil, err
	}

	var result struct {
		Chatrooms []models.Chatroom `json:"Chatrooms"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	return result.Chatrooms, nil
}

func (c *APIClient) GetUserChatrooms() ([]models.Chatroom, error) {
	// Try cache first
	if c.cache != nil {
		if v, ok := c.cache.Get("user_chatrooms"); ok {
			if rooms, ok := v.([]models.Chatroom); ok {
				// Return a copy to avoid external mutation of cached slice
				cp := append([]models.Chatroom(nil), rooms...)
				return cp, nil
			}
		}
	}

	resp, err := c.get("/users/chatrooms")
	if err != nil {
		return nil, err
	}
	var result struct {
		Chatrooms []models.Chatroom `json:"Chatrooms"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if c.cache != nil {
		// Store a copy
		cp := append([]models.Chatroom(nil), result.Chatrooms...)
		c.cache.Set("user_chatrooms", cp, cache.DefaultExpiration)
	}
	return result.Chatrooms, nil
}

func (c *APIClient) GetUsersByChatroom(chatroomID uint, active bool) ([]models.UserChatroom, error) {
	endpoint := fmt.Sprintf("/chatrooms/%v/users", chatroomID)
	if active {
		endpoint += "?active=true"
	}

	resp, err := c.get(endpoint)
	if err != nil {
		// Log the actual error body for debugging
		return nil, fmt.Errorf("failed GET %s: %w", endpoint, err)
	}

	var result struct {
		UserChatroom []models.UserChatroom `json:"userChatroom"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w. Raw response: %s", err, string(resp))
	}

	return result.UserChatroom, nil
}

func (c *APIClient) DeleteChatroom(id uint) error {
	_, err := c.delete(fmt.Sprintf("/chatrooms/%v", id), nil)
	if err == nil && c.cache != nil {
		c.cache.Delete("user_chatrooms")
	}
	return err
}

func (c *APIClient) JoinChatroom(chatroomID uint) error {
	_, err := c.post(fmt.Sprintf("/chatrooms/%v/join", chatroomID), nil)
	if err == nil && c.cache != nil {
		// User chatroom list likely changed
		c.cache.Delete("user_chatrooms")
	}
	return err
}

// JoinChatroomVerbose performs a join and returns the raw response map so the UI can display status.
func (c *APIClient) JoinChatroomVerbose(chatroomID uint) (map[string]any, error) {
	res, err := c.post(fmt.Sprintf("/chatrooms/%v/join", chatroomID), nil)
	if err == nil && c.cache != nil {
		c.cache.Delete("user_chatrooms")
	}
	return res, err
}

func (c *APIClient) LeaveChatroom(chatroomID string) error {
	_, err := c.post(fmt.Sprintf("/chatrooms/%s/leave", chatroomID), nil)
	if err == nil && c.cache != nil {
		c.cache.Delete("user_chatrooms")
		// Also drop any cached messages for this room
		c.cache.Delete("chatroom_messages:" + chatroomID)
	}
	return err
}

// CreateChatroom creates a chatroom; recipient is optional (handled server-side).
func (c *APIClient) CreateChatroom(title string, maxUsers int, isPublic bool) (models.Chatroom, error) {
	data := map[string]any{
		"title":        title,
		"maxUserCount": maxUsers,
		"is_public":    isPublic,
	}
	res, err := c.post("/chatrooms", data)
	if err != nil {
		return models.Chatroom{}, err
	}
	// The API returns a Chatroom or a map depending on handler; handle both
	// Try to marshal back into Chatroom
	b, _ := json.Marshal(res)
	var room models.Chatroom
	if err := json.Unmarshal(b, &room); err == nil && room.Id != 0 {
		return room, nil
	}
	// Fallback if server wraps object (e.g., {"Chatroom": {...}})
	if v, ok := res["Chatroom"]; ok {
		rb, _ := json.Marshal(v)
		if err := json.Unmarshal(rb, &room); err == nil {
			return room, nil
		}
	}
	return models.Chatroom{}, fmt.Errorf("unexpected response creating chatroom")
}
