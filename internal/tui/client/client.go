package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"github.com/gorilla/websocket"
)

type APIClient struct {
	baseURL      string
	httpClient   *http.Client
	accessToken  string
	refreshToken string
	cache        *cache.Cache
}

var DefaultServerURLB64 string // built into the binary with ldflags, refer to ./build.sh

func NewAPIClient() (*APIClient, error) {
	// Load .env locally if available (safe no-op in production)
	_ = godotenv.Load()

	var serverURLB64 string
	source := ""

	// Prefer build-time injected variable
	if strings.TrimSpace(DefaultServerURLB64) != "" {
		serverURLB64 = strings.TrimSpace(DefaultServerURLB64)
		source = "build flag"
	} else if envURL := os.Getenv("SERVER_URL"); envURL != "" {
		// Encode environment URL to base64 for consistency
		serverURLB64 = base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(envURL)))
		source = "environment variable"
	}

	if serverURLB64 == "" {
		return nil, fmt.Errorf("SERVER_URL not found in build flags nor environment")
	}

	// Decode before using
	decodedURLBytes, err := base64.StdEncoding.DecodeString(serverURLB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 server URL: %w", err)
	}

	serverURL := strings.TrimSuffix(string(decodedURLBytes), "/")
	baseURL := serverURL + "/api"

	client := &http.Client{}

	// Quick health check for early feedback
	req, err := http.NewRequest("GET", baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned unexpected status: %s", resp.Status)
	}

	fmt.Printf("Connected to server at %s (source: %s)\n", serverURL, source)

	return &APIClient{
		baseURL:    baseURL,
		httpClient: client,
		cache:      cache.New(2*time.Minute, 30*time.Second),
	}, nil
}

func (c *APIClient) SetToken(token string) {
	c.accessToken = token
}

func (c *APIClient) SetTokenPair(access, refresh string) {
	c.accessToken = access
	c.refreshToken = refresh
}

// InvalidateUserChatrooms clears the cached user chatrooms list so next load is fresh.
func (c *APIClient) InvalidateUserChatrooms() {
	if c.cache != nil {
		c.cache.Delete("user_chatrooms")
	}
}

// Auth endpoints

func (c *APIClient) LoginOrRegister(username, password string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"name":     username,
		"password": password,
	}

	// Attempt login
	res, err := c.post("/users/login", data)
	if err == nil {
		return res, nil // Login successful
	}

	// Check if the error is due to user not existing
	if httpErr, ok := err.(interface{ Error() string }); ok && strings.Contains(httpErr.Error(), "User not found") {
		// Attempt registration if user does not exist
		return c.post("/users", data)
	}

	// Other errors (e.g., server issues, bad credentials) should not trigger registration
	return nil, err
}

func (c *APIClient) Logout() error {
	tokenPair, err := utils.LoadTokenPair()
	if err != nil {
		return fmt.Errorf("failed to load token pair: %w", err)
	}

	utils.AuthCache.Delete(tokenPair.AccessToken)
	utils.AuthCache.Delete(tokenPair.RefreshToken)

	tokenPair.AccessToken = ""
	tokenPair.RefreshToken = ""

	if err := utils.SaveTokenPair(tokenPair); err != nil {
		return fmt.Errorf("error clearing token pair: %w", err)
	}

	if c.cache != nil {
		c.cache.Flush()
	}
	return nil
}

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

// SubscribeChatroom opens a websocket to receive live messages for a chatroom.
// It returns:
//   - a channel of incoming events,
//   - a cancel function to close the stream,
//   - and a send function to push events (e.g., typing / presence) to the server.
func (c *APIClient) SubscribeChatroom(chatroomID uint) (<-chan models.WsEvent, func(), func(models.WsEvent) error, error) {
	if c.baseURL == "" {
		return nil, nil, nil, fmt.Errorf("client not initialized")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, nil, nil, err
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	u.Path = stdpath.Join(u.Path, fmt.Sprintf("chatrooms/%d/ws", chatroomID))

	// Prepare headers
	header := http.Header{}
	if c.accessToken != "" {
		header.Set("Authorization", "Bearer "+c.accessToken)
	}

	// Dial WS
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		// Attempt automatic token refresh
		if resp != nil && resp.StatusCode == 401 && c.refreshToken != "" {
			if rerr := c.refreshTokens(); rerr == nil {
				header = http.Header{}
				if c.accessToken != "" {
					header.Set("Authorization", "Bearer "+c.accessToken)
				}
				conn, resp, err = websocket.DefaultDialer.Dial(u.String(), header)
			}
		}
		// Still failed
		if err != nil {
			if resp != nil {
				return nil, nil, nil, fmt.Errorf("ws dial failed: %s", resp.Status)
			}
			return nil, nil, nil, fmt.Errorf("ws dial error: %w", err)
		}
	}

	ch := make(chan models.WsEvent, 32)
	go func() {
		defer close(ch)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var evt models.WsEvent
			if err := json.Unmarshal(data, &evt); err == nil {
				ch <- evt
			}
		}
	}()

	cancel := func() {
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
		)
		_ = conn.Close()
	}

	send := func(evt models.WsEvent) error {
		data, err := json.Marshal(evt)
		if err != nil {
			return err
		}
		return conn.WriteMessage(websocket.TextMessage, data)
	}

	return ch, cancel, send, nil
}

// Admin actions
func (c *APIClient) InviteUser(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/invite/%s", chatroomID, userID), nil)
	return err
}

func (c *APIClient) KickUser(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/kick/%s", chatroomID, userID), nil)
	return err
}

func (c *APIClient) BanUser(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/ban/%s", chatroomID, userID), nil)
	return err
}

func (c *APIClient) MakeAdmin(chatroomID, userID string) error {
	_, err := c.post(fmt.Sprintf("/users/chatrooms/%s/promote/%s", chatroomID, userID), nil)
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

// Helper methods for HTTP requests
func (c *APIClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil && isUnauthorized(err) && c.refreshToken != "" {
		if rerr := c.refreshTokens(); rerr == nil {
			req2, _ := http.NewRequest("GET", c.baseURL+path, nil)
			return c.doRequest(req2)
		}
	}
	return body, err
}

func (c *APIClient) post(path string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil && isUnauthorized(err) && c.refreshToken != "" {
		if rerr := c.refreshTokens(); rerr == nil {
			req2, _ := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(jsonData))
			resp, err = c.doRequest(req2)
		}
	}
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	return result, err
}

func (c *APIClient) delete(path string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", c.baseURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil && isUnauthorized(err) && c.refreshToken != "" {
		if rerr := c.refreshTokens(); rerr == nil {
			req2, _ := http.NewRequest("POST", c.baseURL+path, bytes.NewBuffer(jsonData))
			resp, err = c.doRequest(req2)
		}
	}
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	return result, err
}

func (c *APIClient) doRequest(req *http.Request) ([]byte, error) {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check if HTTP status code is not in the success range
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %s, Response: %s", resp.Status, string(body))
	}

	return body, nil // Return the actual response body
}

func isUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "HTTP error: 401")
}

func (c *APIClient) refreshTokens() error {
	payload := map[string]string{"refreshToken": c.refreshToken}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", c.baseURL+"/users/refresh", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("refresh failed: %s, Response: %s", resp.Status, string(rb))
	}
	var res map[string]any
	if err := json.Unmarshal(rb, &res); err != nil {
		return err
	}
	newAccess, _ := res["AccessToken"].(string)
	newRefresh, _ := res["RefreshToken"].(string)
	if newAccess == "" || newRefresh == "" {
		return fmt.Errorf("refresh failed: missing tokens")
	}
	c.accessToken = newAccess
	c.refreshToken = newRefresh
	if c.cache != nil {
		c.cache.Flush()
	}
	_ = utils.SaveTokenPair(utils.TokenPair{AccessToken: newAccess, RefreshToken: newRefresh})
	return nil
}
