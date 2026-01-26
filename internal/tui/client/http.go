package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func (c *APIClient) post(path string, data any) (map[string]any, error) {
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

	var result map[string]any
	err = json.Unmarshal(resp, &result)
	return result, err
}

func (c *APIClient) delete(path string, data any) (map[string]any, error) {
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

	var result map[string]any
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
