package main

import (
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/models"
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
		down := models.NewServerDownModel()
		program := tea.NewProgram(mainModel{currentModel: down}, tea.WithAltScreen())
		_, _ = program.Run()
		return
	}

	var currentModel tea.Model

	if tokenPair, err := utils.LoadTokenPair(); err == nil && tokenPair.AccessToken != "" {

		apiClient.SetTokenPair(tokenPair.AccessToken, tokenPair.RefreshToken)
		tokenClaims, err := utils.GetClaimsFromToken(tokenPair.AccessToken)

		if err == nil {
			username, ok := tokenClaims["username"].(string)
			if ok {
				var uid uint
				if idf, ok := tokenClaims["userID"].(float64); ok {
					uid = uint(idf)
				}
				currentModel = models.NewMainChatModel(username, uid, apiClient)
			}
		}
	}

	if currentModel == nil {
		currentModel = models.NewLoginModel(apiClient)
	}

	program := tea.NewProgram(mainModel{currentModel: currentModel}, tea.WithAltScreen())
	_, _ = program.Run()
}
