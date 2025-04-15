
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

const maxVisibleMessages = 5

type ChatroomModel struct {
	apiClient   *client.APIClient
	username    string
	chatroom    models.Chatroom
	messages    []models.Message
	input       textarea.Model
	scrollIndex int
}

func NewChatroomModel(username string, chatroom models.Chatroom, apiClient *client.APIClient) ChatroomModel {
	input := textarea.New()
	input.Placeholder = "Type a message..."
	input.Focus()
	input.SetWidth(80)
	input.SetHeight(6)
	input.CharLimit = 500

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
		scrollIndex: 0,
	}
}

func (m ChatroomModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if inputValue := strings.TrimSpace(m.input.Value()); inputValue != "" {
				result, err := m.apiClient.SendMessage(strconv.FormatUint(uint64(m.chatroom.Id), 10), inputValue)
				if err != nil {
					panic(err)
				}

				var message models.Message
				msgData, err := json.Marshal(result["Message"])
				if err != nil {
					panic(err)
				}

				if err := json.Unmarshal(msgData, &message); err != nil {
					panic(err)
				}

				m.messages = append(m.messages, message)
				m.input.Reset()

				m.scrollIndex = len(m.messages) - maxVisibleMessages
				if m.scrollIndex < 0 {
					m.scrollIndex = 0
				}
			}
		case "esc":
			return NewMainChatModel(m.username, m.apiClient), nil
		case "up":
			if m.scrollIndex > 0 {
				m.scrollIndex--
			}
		case "down":
			if m.scrollIndex+maxVisibleMessages < len(m.messages) {
				m.scrollIndex++
			}
		default:
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	default:
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m ChatroomModel) View() string {
	var sb strings.Builder

	sb.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Chatroom: %s", m.chatroom.Title)) + "\n\n")

	if len(m.messages) > maxVisibleMessages {
		if m.scrollIndex < 0 {
			m.scrollIndex = 0
		} else if m.scrollIndex+maxVisibleMessages > len(m.messages) {
			m.scrollIndex = len(m.messages) - maxVisibleMessages
		}
	}

	start := m.scrollIndex
	end := start + maxVisibleMessages
	end = min(len(m.messages), end)
	
	displayedMessages := m.messages[start:end]
	for _, message := range displayedMessages {
		sb.WriteString(styles.MessageStyle.Render(message.Content) + "\n")
	}

	if len(m.messages) > maxVisibleMessages {
		sb.WriteString("\n")
		if m.scrollIndex > 0 {
			sb.WriteString(styles.NavStyle.Render("[↑] "))
		} else {
			sb.WriteString("    ")
		}

		if m.scrollIndex+maxVisibleMessages < len(m.messages) {
			sb.WriteString(styles.NavStyle.Render("[↓]"))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n" + styles.InputStyle.Render(m.input.View()))
	sb.WriteString(styles.CommandStyle.Render("[Esc] Exit Chatroom • [Enter] Send Message") + "\n")
	return styles.ContainerStyle.Render(sb.String())
}