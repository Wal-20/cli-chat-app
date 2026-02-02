package client

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

type APIClient struct {
	baseURL      string
	httpClient   *http.Client
	accessToken  string
	refreshToken string
	cache        *cache.Cache
}

// built into the binary with ldflags, refer to ./build.sh
var DefaultServerURLB64 string
var DefaultJWTSecret string

func NewAPIClient() (*APIClient, error) {
	// Load .env locally if available (safe no-op in production)
	_ = godotenv.Load()

	// Ensure JWT secret is available for client-side token parsing.
	if os.Getenv("JWT_SECRET") == "" && strings.TrimSpace(DefaultJWTSecret) != "" {
		_ = os.Setenv("JWT_SECRET", strings.TrimSpace(DefaultJWTSecret))
	}

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
