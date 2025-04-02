
package models

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const maxVisibleMessages = 5 // Number of messages visible at a time

type ChatroomModel struct {
	apiClient   *client.APIClient
	username    string
	chatroom    models.Chatroom
	messages    []models.Message
	input       textinput.Model
	scrollIndex int // Tracks scrolling position
}

// NewChatroomModel initializes the chatroom model
func NewChatroomModel(username string, chatroom models.Chatroom, apiClient *client.APIClient) ChatroomModel {
	input := textinput.New()
	input.Placeholder = "Type a message..."
	input.Focus()
	input.CharLimit = 256
	input.Width = 50
	input.Cursor.Style = styles.CursorStyle
	input.Cursor.SetMode(cursor.CursorBlink)

	messages, err := apiClient.GetMessages(chatroom.Id)
	if err != nil {
		panic(err)
	}

	return ChatroomModel{
		apiClient:   apiClient,
		username:    username,
		chatroom:    chatroom,
		messages:    messages,
		input:       input,
		scrollIndex: 0, // Start at bottom (latest messages)
	}
}

// Init ensures the cursor blinks
func (m ChatroomModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keypress events
func (m ChatroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if inputValue := m.input.Value(); inputValue != "" {

				result, err := m.apiClient.SendMessage(strconv.FormatUint(uint64(m.chatroom.Id), 10), inputValue)

				if(err != nil) {
					panic(err)
				}

				message := result["Message"].(models.Message)

				m.messages = append(m.messages, message)
				m.input.SetValue("") // Clear input after sending message

				m.scrollIndex = len(m.messages) - maxVisibleMessages // Scroll to latest message
				if m.scrollIndex < 0 {
					m.scrollIndex = 0
				}
			}
		case "esc":
			return NewMainChatModel(m.username, m.apiClient), nil
		case "up":
			if m.scrollIndex > 0 {
				m.scrollIndex-- // Scroll up
			}
		case "down":
			if m.scrollIndex+maxVisibleMessages < len(m.messages) {
				m.scrollIndex++ // Scroll down
			}
		default:
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the chat UI
func (m ChatroomModel) View() string {
	var sb strings.Builder

	// Chatroom title
	sb.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Chatroom: %s", m.chatroom.Title)) + "\n\n")

	// Handle scrolling bounds
	if len(m.messages) > maxVisibleMessages {
		if m.scrollIndex < 0 {
			m.scrollIndex = 0
		} else if m.scrollIndex+maxVisibleMessages > len(m.messages) {
			m.scrollIndex = len(m.messages) - maxVisibleMessages
		}
	}

	// Display visible messages based on scrolling
	start := m.scrollIndex
	end := start + maxVisibleMessages

	end = min(end, len(m.messages))
	displayedMessages := m.messages[start:end]

	for _, message := range displayedMessages {
		sb.WriteString(styles.MessageStyle.Render(message.Content) + "\n")
	}

	// Scroll indicators
	if len(m.messages) > maxVisibleMessages {
		sb.WriteString("\n")
		if m.scrollIndex > 0 {
			sb.WriteString(styles.NavStyle.Render("[↑] "))
		} else {
			sb.WriteString("    ") // Align spacing
		}

		if m.scrollIndex+maxVisibleMessages < len(m.messages) {
			sb.WriteString(styles.NavStyle.Render("[↓]"))
		}
		sb.WriteString("\n")
	}

	// Input field
	sb.WriteString("\n" + styles.InputStyle.Render(m.input.View()))

	return styles.ContainerStyle.Render(sb.String())
}

