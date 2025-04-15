package client

import (
	"bytes"
	"encoding/json"
	"strings"
	"github.com/joho/godotenv"
	"os"
	"fmt"
	"log"
	"io"
	"net/http"
	"github.com/Wal-20/cli-chat-app/internal/models"
)

type APIClient struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
}


func NewAPIClient() (*APIClient, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	SERVER_URL := os.Getenv("SERVER_URL") 
	if SERVER_URL == "" {
		log.Fatal("CANNOT READ SERVER_URI IN ENVIRONMENT")
	}

	baseURL := SERVER_URL + "/api"
	client := &http.Client{}

	req, err := http.NewRequest("GET", baseURL + "/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned unexpected status: %s", resp.Status)
	}

	return &APIClient{
		baseURL:    baseURL,
		httpClient: client,
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
	return err
}

// Chatroom endpoints
func (c *APIClient) GetChatrooms() ([]models.Chatroom, error) {
	resp, err := c.get("/chatrooms")
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
	resp, err := c.get("/users/chatrooms")
	if err != nil {
		return nil, err
	}
	var result struct {
		Chatrooms []models.Chatroom `json:"Chatrooms"`
	}
	err = json.Unmarshal(resp, &result)
	return result.Chatrooms, err
}

func (c *APIClient) JoinChatroom(chatroomID string) error {
	_, err := c.post(fmt.Sprintf("/chatrooms/%s/join", chatroomID), nil)
	return err
}

func (c *APIClient) LeaveChatroom(chatroomID string) error {
	_, err := c.post(fmt.Sprintf("/chatrooms/%s/leave", chatroomID), nil)
	return err
}

// Message endpoints
func (c *APIClient) GetMessages(chatroomID uint) ([]models.Message, error) {
	resp, err := c.get(fmt.Sprintf("/chatrooms/%v/messages", chatroomID))
	if err != nil {
		return nil, err
	}
	var result struct {
		Messages []models.Message `json:"Messages"`
	}
	err = json.Unmarshal(resp, &result)
	return result.Messages, err
}

func (c *APIClient) SendMessage(chatroomID, content string) (map[string]any, error){
	data := map[string]any{
		"content": content,
	}
	res, err := c.post(fmt.Sprintf("/chatrooms/%s/messages", chatroomID), data)
	if err != nil {
		return nil, err
	}
	return res, nil
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

