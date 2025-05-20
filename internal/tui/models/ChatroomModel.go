
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

const maxVisibleMessages = 3

type ChatroomModel struct {
	apiClient   *client.APIClient
	username    string
	chatroom    models.Chatroom
	messages    []models.MessageWithUser
	users       []models.UserChatroom
	input       textarea.Model
	scrollIndex int
}

func NewChatroomModel(username string, chatroom models.Chatroom, apiClient *client.APIClient) ChatroomModel {
	input := textarea.New()
	input.Placeholder = "Type a message..."
	input.Focus()
	input.SetWidth(80)
	input.SetHeight(4)
	input.CharLimit = 500

	messages, err := apiClient.GetMessages(chatroom.Id)
	if err != nil {
		panic(err)
	}

	users, err := apiClient.GetUsersByChatroom(chatroom.Id)
	if err != nil {
		panic(err)
	}

	return ChatroomModel{
		apiClient:   apiClient,
		username:    username,
		chatroom:    chatroom,
		users:       users,
		messages:    messages,
		input:       input,
		scrollIndex: max(0, len(messages) - maxVisibleMessages),
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
			inputValue := strings.TrimSpace(m.input.Value())
			if inputValue != "" {
				result, err := m.apiClient.SendMessage(strconv.FormatUint(uint64(m.chatroom.Id), 10), inputValue)
				if err != nil {
					panic(err)
				}

				var message models.MessageWithUser
				msgData, _ := json.Marshal(result["Message"])
				_ = json.Unmarshal(msgData, &message)

				message.Username = m.username // populate it manually since API doesn't return it

				m.messages = append(m.messages, message)
				m.input.Reset()

				m.scrollIndex = max(0, len(m.messages)-maxVisibleMessages)
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

	if m.scrollIndex < 0 {
		m.scrollIndex = 0
	} else if m.scrollIndex+maxVisibleMessages > len(m.messages) {
		m.scrollIndex = max(0, len(m.messages)-maxVisibleMessages)
	}

	start := m.scrollIndex
	end := min(len(m.messages), start+maxVisibleMessages)
	displayedMessages := m.messages[start:end]

	var prevDate string
	for _, message := range displayedMessages {
		msgDate := message.CreatedAt.Format("2006-01-02")
		if msgDate != prevDate {
			sb.WriteString(styles.DateSeparatorStyle.Render(formatDateSeparator(message.CreatedAt)) + "\n\n")
			prevDate = msgDate
		}

		role := getUserRole(m.users, message.Username, m.username)
		timestamp := message.CreatedAt.Format("15:04")

		var msgBlock strings.Builder
		roleStyle := styles.UsernameStyle

		switch role {
		case "Owner":
			roleStyle = styles.OwnerStyle
		case "Admin":
			roleStyle = styles.AdminStyle
		}

		msgBlock.WriteString(roleStyle.Render(message.Username) + "\n")
		msgBlock.WriteString(styles.MessageStyle.Render(message.Content) + "\n")
		msgBlock.WriteString(styles.TimestampStyle.Render(timestamp) + "\n")

		sb.WriteString(msgBlock.String() + "\n")
	}

	// Pagination Controls + Page Indicator
	if len(m.messages) > maxVisibleMessages {
		sb.WriteString("\n")

		// Navigation Arrows
		if m.scrollIndex > 0 {
			sb.WriteString(styles.NavStyle.Render("[↑] "))
		} else {
			sb.WriteString("  ")
		}
		if m.scrollIndex+maxVisibleMessages < len(m.messages) {
			sb.WriteString(styles.NavStyle.Render(" [↓]"))
		}

	}

	sb.WriteString("\n" + styles.InputStyle.Render(m.input.View()))
	sb.WriteString(styles.CommandStyle.Render("[Esc] Exit Chatroom • [Enter] Send Message") + "\n")

	return styles.ContainerStyle.Render(sb.String())
}

func getUserRole(users []models.UserChatroom, username string, currentUsername string) string {
	if username == currentUsername {
		return "You"
	}
	for _, u := range users {
		if u.IsOwner {
			return "Owner"
		}
		if u.IsAdmin {
			return "Admin"
		}
	}
	return "User"
}

func formatDateSeparator(t time.Time) string {
	today := time.Now().Format("2006-01-02")
	msgDate := t.Format("2006-01-02")
	if msgDate == today {
		return "-- Today --"
	}
	return "-- " + t.Format("Mon Jan-02") + " --"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

