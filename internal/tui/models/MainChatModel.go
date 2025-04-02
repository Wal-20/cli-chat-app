package models

import (
	"fmt"
	"strings"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type MainChatModel struct {
	apiClient    *client.APIClient
	username     string
	chatrooms    []models.Chatroom
	selectedIdx  int
}

func NewMainChatModel(username string, apiClient *client.APIClient) MainChatModel {
	chatrooms, err := apiClient.GetChatrooms()
	if err != nil {
		panic(err)		
	}

	return MainChatModel{
		apiClient:   apiClient,
		username:    username,
		chatrooms:   chatrooms,
		selectedIdx: 0,
	}
}

func (m MainChatModel) Init() tea.Cmd {
	return nil
}

func (m MainChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case "down", "j":
			if m.selectedIdx < len(m.chatrooms) - 1 {
				m.selectedIdx++
			}

		case "enter":
			if(len(m.chatrooms) == 0) {
				return m, nil
			}
			selectedChatroom := m.chatrooms[m.selectedIdx]
			return NewChatroomModel(m.username, selectedChatroom, m.apiClient), nil

		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m MainChatModel) View() string {
	var sb strings.Builder

	sb.WriteString(styles.TitleStyle.Render(fmt.Sprintf("Welcome %s", m.username)) + "\n\n")
	sb.WriteString("Your Chatrooms\n")
	sb.WriteString(strings.Repeat("-", 20) + "\n")

	for i, chatroom := range m.chatrooms {

		if i == m.selectedIdx {
			sb.WriteString(styles.SelectedItemStyle.Render("> " + chatroom.Title) + "\n")
		} else {
			sb.WriteString("  " + chatroom.Title + "\n")
		}
	}

	sb.WriteString("\n[Up/Down] Navigate | [Enter] Select | [q] Quit")

	return styles.ContainerStyle.Render(sb.String())
}

