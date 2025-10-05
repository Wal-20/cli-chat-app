package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    stdpath "path"
    "path/filepath"
    "strings"
    "time"

    "github.com/joho/godotenv"
    "github.com/patrickmn/go-cache"

    "github.com/Wal-20/cli-chat-app/internal/models"
    "github.com/gorilla/websocket"
)

type APIClient struct {
    baseURL     string
    httpClient  *http.Client
    accessToken string
    cache       *cache.Cache
}

// loadEnvFile attempts to ensure SERVER_URL is available. It first respects an
// existing environment variable, then attempts to load .env from the current
// directory or any parent directory up to the filesystem root.
func loadEnvFile() error {
	if strings.TrimSpace(os.Getenv("SERVER_URL")) != "" {
		return nil
	}

	if err := godotenv.Load(); err == nil { // CWD
		if strings.TrimSpace(os.Getenv("SERVER_URL")) != "" {
			return nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		envPath := filepath.Join(cwd, ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			if err := godotenv.Load(envPath); err == nil {
				if strings.TrimSpace(os.Getenv("SERVER_URL")) != "" {
					return nil
				}
			}
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}

	return fmt.Errorf("SERVER_URL not set; create a .env with SERVER_URL or export it")
}

func NewAPIClient() (*APIClient, error) {
	if err := loadEnvFile(); err != nil {
		return nil, fmt.Errorf("load environment: %w", err)
	}

	serverURL := strings.TrimSuffix(strings.TrimSpace(os.Getenv("SERVER_URL")), "/")
	if serverURL == "" {
		return nil, fmt.Errorf("SERVER_URL is not set in environment")
	}

	baseURL := serverURL + "/api"
	client := &http.Client{}

	// quick health check so we can surface a nice error early
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

    return &APIClient{
        baseURL:    baseURL,
        httpClient: client,
        // Cache with a small TTL to improve UX while keeping data fresh
        cache:      cache.New(2*time.Minute, 30*time.Second),
    }, nil
}

func (c *APIClient) SetToken(token string) {
	c.accessToken = token
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
    _, err := c.post("/users/logout", nil)
    if err == nil && c.cache != nil {
        c.cache.Flush()
    }
    return err
}

// Chatroom endpoints
func (c *APIClient) GetChatrooms() ([]models.Chatroom, error) {
	resp, err := c.get("/chatrooms/public")
	if err != nil {
		return nil, err
	}

	var result struct {
		Chatrooms []models.Chatroom `json:"chatrooms"` // Match API structure
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

func (c *APIClient) GetUsersByChatroom(chatroomID uint) ([]models.UserChatroom, error) {
	endpoint := fmt.Sprintf("/chatrooms/%v/users", chatroomID)
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

func (c *APIClient) JoinChatroom(chatroomID uint) error {
    _, err := c.post(fmt.Sprintf("/chatrooms/%v/join", chatroomID), nil)
    if err == nil && c.cache != nil {
        // User chatroom list likely changed
        c.cache.Delete("user_chatrooms")
    }
    return err
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
    key := fmt.Sprintf("chatroom_messages:%v", chatroomID)
    if c.cache != nil {
        if v, ok := c.cache.Get(key); ok {
            if msgs, ok := v.([]models.MessageWithUser); ok {
                cp := append([]models.MessageWithUser(nil), msgs...)
                return cp, nil
            }
        }
    }

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
    if c.cache != nil {
        cp := append([]models.MessageWithUser(nil), result.Messages...)
        c.cache.Set(key, cp, cache.DefaultExpiration)
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
    // Invalidate cached messages for this chatroom to ensure refresh
    if c.cache != nil {
        c.cache.Delete("chatroom_messages:" + chatroomID)
    }
    return res, nil
}

// SubscribeChatroom opens a websocket to receive live messages for a chatroom.
// It returns a channel of messages and a cancel function to close the stream.
func (c *APIClient) SubscribeChatroom(chatroomID uint) (<-chan models.MessageWithUser, func(), error) {
    if c.baseURL == "" {
        return nil, nil, fmt.Errorf("client not initialized")
    }
    u, err := url.Parse(c.baseURL)
    if err != nil {
        return nil, nil, err
    }
    if u.Scheme == "https" {
        u.Scheme = "wss"
    } else {
        u.Scheme = "ws"
    }
    u.Path = stdpath.Join(u.Path, fmt.Sprintf("chatrooms/%d/ws", chatroomID))

    header := http.Header{}
    if c.accessToken != "" {
        header.Set("Authorization", "Bearer "+c.accessToken)
    }
    conn, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
    if err != nil {
        if resp != nil {
            return nil, nil, fmt.Errorf("ws dial failed: %s", resp.Status)
        }
        return nil, nil, fmt.Errorf("ws dial error: %w", err)
    }

    ch := make(chan models.MessageWithUser, 32)
    go func() {
        defer close(ch)
        for {
            _, data, err := conn.ReadMessage()
            if err != nil {
                return
            }
            var m models.MessageWithUser
            if err := json.Unmarshal(data, &m); err == nil {
                ch <- m
            }
        }
    }()

    cancel := func() {
        _ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
        _ = conn.Close()
    }
    return ch, cancel, nil
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

// Helper methods for HTTP requests
func (c *APIClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
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
