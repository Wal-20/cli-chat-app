package models

import (
	"fmt"
	"strings"

	appmodels "github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ServerConfigModal shows the server's redacted runtime configuration.
type ServerConfigModal struct {
	apiClient *client.APIClient
	returnTo  tea.Model

	config  appmodels.ServerConfig
	loaded  bool
	loading bool
	err     error
	width   int
	height  int
}

func NewServerConfigModal(api *client.APIClient, returnTo tea.Model) ServerConfigModal {
	return ServerConfigModal{
		apiClient: api,
		returnTo:  returnTo,
		loading:   true,
	}
}

func (m ServerConfigModal) Init() tea.Cmd { return loadServerConfig(m.apiClient) }

type serverConfigLoadedMsg struct {
	config appmodels.ServerConfig
	err    error
}

func loadServerConfig(api *client.APIClient) tea.Cmd {
	return func() tea.Msg {
		cfg, err := api.GetServerConfig()
		return serverConfigLoadedMsg{config: cfg, err: err}
	}
}

func (m ServerConfigModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m.returnTo, nil
		case "r":
			if m.loading {
				return m, nil
			}
			m.loading = true
			m.err = nil
			return m, loadServerConfig(m.apiClient)
		}
	case serverConfigLoadedMsg:
		m.loading = false
		m.loaded = true
		m.err = msg.err
		m.config = msg.config
		return m, nil
	}
	return m, nil
}

func (m ServerConfigModal) View() string {
	title := styles.CardTitleStyle.Render("Server Configuration")
	subtitle := styles.CardSubtitleStyle.Render("Redacted runtime config (secrets are never shown)")

	var bodyView string
	switch {
	case m.loading:
		bodyView = styles.StatusInfoStyle.Render("Loading server configuration...")
	case m.err != nil:
		bodyView = styles.StatusErrorStyle.Render(m.err.Error())
	default:
		bodyView = renderConfigRows(m.config)
	}

	help := styles.HelpStyle.Render(strings.Join([]string{
		styles.RenderKeyBinding("r", "Refresh"),
		styles.RenderKeyBinding("Esc", "Back"),
	}, styles.HelpStyle.Render("  ")))

	content := strings.Join([]string{title, subtitle, bodyView, help}, "\n\n")
	card := styles.CardStyle.Copy().Width(64).Render(content)

	if m.width > 0 && m.height > 0 {
		centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
		return styles.AppStyle.Copy().Width(m.width).Height(m.height).Render(centered)
	}
	return styles.AppStyle.Render(card)
}

func renderConfigRows(c appmodels.ServerConfig) string {
	rows := [][2]string{
		{"Server URL", orDash(c.ServerURL)},
		{"DB host", configuredLabel(c.DBHostConfigured)},
		{"DB port", configuredLabel(c.DBPortConfigured)},
		{"DB name", configuredLabel(c.DBNameConfigured)},
		{"DB user", configuredLabel(c.DBUserConfigured)},
		{"DB password", configuredLabel(c.DBPasswordConfigured)},
		{"JWT secret", configuredLabel(c.JWTConfigured)},
		{"Go version", orDash(c.GoVersion)},
		{"Platform", fmt.Sprintf("%s/%s", c.OS, c.Arch)},
		{"Goroutines", fmt.Sprintf("%d", c.Goroutines)},
		{"Uptime", orDash(c.Uptime)},
		{"Server time", c.ServerTime.Format("2006-01-02 15:04:05 MST")},
	}

	var b strings.Builder
	for i, row := range rows {
		label := styles.MutedTextStyle.Render(fmt.Sprintf("%-14s", row[0]))
		b.WriteString(label + "  " + row[1])
		if i < len(rows)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return styles.MutedTextStyle.Render("—")
	}
	return s
}

func configuredLabel(ok bool) string {
	if ok {
		return styles.StatusSuccessStyle.Render("configured")
	}
	return styles.StatusErrorStyle.Render("not set")
}
