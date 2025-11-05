package models

import (
	"strconv"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// JoinChatroomModal provides a centered input to join a chatroom by ID.
type JoinChatroomModal struct {
	apiClient *client.APIClient
	returnTo  tea.Model

	input      textinput.Model
	submitting bool
	width      int
	height     int
	status     string
	statusOkay bool
}

func NewJoinChatroomModal(api *client.APIClient, returnTo tea.Model) JoinChatroomModal {
	in := textinput.New()
	in.Prompt = "> "
	in.Placeholder = "chatroom id (e.g., 42)"
	in.PromptStyle = styles.InputPromptFocusedStyle
	in.TextStyle = styles.InputTextFocusedStyle
	in.PlaceholderStyle = styles.InputPlaceholderStyle
	in.Cursor.Style = styles.KeyStyle
	in.Focus()

	return JoinChatroomModal{
		apiClient: api,
		returnTo:  returnTo,
		input:     in,
	}
}

func (m JoinChatroomModal) Init() tea.Cmd { return textinput.Blink }

type joinDoneMsg struct{ err error }

// NewJoinChatroomModalWithID returns a prefilled modal with the given chatroom ID.
func NewJoinChatroomModalWithID(api *client.APIClient, returnTo tea.Model, id string) JoinChatroomModal {
	m := NewJoinChatroomModal(api, returnTo)
	m.input.SetValue(id)
	return m
}

func (m JoinChatroomModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		fieldWidth := m.width - 20
		if fieldWidth > 48 {
			fieldWidth = 48
		}
		if fieldWidth < 28 {
			fieldWidth = 28
		}
		m.input.Width = fieldWidth
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.submitting {
				return m, nil
			}
			return m.returnTo, nil
		case "enter":
			if m.submitting {
				return m, nil
			}
			ident := strings.TrimSpace(m.input.Value())
			if ident == "" {
				m.status = "Enter a chatroom ID"
				m.statusOkay = false
				return m, nil
			}
			if _, err := strconv.ParseUint(ident, 10, 64); err != nil {
				m.status = "Chatroom ID must be a number"
				m.statusOkay = false
				return m, nil
			}
			m.submitting = true
			m.status = "Joining..."
			m.statusOkay = true
			return m, joinByIDCmd(m.apiClient, ident)
		}
	case joinDoneMsg:
		m.submitting = false
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusOkay = false
			return m, nil
		}
		// On success, go back to previous screen (lists will refresh from cache invalidation)
		return m.returnTo, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m JoinChatroomModal) View() string {
	title := styles.CardTitleStyle.Render("Join Chatroom")
	subtitle := styles.CardSubtitleStyle.Render("Enter a chatroom ID to join")
	field := styles.InputFieldFocusedStyle.Render(m.input.View())

	statusView := ""
	if m.status != "" {
		if m.statusOkay {
			statusView = styles.StatusSuccessStyle.Render(m.status)
		} else {
			statusView = styles.StatusErrorStyle.Render(m.status)
		}
	}

	help := styles.HelpStyle.Render(strings.Join([]string{
		styles.RenderKeyBinding("Enter", "Join"),
		styles.RenderKeyBinding("Esc", "Cancel"),
	}, styles.HelpStyle.Render("  ")))

	content := strings.Join([]string{title, subtitle, field, statusView, help}, "\n\n")
	card := styles.CardStyle.Render(content)
	if m.width > 0 && m.height > 0 {
		centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
		return styles.AppStyle.Copy().Width(m.width).Height(m.height).Render(centered)
	}
	return styles.AppStyle.Render(card)
}

func joinByIDCmd(api *client.APIClient, idStr string) tea.Cmd {
	return func() tea.Msg {
		// Convert to uint and call client join
		id64, _ := strconv.ParseUint(idStr, 10, 64)
		err := api.JoinChatroom(uint(id64))
		if err != nil {
			return joinDoneMsg{err: err}
		}
		return joinDoneMsg{}
	}
}
