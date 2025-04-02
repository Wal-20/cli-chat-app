package main

import (
	"github.com/Wal-20/cli-chat-app/internal/tui/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
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

	if(err != nil) {
		panic(err)
	}

	model := mainModel{
		currentModel: models.NewLoginModel(apiClient),
	}

	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		panic(err)
	}
}

