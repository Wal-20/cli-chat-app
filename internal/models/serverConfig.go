package models

import "time"

// ServerConfig is a redacted, allowlisted view of the server's runtime
// configuration. It never exposes DB connection values nor secrets (JWT secret,
// DB password, full DB URI); for those, only a boolean "configured" flag is shown.
type ServerConfig struct {
	ServerURL  string `json:"serverUrl"`
	GoVersion  string `json:"goVersion"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Goroutines int    `json:"goroutines"`

	StartedAt  time.Time `json:"startedAt"`
	Uptime     string    `json:"uptime"`
	ServerTime time.Time `json:"serverTime"`

	DBHostConfigured     bool `json:"dbHostConfigured"`
	DBPortConfigured     bool `json:"dbPortConfigured"`
	DBNameConfigured     bool `json:"dbNameConfigured"`
	DBUserConfigured     bool `json:"dbUserConfigured"`
	DBPasswordConfigured bool `json:"dbPasswordConfigured"`
	JWTConfigured        bool `json:"jwtConfigured"`
}

// ServerConfigResponse is the envelope returned by GET /api/server/config.
type ServerConfigResponse struct {
	Config ServerConfig `json:"Config"`
}
