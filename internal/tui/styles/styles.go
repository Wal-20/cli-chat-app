package styles

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
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
	PrimaryColor   = ColorType{lipgloss.Color("#D12182")}
	SecondaryColor = ColorType{lipgloss.Color("#874BFD")}
	AccentColor    = ColorType{lipgloss.Color("#FFFFFF")}
	MutedColor     = ColorType{lipgloss.Color("#4A4A4A")}

	RedColor     = ColorType{lipgloss.Color("9")}
	MagentaColor = ColorType{lipgloss.Color("5")}
	AquaColor    = ColorType{lipgloss.Color("86")}
	LimeColor = ColorType{lipgloss.Color("#00FF77")}

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor.Value())

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
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(MutedColor.value)).
			Padding(2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Margin(1, 2)

	ContainerStyle = lipgloss.NewStyle().
			Padding(2).
			Margin(2, 0, 2, 2)

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
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor.value)).
			Padding(2)



			MessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1).
			PaddingLeft(2)


			UsernameStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor.Value()).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(PrimaryColor.Value()).
			BorderLeft(true).
			PaddingLeft(1)


			OwnerStyle = lipgloss.NewStyle().
			Foreground(LimeColor.Value()).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(LimeColor.Value()).
			BorderLeft(true).
			PaddingLeft(1)


			AdminStyle = lipgloss.NewStyle().
			Foreground(AquaColor.Value()).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(AquaColor.Value()).
			BorderLeft(true).
			PaddingLeft(1)


	CommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			MarginTop(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Width(40).MarginTop(1)

	NavStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF8800")) // Orange

	FocusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle         = FocusedStyle
	NoStyle             = lipgloss.NewStyle()
	CursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	FocusedButton = FocusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", BlurredStyle.Render("Submit"))

	RightAlign = lipgloss.NewStyle().Align(lipgloss.Right)
	TimestampStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Italic(true)

	DateSeparatorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#71717a")).
	Align(lipgloss.Center)

	PaginationStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	PaddingLeft(2)
)
