package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/utils"
)

// Auth endpoints

func (c *APIClient) LoginOrRegister(username, password string) (map[string]any, error) {
	data := map[string]any{
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
