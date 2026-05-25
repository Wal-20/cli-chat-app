package client

import (
	"encoding/json"

	"github.com/Wal-20/cli-chat-app/internal/models"
)

// GetServerConfig fetches the server's redacted runtime configuration.
func (c *APIClient) GetServerConfig() (models.ServerConfig, error) {
	body, err := c.get("/server/config")
	if err != nil {
		return models.ServerConfig{}, err
	}

	var resp models.ServerConfigResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return models.ServerConfig{}, err
	}
	return resp.Config, nil
}
