package models

import (
	"fmt"
	"io"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type chatroomItem struct {
	chatroom models.Chatroom
	isMember bool
}

func (i chatroomItem) Title() string       { return i.chatroom.Title }
func (i chatroomItem) FilterValue() string { return i.chatroom.Title }

type chatroomDelegate struct {
	styles list.DefaultItemStyles
}

func NewChatroomDelegate() chatroomDelegate {
	d := chatroomDelegate{}

	// Initialize with default styles
	d.styles = list.NewDefaultItemStyles()

	// Customize the styles
	d.styles.NormalTitle = d.styles.NormalTitle.
		Foreground(lipgloss.Color("241")).
		Padding(0, 0, 0, 2)

	d.styles.SelectedTitle = d.styles.SelectedTitle.
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("170"))

	d.styles.DimmedTitle = d.styles.DimmedTitle.
		Foreground(lipgloss.Color("239"))

	d.styles.FilterMatch = d.styles.FilterMatch.Foreground(styles.AquaColor.Value())

	return d
}

func (d chatroomDelegate) Height() int                               { return 1 }
func (d chatroomDelegate) Spacing() int                              { return 0 }
func (d chatroomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d chatroomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(chatroomItem)
	if !ok {
		return
	}

	label := item.chatroom.Title
	if !item.isMember {
		label += " (public)"
	}

	var rendered string
	if index == m.Index() {
		rendered = lipgloss.NewStyle().Margin(0, 0, 1, 0).Render("> " + label)
	} else {
		rendered = lipgloss.NewStyle().Margin(0, 0, 1, 0).Render("  " + label)
	}

	fmt.Fprint(w, rendered)
}

type MainChatModel struct {
	apiClient         *client.APIClient
	username          string
	userChatrooms     list.Model
	publicChatrooms   list.Model
	activeListPointer int
	width             int
	height            int
}

func NewMainChatModel(username string, apiClient *client.APIClient) MainChatModel {
	userChatroomsData, err := apiClient.GetUserChatrooms()
	if err != nil {
		panic(err)
	}
	publicChatroomsData, err := apiClient.GetChatrooms()
	if err != nil {
		panic(err)
	}

	userItems := make([]list.Item, len(userChatroomsData))
	for i, c := range userChatroomsData {
		userItems[i] = chatroomItem{chatroom: c, isMember: true}
	}
	publicItems := make([]list.Item, len(publicChatroomsData))
	for i, c := range publicChatroomsData {
		publicItems[i] = chatroomItem{chatroom: c, isMember: false}
	}

	delegate := NewChatroomDelegate()

	// Initialize lists with proper height
	userList := list.New(userItems, delegate, 20, 10) // Set initial size
	userList.Title = "Your Chatrooms"
	userList.SetShowHelp(false)
	userList.SetFilteringEnabled(true)
	userList.Styles.Title = styles.TitleStyle // Add custom styling
	userList.DisableQuitKeybindings()         // Disable default quit
	userList.SetShowPagination(true)

	publicList := list.New(publicItems, delegate, 20, 10)
	publicList.Title = "Public Chatrooms"
	publicList.SetShowHelp(false)
	publicList.SetFilteringEnabled(true)
	publicList.Styles.Title = styles.TitleStyle
	publicList.DisableQuitKeybindings()
	publicList.SetShowPagination(true)

	return MainChatModel{
		apiClient:         apiClient,
		username:          username,
		userChatrooms:     userList,
		publicChatrooms:   publicList,
		activeListPointer: 0,
	}
}

func (m MainChatModel) Init() tea.Cmd {
	return nil
}

func (m MainChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := m.width - 8
		listHeight := m.height - 6

		if listWidth < 0 {
			listWidth = 0
		}
		if listHeight < 3 {
			listHeight = 3
		}

		if m.activeListPointer == 0 {
			m.userChatrooms.SetSize(listWidth, listHeight)
		} else {
			m.publicChatrooms.SetSize(listWidth, listHeight)
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.activeListPointer = (m.activeListPointer + 1) % 2
		case "enter":
			if m.activeListPointer == 0 {
				if item, ok := m.userChatrooms.SelectedItem().(chatroomItem); ok {
					return NewChatroomModel(m.username, item.chatroom, m.apiClient), nil
				}
			} else {
				if item, ok := m.publicChatrooms.SelectedItem().(chatroomItem); ok {
					err := m.apiClient.JoinChatroom(item.chatroom.Id)
					if err != nil {
						return m, nil
					}
					return NewChatroomModel(m.username, item.chatroom, m.apiClient), nil
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "L":
			err := m.apiClient.Logout()
			if err != nil {
				return m, nil
			}
			return NewLoginModel(m.apiClient), nil
		}
	}

	if m.activeListPointer == 0 {
		m.userChatrooms, cmd = m.userChatrooms.Update(msg)
	} else {
		m.publicChatrooms, cmd = m.publicChatrooms.Update(msg)
	}

	return m, cmd
}

func (m MainChatModel) View() string {
	header := styles.TitleStyle.Render(fmt.Sprintf("Welcome %s!", m.username)) + "\n"
	helpText := styles.CommandStyle.Render("[Tab] Switch lists • [Enter] Join/View • [q] Quit • [L] Log Out") + "\n"

	var listView string
	if m.activeListPointer == 0 {
		listView = m.userChatrooms.View()
	} else {
		listView = m.publicChatrooms.View()
	}

	listView = styles.ContainerStyle.Render(listView)

	renderStyle := styles.ContainerStyle.MaxWidth(m.width).MaxHeight(m.height).
		Width(m.width).
		Height(m.height).Padding(0).Margin(0)

	return renderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, helpText, listView))
}
