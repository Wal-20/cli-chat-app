package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Wal-20/cli-chat-app/internal/models"
)

var serverStartTime = time.Now()

// isSet reports whether an environment variable is present and non-empty.
func isSet(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}

func GetServerConfig(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	cfg := models.ServerConfig{
		ServerURL:  os.Getenv("SERVER_URL"),
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Goroutines: runtime.NumGoroutine(),

		StartedAt:  serverStartTime,
		Uptime:     now.Sub(serverStartTime).Round(time.Second).String(),
		ServerTime: now,

		DBHostConfigured:     isSet("db_host"),
		DBPortConfigured:     isSet("db_port"),
		DBNameConfigured:     isSet("db_name"),
		DBUserConfigured:     isSet("db_username"),
		DBPasswordConfigured: isSet("db_password"),
		JWTConfigured:        isSet("JWT_SECRET"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ServerConfigResponse{Config: cfg})
}
