package models

import (
	"fmt"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginModel struct {
	apiClient *client.APIClient
	inputs []textinput.Model
	cursorMode cursor.Mode
	focusIndex int
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func NewLoginModel(apiClient *client.APIClient) LoginModel {

	username := textinput.New()
	username.Placeholder = "Username"
	username.Focus()

	password := textinput.New()
	password.Placeholder = "Password"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'

	return LoginModel{
		apiClient: apiClient,
		inputs: []textinput.Model{username, password},
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

		case tea.KeyMsg:

			switch msg.String() {

			case "ctrl+c", "q":
				return m, tea.Quit

			case "ctrl+r":
				m.cursorMode++
				if m.cursorMode > cursor.CursorHide {
					m.cursorMode = cursor.CursorBlink
				}
				cmds := make([]tea.Cmd, len(m.inputs))

				for i := range m.inputs {
					cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
				}
				
				return m, tea.Batch(cmds...) // returns the updated model and performs a bunch of commands

			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()
				if s == "enter" && m.focusIndex == len(m.inputs) {
					username := m.inputs[0].Value()
					password := m.inputs[1].Value()

					res, err := m.apiClient.LoginOrRegister(username, password)

					if err != nil {
						panic(err)
					}
					m.apiClient.SetToken(res["AccessToken"].(string))

					return NewMainChatModel(username, m.apiClient), nil
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

				// checking for selected input, updating it's style
				cmds := make([]tea.Cmd, len(m.inputs))
				for i := 0; i < len(m.inputs); i++ {
					if i == m.focusIndex {
						cmds[i] = m.inputs[i].Focus()
						m.inputs[i].PromptStyle = focusedStyle
						m.inputs[i].TextStyle = focusedStyle
						continue
					}
					// Remove focused state
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = noStyle
					m.inputs[i].TextStyle = noStyle
				}
				return m, tea.Batch(cmds...)
			}
			cmd := m.updateInputs(msg)
			return m, cmd
	}
	return m, nil
}

func (m *LoginModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m LoginModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs) - 1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return styles.ContainerStyle.Render(b.String())
}

