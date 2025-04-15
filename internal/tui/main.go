package main

import (
	"github.com/Wal-20/cli-chat-app/internal/tui/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type mainModel struct {
	currentModel tea.Model
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.currentModel.Update(msg)
}

func (m mainModel) View() string {
	return m.currentModel.View()
}


func main() {
	apiClient, err := client.NewAPIClient()
	if err != nil {
		panic(err)
	}

	var currentModel tea.Model

	// Try to load token pair and set token
	if tokenPair, err := utils.LoadTokenPair(); err == nil && tokenPair.AccessToken != "" {

		apiClient.SetToken(tokenPair.AccessToken)
		tokenClaims, err := utils.GetClaimsFromToken(tokenPair.AccessToken)
		
		if err == nil {
			username, ok := tokenClaims["username"].(string)
			if ok {
				currentModel = models.NewMainChatModel(username, apiClient)
			}
		}
	}

	// Fall back to login if token is invalid or username retrieval fails
	if currentModel == nil {
		currentModel = models.NewLoginModel(apiClient)
	}

	program := tea.NewProgram(mainModel{currentModel: currentModel}, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		panic(err)
	}
}


