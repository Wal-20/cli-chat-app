package models

import (
	"fmt"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginModel struct {
	apiClient     *client.APIClient
	inputs        []textinput.Model
	cursorMode    cursor.Mode
	focusIndex    int
	width         int
	height        int
	submitting    bool
	statusMessage string
	statusStyle   lipgloss.Style
}

type loginResultMsg struct {
	username string
	token    string
	userID   uint
	err      error
}

func NewLoginModel(apiClient *client.APIClient) LoginModel {
	username := textinput.New()
	username.Placeholder = "Username"
	username.CharLimit = 48
	username.Prompt = "> "
	username.PromptStyle = styles.InputPromptFocusedStyle
	username.TextStyle = styles.InputTextFocusedStyle
	username.PlaceholderStyle = styles.InputPlaceholderStyle
	username.Cursor.Style = styles.KeyStyle
	username.Width = 36
	username.Focus()

	password := textinput.New()
	password.Placeholder = "Password"
	password.CharLimit = 64
	password.Prompt = "> "
	password.PromptStyle = styles.InputPromptStyle
	password.TextStyle = styles.InputTextStyle
	password.PlaceholderStyle = styles.InputPlaceholderStyle
	password.Cursor.Style = styles.KeyStyle
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '*'
	password.Width = 36

	return LoginModel{
		apiClient:     apiClient,
		inputs:        []textinput.Model{username, password},
		cursorMode:    cursor.CursorBlink,
		focusIndex:    0,
		statusMessage: "Enter your credentials to join or create chatrooms.",
		statusStyle:   styles.StatusMessageStyle,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		fieldWidth := msg.Width - 20
		if fieldWidth > 48 {
			fieldWidth = 48
		}
		if fieldWidth < 28 {
			fieldWidth = 28
		}
		for i := range m.inputs {
			m.inputs[i].Width = fieldWidth
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+r":
			if m.submitting {
				return m, nil
			}
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			m.statusMessage = fmt.Sprintf("Cursor mode: %s", m.cursorMode.String())
			m.statusStyle = styles.StatusInfoStyle
			return m, tea.Batch(cmds...)

		case "tab", "shift+tab", "enter", "up", "down":
			if m.submitting {
				return m, nil
			}

			s := msg.String()
			if s == "enter" && m.focusIndex == len(m.inputs) {
				username := strings.TrimSpace(m.inputs[0].Value())
				password := strings.TrimSpace(m.inputs[1].Value())
				if username == "" || password == "" {
					m.statusMessage = "Username and password are required."
					m.statusStyle = styles.StatusErrorStyle
					return m, nil
				}

				m.submitting = true
				m.statusMessage = "Authenticating..."
				m.statusStyle = styles.StatusInfoStyle
				return m, tea.Batch(m.applyFocusStyles(), login(m.apiClient, username, password))
			}

			if s == "tab" || s == "enter" || s == "down" {
				m.focusIndex++
			} else {
				m.focusIndex--
			}

			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			} else if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}

			return m, m.applyFocusStyles()
		}

	case loginResultMsg:
		m.submitting = false
		if msg.err != nil {
			m.statusMessage = msg.err.Error()
			m.statusStyle = styles.StatusErrorStyle
			m.focusIndex = 0
			m.inputs[1].Reset()
			return m, m.applyFocusStyles()
		}

		m.apiClient.SetToken(msg.token)
		return NewMainChatModel(msg.username, msg.userID, m.apiClient), nil
	}

	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m LoginModel) applyFocusStyles() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].PromptStyle = styles.InputPromptFocusedStyle
			m.inputs[i].TextStyle = styles.InputTextFocusedStyle
			m.inputs[i].PlaceholderStyle = styles.InputPlaceholderStyle
			if !m.inputs[i].Focused() {
				cmds[i] = m.inputs[i].Focus()
			}
			continue
		}

		m.inputs[i].PromptStyle = styles.InputPromptStyle
		m.inputs[i].TextStyle = styles.InputTextStyle
		m.inputs[i].PlaceholderStyle = styles.InputPlaceholderStyle
		if m.inputs[i].Focused() {
			m.inputs[i].Blur()
		}
	}

	return tea.Batch(cmds...)
}

func (m LoginModel) View() string {
	fields := make([]string, len(m.inputs))
	for i := range m.inputs {
		view := m.inputs[i].View()
		if i == m.focusIndex {
			fields[i] = styles.InputFieldFocusedStyle.Render(view)
		} else {
			fields[i] = styles.InputFieldStyle.Render(view)
		}
	}

	form := strings.Join(fields, "\n\n")
	button := styles.RenderButton("Sign in / Register", m.focusIndex == len(m.inputs))

	status := ""
	if m.statusMessage != "" {
		status = m.statusStyle.Render(m.statusMessage)
	}

	helpItems := []string{
		styles.RenderKeyBinding("Tab", "Next field"),
		styles.RenderKeyBinding("Shift+Tab", "Previous"),
		styles.RenderKeyBinding("Enter", "Submit"),
		styles.RenderKeyBinding("Ctrl+R", fmt.Sprintf("Cursor: %s", m.cursorMode.String())),
		styles.RenderKeyBinding("q", "Quit"),
	}
	help := strings.Join(helpItems, styles.HelpStyle.Render("  "))

	sections := []string{
		styles.CardTitleStyle.Render("CLI Chat"),
		styles.CardSubtitleStyle.Render("Your space to talk from the terminal."),
		form,
		button,
	}

	if status != "" {
		sections = append(sections, status)
	}

	sections = append(sections, styles.HelpStyle.Render(help))

	content := strings.Join(sections, "\n\n")
	card := styles.CardStyle.Render(content)

	if m.width > 0 && m.height > 0 {
		card = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			card,
		)
		return styles.AppStyle.Copy().Width(m.width).Height(m.height).Render(card)
	}

	return styles.AppStyle.Render(card)
}

func login(apiClient *client.APIClient, username, password string) tea.Cmd {
	return func() tea.Msg {
		if err := utils.ValidatePassword(password); err != nil {
			return loginResultMsg{err: err}
		}
		res, err := apiClient.LoginOrRegister(username, password)
		if err != nil {
			return loginResultMsg{err: err}
		}

		token, ok := res["AccessToken"].(string)
		if !ok || token == "" {
			return loginResultMsg{err: fmt.Errorf("authentication failed: missing access token")}
		}

		// Persist token pair locally and hydrate client for future requests
		refresh, _ := res["RefreshToken"].(string)
		_ = utils.SaveTokenPair(utils.TokenPair{AccessToken: token, RefreshToken: refresh})
		apiClient.SetTokenPair(token, refresh)
		// Extract user ID from access token claims
		var uid uint
		if claims, err2 := utils.GetClaimsFromToken(token); err2 == nil {
			if idf, ok := claims["userID"].(float64); ok {
				uid = uint(idf)
			}
		}
		return loginResultMsg{username: username, token: token, userID: uid}
	}
}
