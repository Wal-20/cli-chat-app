package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Keep a simple, terminal-native look: no backgrounds and minimal borders.
var (
	// Palette (used for foregrounds only)
	primaryColor   = lipgloss.Color("#c084fc") // accent, not background
	secondaryColor = lipgloss.Color("#38bdf8")
	successColor   = lipgloss.Color("#34d399")
	dangerColor    = lipgloss.Color("#f87171")
	textMutedColor = lipgloss.Color("#94a3b8")
)

var (
	// Root app style: no padding/background so terminal shows through
	AppStyle = lipgloss.NewStyle()

	TitleStyle              = lipgloss.NewStyle().Bold(true)
	SubtitleStyle           = lipgloss.NewStyle().Foreground(textMutedColor)
	SectionTitleStyle       = lipgloss.NewStyle().Bold(true)
	SectionDescriptionStyle = lipgloss.NewStyle().Foreground(textMutedColor)
	MutedTextStyle          = lipgloss.NewStyle().Foreground(textMutedColor)
	EmphasisTextStyle       = lipgloss.NewStyle().Foreground(secondaryColor)
	HelpStyle               = lipgloss.NewStyle().Foreground(textMutedColor)
	KeyStyle                = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	// Status/footers with no backgrounds
	StatusBarStyle     = lipgloss.NewStyle()
	StatusMessageStyle = lipgloss.NewStyle().Foreground(textMutedColor)
	StatusInfoStyle    = StatusMessageStyle.Copy().Foreground(secondaryColor)
	StatusSuccessStyle = StatusMessageStyle.Copy().Foreground(successColor)
	StatusErrorStyle   = StatusMessageStyle.Copy().Foreground(dangerColor)

	// Card and pane wrappers simplified (no borders/backgrounds)
	CardStyle         = lipgloss.NewStyle().Width(56)
	CardTitleStyle    = lipgloss.NewStyle().Bold(true)
	CardSubtitleStyle = lipgloss.NewStyle().Foreground(textMutedColor)
	PaneStyle         = lipgloss.NewStyle()
	ActivePaneStyle   = PaneStyle.Copy().Bold(true)
	InactivePaneStyle = PaneStyle.Copy()
	PaneHeadingStyle  = lipgloss.NewStyle().Bold(true)

	// Lists
	ListItemTitleStyle         = lipgloss.NewStyle()
	ListItemTitleSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	ListItemMetaStyle          = lipgloss.NewStyle().Foreground(textMutedColor)

	// Conversation/message area
	ConversationWrapperStyle = lipgloss.NewStyle() // width will be applied by views
	MessageContainerStyle    = lipgloss.NewStyle().MarginBottom(1)
	MessageBubbleStyle       = lipgloss.NewStyle() // no border/background
	MessageBubbleSelfStyle   = MessageBubbleStyle.Copy().Foreground(primaryColor)
	MessageAuthorStyle       = lipgloss.NewStyle().Bold(true)
	MessageAuthorSelfStyle   = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	MessageTimestampStyle    = lipgloss.NewStyle().Foreground(textMutedColor)
	MessageContentStyle      = lipgloss.NewStyle()
	DateDividerStyle         = lipgloss.NewStyle().Foreground(textMutedColor).Align(lipgloss.Center)

	// Sidebar (members)
	SidebarStyle               = lipgloss.NewStyle().Width(30)
	SidebarTitleStyle          = lipgloss.NewStyle().Bold(true)
	ParticipantLineStyle       = lipgloss.NewStyle()
	ParticipantBadgeStyle      = lipgloss.NewStyle().Foreground(textMutedColor)
	ParticipantBadgeOwnerStyle = ParticipantBadgeStyle.Copy().Foreground(primaryColor)
	ParticipantBadgeAdminStyle = ParticipantBadgeStyle.Copy().Foreground(secondaryColor)
	ParticipantBadgeYouStyle   = ParticipantBadgeStyle.Copy().Foreground(successColor)

	// Inputs simplified: no borders/backgrounds
	InputAreaStyle          = lipgloss.NewStyle()
	InputFieldStyle         = lipgloss.NewStyle()
	InputFieldFocusedStyle  = lipgloss.NewStyle().Bold(true)
	InputPromptStyle        = lipgloss.NewStyle().Foreground(textMutedColor)
	InputPromptFocusedStyle = lipgloss.NewStyle().Foreground(primaryColor)
	InputTextStyle          = lipgloss.NewStyle()
	InputTextFocusedStyle   = lipgloss.NewStyle()
	InputPlaceholderStyle   = lipgloss.NewStyle().Foreground(textMutedColor)

	// Modals
	ModalInputStyle = lipgloss.NewStyle().Foreground(primaryColor).Align(lipgloss.Center)

	// Buttons simplified
	ButtonStyle        = lipgloss.NewStyle()
	ButtonFocusedStyle = lipgloss.NewStyle().Bold(true)

	CommandStyle = lipgloss.NewStyle().Foreground(textMutedColor)
)

func RenderButton(label string, focused bool) string {
	if focused {
		return ButtonFocusedStyle.Render(label)
	}
	return ButtonStyle.Render(label)
}

func RenderKeyBinding(key, description string) string {
	return fmt.Sprintf("%s %s", KeyStyle.Render(key), HelpStyle.Render(description))
}
