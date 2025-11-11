package models

import (
	"fmt"
	"io"
	"strings"

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

type chatroomDelegate struct{}

func NewChatroomDelegate() chatroomDelegate {
	return chatroomDelegate{}
}

func (d chatroomDelegate) Height() int                               { return 3 }
func (d chatroomDelegate) Spacing() int                              { return 1 }
func (d chatroomDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d chatroomDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(chatroomItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	titleStyle := styles.ListItemTitleStyle
	if isSelected {
		titleStyle = styles.ListItemTitleSelectedStyle
	}

	title := titleStyle.Render(item.chatroom.Title)

	metaParts := []string{}
	if item.isMember {
		metaParts = append(metaParts, "joined")
	} else if item.chatroom.IsPublic {
		metaParts = append(metaParts, "public")
	} else {
		metaParts = append(metaParts, "private")
	}
	metaParts = append(metaParts, fmt.Sprintf("capacity %d", item.chatroom.MaxUserCount))
	meta := styles.ListItemMetaStyle.Render(strings.Join(metaParts, " | "))

	pointer := "  "
	if isSelected {
		pointer = styles.KeyStyle.Render("> ")
	}

	titleLine := pointer + title
	metaLine := "    " + meta

	fmt.Fprintf(w, "%s\n%s", titleLine, metaLine)
}

type MainChatModel struct {
	apiClient       *client.APIClient
	userID          uint
	username        string
	userChatrooms   list.Model
	publicChatrooms list.Model
	activeList      int
	width           int
	height          int
	flashMessage    string
	flashStyle      lipgloss.Style
}

func NewMainChatModel(username string, userID uint, apiClient *client.APIClient) MainChatModel {
	userItems := []list.Item{}
	publicItems := []list.Item{}
	loadErrors := []string{}

	if userChatroomsData, err := apiClient.GetUserChatrooms(); err != nil {
		loadErrors = append(loadErrors, fmt.Sprintf("Your chatrooms unavailable: %s", err.Error()))
	} else {
		userItems = make([]list.Item, len(userChatroomsData))
		for i, c := range userChatroomsData {
			userItems[i] = chatroomItem{chatroom: c, isMember: true}
		}
	}

	if publicChatroomsData, err := apiClient.GetChatrooms(); err != nil {
		loadErrors = append(loadErrors, fmt.Sprintf("Discover feed unavailable: %s", err.Error()))
	} else {
		publicItems = make([]list.Item, len(publicChatroomsData))
		for i, c := range publicChatroomsData {
			publicItems[i] = chatroomItem{chatroom: c, isMember: false}
		}
	}

	delegate := NewChatroomDelegate()

	userList := list.New(userItems, delegate, 40, 12)
	userList.SetShowHelp(false)
	userList.SetShowTitle(false)
	userList.SetShowStatusBar(false)
	userList.SetShowPagination(false)
	userList.SetFilteringEnabled(true)
	userList.DisableQuitKeybindings()

	publicList := list.New(publicItems, delegate, 40, 12)
	publicList.SetShowHelp(false)
	publicList.SetShowTitle(false)
	publicList.SetShowStatusBar(false)
	publicList.SetShowPagination(false)
	publicList.SetFilteringEnabled(true)
	publicList.DisableQuitKeybindings()

	flashMessage := ""
	flashStyle := styles.StatusInfoStyle
	if len(loadErrors) > 0 {
		flashMessage = strings.Join(loadErrors, " | ")
		flashStyle = styles.StatusErrorStyle
	}

	return MainChatModel{
		apiClient:       apiClient,
		userID:          userID,
		username:        username,
		userChatrooms:   userList,
		publicChatrooms: publicList,
		activeList:      0,
		flashMessage:    flashMessage,
		flashStyle:      flashStyle,
	}
}

func (m MainChatModel) Init() tea.Cmd {
	return nil
}

func (m MainChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		paneWidth := m.paneWidth()
		listWidth := paneWidth - 4
		if listWidth < 16 {
			listWidth = 16
		}

		listHeight := msg.Height - 10
		if listHeight < 8 {
			listHeight = 8
		}

		m.userChatrooms.SetSize(listWidth, listHeight)
		m.publicChatrooms.SetSize(listWidth, listHeight)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.activeList = (m.activeList + 1) % 2
			return m, nil
		case "c":
			// create new chatroom
			return NewCreateChatroomModel(m.username, m.userID, m.apiClient), nil
		case "ctrl+j":
			// open modal to join by ID
			jm := NewJoinChatroomModal(m.apiClient, m)
			return jm, jm.Init()
		case "n":
			nm := NewNotificationsModel(m.username, m.userID, m.apiClient)
			return nm, loadNotifications(m.apiClient)
		case "enter":
			if m.activeList == 0 {
				if item, ok := m.userChatrooms.SelectedItem().(chatroomItem); ok {
					return NewChatroomModel(m.username, m.userID, item.chatroom, m.apiClient), nil
				}
			} else {
				if item, ok := m.publicChatrooms.SelectedItem().(chatroomItem); ok {
					if err := m.apiClient.JoinChatroom(item.chatroom.Id); err != nil {
						m.flashMessage = fmt.Sprintf("Could not join %s: %s", item.chatroom.Title, err.Error())
						m.flashStyle = styles.StatusErrorStyle
						return m, nil
					}
					return NewChatroomModel(m.username, m.userID, item.chatroom, m.apiClient), nil
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "L":
			if err := m.apiClient.Logout(); err != nil {
				m.flashMessage = fmt.Sprintf("Logout failed: %s", err.Error())
				m.flashStyle = styles.StatusErrorStyle
				return m, nil
			}
			return NewLoginModel(m.apiClient), nil
		}
	}

	if m.activeList == 0 {
		var cmd tea.Cmd
		m.userChatrooms, cmd = m.userChatrooms.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.publicChatrooms, cmd = m.publicChatrooms.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainChatModel) View() string {
	header := styles.TitleStyle.Render(fmt.Sprintf("Welcome, %s", m.username))
	subtitle := styles.SubtitleStyle.Render("Use Tab to switch panes, Enter to dive into a room.")

	paneWidth := m.paneWidth()
	leftPane := renderPane("Your chatrooms", m.userChatrooms, m.activeList == 0, paneWidth, true)
	rightPane := renderPane("Discover", m.publicChatrooms, m.activeList == 1, paneWidth, false)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	joinedCount := len(m.userChatrooms.VisibleItems())
	discoverCount := len(m.publicChatrooms.VisibleItems())
	info := fmt.Sprintf("%d joined | %d discoverable", joinedCount, discoverCount)
	statusStyle := styles.StatusInfoStyle
	if m.flashMessage != "" {
		info = m.flashMessage
		statusStyle = m.flashStyle
	}

	helpItems := []string{
		styles.RenderKeyBinding("Tab", "Switch pane"),
		styles.RenderKeyBinding("Enter", "Open or join"),
		styles.RenderKeyBinding("Ctrl+J", "Join by ID"),
		styles.RenderKeyBinding("L", "Log out"),
		styles.RenderKeyBinding("n", "Notifications"),
		styles.RenderKeyBinding("q", "Quit"),
		styles.RenderKeyBinding("c", "Create a chatroom"),
	}
	help := strings.Join(helpItems, styles.HelpStyle.Render("  "))

	footerContent := statusStyle.Render(info) + "\n" + styles.HelpStyle.Render(help)
	footer := styles.StatusBarStyle.Render(footerContent)

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		subtitle,
		"",
		columns,
		"",
		footer,
	)

	if m.width > 0 && m.height > 0 {
		app := styles.AppStyle.Copy().Width(m.width).Height(m.height)
		return app.Render(layout)
	}

	return styles.AppStyle.Render(layout)
}

func (m MainChatModel) paneWidth() int {
	if m.width <= 0 {
		return 48
	}
	paneWidth := (m.width - 8) / 2
	if paneWidth < 28 {
		paneWidth = 28
	}
	return paneWidth
}

func renderPane(title string, lst list.Model, active bool, width int, addMargin bool) string {
	header := styles.PaneHeadingStyle.Render(title)
	body := lipgloss.JoinVertical(lipgloss.Left, header, lst.View())

	paneStyle := styles.InactivePaneStyle
	if active {
		paneStyle = styles.ActivePaneStyle
	}

	if width > 0 {
		paneStyle = paneStyle.Copy().Width(width)
	}

	if addMargin {
		paneStyle = paneStyle.Copy().MarginRight(2)
	}

	return paneStyle.Render(body)
}
