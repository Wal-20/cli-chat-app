package models

import (
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ServerDownModel shows a centered, friendly message when the API server is unreachable.
type ServerDownModel struct {
	width  int
	height int
}

func NewServerDownModel() ServerDownModel {
	return ServerDownModel{}
}

func (m ServerDownModel) Init() tea.Cmd { return nil }

func (m ServerDownModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ServerDownModel) View() string {
	// Content width capped for better readability
	cw := m.width - 8
	if cw > 64 {
		cw = 64
	}
	if cw < 32 {
		cw = m.width - 4
		if cw < 20 {
			cw = m.width
		}
	}

	title := styles.TitleStyle.Render("CLI Chat")
	friendly := styles.MutedTextStyle.Render("We can’t reach the server right now. Please try again later.")

	// Center each section within the content width
	titleLine := lipgloss.Place(cw, 1, lipgloss.Center, lipgloss.Center, title)
	bodyLine := lipgloss.Place(cw, 1, lipgloss.Center, lipgloss.Center, friendly)

	helpItems := []string{styles.RenderKeyBinding("q", "Quit")}
	help := strings.Join(helpItems, styles.HelpStyle.Render("  "))
	footer := lipgloss.Place(cw, 1, lipgloss.Center, lipgloss.Center, styles.HelpStyle.Render(help))

	card := styles.CardStyle.Copy().Width(cw).Render(strings.Join([]string{
		titleLine,
		"",
		bodyLine,
		"",
		footer,
	}, "\n"))

	if m.width > 0 && m.height > 0 {
		centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
		return styles.AppStyle.Copy().Width(m.width).Height(m.height).Render(centered)
	}
	return styles.AppStyle.Render(card)
}
