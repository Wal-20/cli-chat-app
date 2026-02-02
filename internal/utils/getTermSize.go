package utils

import (
	"golang.org/x/term"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func GetSizeCmd() tea.Cmd {
	return func() tea.Msg {
		w, h, _ := term.GetSize(int(os.Stdout.Fd()))
		return tea.WindowSizeMsg{Width: w, Height: h}
	}
}
