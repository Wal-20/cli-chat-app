package models

import (
	"fmt"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InviteUserModal is a focused model that prompts for a username or ID
// and submits an invite for the current chatroom.
type InviteUserModal struct {
	apiClient  *client.APIClient
	chatroomID uint
	returnTo   tea.Model

	input      textinput.Model
	submitting bool
	width      int
	height     int
	status     string
	statusOkay bool
}

func NewInviteUserModal(api *client.APIClient, chatroomID uint, returnTo tea.Model) InviteUserModal {
	in := textinput.New()
	in.Prompt = "> "
	in.Placeholder = "username or id (e.g., alice or 42)"
	in.PromptStyle = styles.InputPromptFocusedStyle
	in.TextStyle = styles.InputTextFocusedStyle
	in.PlaceholderStyle = styles.InputPlaceholderStyle
	in.Cursor.Style = styles.KeyStyle
	in.Focus()

	return InviteUserModal{
		apiClient:  api,
		chatroomID: chatroomID,
		returnTo:   returnTo,
		input:      in,
	}
}

func (m InviteUserModal) Init() tea.Cmd { return textinput.Blink }

// internal message types
type inviteDoneMsg struct{ err error }

func (m InviteUserModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// adjust width of the field to fit the card nicely
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
				m.status = "Enter a username or numeric ID"
				m.statusOkay = false
				return m, nil
			}
			m.submitting = true
			m.status = "Sending invite..."
			m.statusOkay = true
			return m, sendInviteCmd(m.apiClient, m.chatroomID, ident)
		}
	case inviteDoneMsg:
		m.submitting = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Invite failed: %s", msg.err.Error())
			m.statusOkay = false
			return m, nil
		}
		// on success, return to the chatroom
		return m.returnTo, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m InviteUserModal) View() string {
	// Build card content
	title := styles.CardTitleStyle.Render("Invite User")
	subtitle := styles.CardSubtitleStyle.Render("Enter a username or numeric ID")
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
		styles.RenderKeyBinding("Enter", "Send invite"),
		styles.RenderKeyBinding("Esc", "Cancel"),
	}, styles.HelpStyle.Render("  ")))

	content := strings.Join([]string{title, subtitle, field, statusView, help}, "\n\n")
	card := styles.CardStyle.Render(content)

	// center on screen
	if m.width > 0 && m.height > 0 {
		centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
		return styles.AppStyle.Copy().Width(m.width).Height(m.height).Render(centered)
	}
	return styles.ModalInputStyle.Render(card)
}

// command
func sendInviteCmd(api *client.APIClient, chatroomID uint, ident string) tea.Cmd {
	return func() tea.Msg {
		err := api.InviteUser(fmt.Sprintf("%d", chatroomID), ident)
		return inviteDoneMsg{err: err}
	}
}
