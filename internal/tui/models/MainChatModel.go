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
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("170"))

	d.styles.DimmedTitle = d.styles.DimmedTitle.
		Foreground(lipgloss.Color("239"))
	
	d.styles.FilterMatch = d.styles.FilterMatch.Foreground(styles.AquaColor.Value())

	return d
}

func (d chatroomDelegate) Height() int { return 1 }
func (d chatroomDelegate) Spacing() int { return 0 }
func (d chatroomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d chatroomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(chatroomItem)
	if !ok {
		return
	}

	// Build the title string
	title := item.chatroom.Title
	if !item.isMember {
		title += " (public)"
	}

	var renderedTitle string
	if index == m.Index() {
		renderedTitle = d.styles.SelectedTitle.Render(title)
	} else {
		if m.FilterState() == list.Filtering {
			renderedTitle = d.styles.DimmedTitle.Render(title)
		} else {
			renderedTitle = d.styles.NormalTitle.Render(title)
		}
	}

	fmt.Fprintf(w, "%s\n", renderedTitle)
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
	userList.DisableQuitKeybindings() // Disable default quit    
	userList.SetShowPagination(true)
     
	publicList := list.New(publicItems, delegate, 20, 10) // Set initial size
	publicList.Title = "Public Chatrooms"
	publicList.SetShowHelp(false)
	publicList.SetFilteringEnabled(true)
	publicList.Styles.Title = styles.TitleStyle // Add custom styling
	publicList.DisableQuitKeybindings()         // Disable default quit
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
		m.width, m.height = msg.Width, msg.Height
		h, v := styles.ContainerStyle.GetFrameSize()

		// Calculate dimensions for side-by-side lists
		headerHeight := 4
		listHeight := m.height - v - headerHeight
		listWidth := (m.width-h)/2 - 2 // -2 for padding between lists

		// Set dimensions for both lists
		m.userChatrooms.SetSize(listWidth, listHeight)
		m.publicChatrooms.SetSize(listWidth, listHeight)

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
					err := m.apiClient.JoinChatroom(fmt.Sprint(item.chatroom.Id))
					if err != nil {
						return m, nil
					}
					return NewChatroomModel(m.username, item.chatroom, m.apiClient), nil
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
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
	helpText := styles.CommandStyle.Render("[Tab] Switch lists • [/] Filter • [Esc] Clear filter • [Enter] Join/View • [q] Quit") + "\n"

	// Create horizontal layout
	userListView := m.userChatrooms.View()
	publicListView := m.publicChatrooms.View()

	// Add highlighting for active list
	if m.activeListPointer == 0 {
		userListView = styles.ContainerStyle.Render(userListView)
		publicListView = styles.InactiveItemStyle.Render(publicListView)
	} else {
		userListView = styles.InactiveItemStyle.Render(userListView)
		publicListView = styles.ContainerStyle.Render(publicListView)
	}

	// Combine lists horizontally
	lists := lipgloss.JoinHorizontal(
		lipgloss.Top,
		userListView,
		"   ", // Add spacing between lists
		publicListView,
	)

	return styles.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			helpText,
			lists,
		),
	)
}
