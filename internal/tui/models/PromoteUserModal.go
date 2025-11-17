package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PromoteUserModal promotes a user in the current chatroom to admin.
type PromoteUserModal struct {
	apiClient  *client.APIClient
	chatroomID uint
	returnTo   tea.Model

	input      textinput.Model
	submitting bool
	width      int
	height     int
	status     string
	statusOK   bool
}

func NewPromoteUserModal(api *client.APIClient, chatroomID uint, returnTo tea.Model) PromoteUserModal {
	in := textinput.New()
	in.Prompt = "> "
	in.Placeholder = "user id or name (e.g., 42 or alice)"
	in.PromptStyle = styles.InputPromptFocusedStyle
	in.TextStyle = styles.InputTextFocusedStyle
	in.PlaceholderStyle = styles.InputPlaceholderStyle
	in.Cursor.Style = styles.KeyStyle
	in.Focus()
	return PromoteUserModal{apiClient: api, chatroomID: chatroomID, returnTo: returnTo, input: in}
}

func (m PromoteUserModal) Init() tea.Cmd { return textinput.Blink }

type promoteDoneMsg struct{ err error }

func (m PromoteUserModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := m.width - 20
		if w > 48 {
			w = 48
		}
		if w < 28 {
			w = 28
		}
		m.input.Width = w
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
				m.status = "Enter a user id or name"
				m.statusOK = false
				return m, nil
			}
			m.submitting = true
			m.status = "Promoting..."
			m.statusOK = true
			return m, promoteUserCmd(m.apiClient, m.chatroomID, ident)
		}
	case promoteDoneMsg:
		m.submitting = false
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusOK = false
			return m, nil
		}
		return m.returnTo, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m PromoteUserModal) View() string {
	title := styles.CardTitleStyle.Render("Promote User to Admin")
	subtitle := styles.CardSubtitleStyle.Render("Grant admin privileges in this chatroom")
	field := styles.InputFieldFocusedStyle.Render(m.input.View())

	statusView := ""
	if m.status != "" {
		if m.statusOK {
			statusView = styles.StatusSuccessStyle.Render(m.status)
		} else {
			statusView = styles.StatusErrorStyle.Render(m.status)
		}
	}
	help := styles.HelpStyle.Render(strings.Join([]string{
		styles.RenderKeyBinding("Enter", "Promote user"),
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

func promoteUserCmd(api *client.APIClient, chatroomID uint, ident string) tea.Cmd {
	return func() tea.Msg {
		// If ident is not numeric, resolve username to id via membership list
		if _, err := strconv.ParseUint(ident, 10, 64); err != nil {
			users, uerr := api.GetUsersByChatroom(chatroomID, false)
			if uerr == nil {
				lowered := strings.ToLower(ident)
				for _, u := range users {
					if strings.ToLower(u.Name) == lowered {
						ident = fmt.Sprintf("%d", u.UserID)
						break
					}
				}
			}
		}
		err := api.MakeAdmin(fmt.Sprintf("%d", chatroomID), ident)
		return promoteDoneMsg{err: err}
	}
}
