package styles

import (
	"github.com/charmbracelet/lipgloss"
	"fmt"
)

type color interface {
	Value() lipgloss.Color
}

type ColorType struct {
	value lipgloss.Color
}

func (c ColorType) Value() lipgloss.Color {
	return c.value
}


var (

	PrimaryColor   = ColorType{lipgloss.Color("#7D56F4")}
	SecondaryColor = ColorType{lipgloss.Color("#874BFD")}
	AccentColor    = ColorType{lipgloss.Color("#FFFFFF")}
	MutedColor     = ColorType{lipgloss.Color("#4A4A4A")}

	RedColor = ColorType{lipgloss.Color("9")}
	MagentaColor = ColorType{lipgloss.Color("5")}
	AquaColor = ColorType{lipgloss.Color("86")}

	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginLeft(2)

	SectionStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1).
		Margin(1).
		Width(40)

	ActiveItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 1)

	InactiveItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4A4A4A"))

	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Margin(1, 2)

	ContainerStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(2).
		Margin(1).
		Width(50)

		SidebarStyle = lipgloss.NewStyle().
		Width(30).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	ChatroomStyle = lipgloss.NewStyle().
		Width(60).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)


	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF9900")).
				Bold(true)

	MessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Italic(true)

	NavStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF8800")) // Orange

	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle = FocusedStyle
	NoStyle = lipgloss.NewStyle()
	CursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	FocusedButton = FocusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", BlurredStyle.Render("Submit"))
)

