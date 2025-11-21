package models

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	appmodels "github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NotificationsModel struct {
	apiClient     *client.APIClient
	username      string
	userID        uint
	notifications list.Model
	width         int
	height        int
	flashMessage  string
	flashStyle    lipgloss.Style
	loading       bool
	joining       bool
}

func NewNotificationsModel(username string, userID uint, apiClient *client.APIClient) NotificationsModel {
	alertDelegate := newAlertDelegate()
	alertList := list.New([]list.Item{}, alertDelegate, 80, 18)
	alertList.SetShowHelp(false)
	alertList.SetShowTitle(false)
	alertList.SetShowStatusBar(false)
	alertList.SetShowPagination(false)
	alertList.DisableQuitKeybindings()

	return NotificationsModel{
		apiClient:     apiClient,
		username:      username,
		userID:        userID,
		notifications: alertList,
		flashMessage:  "Loading notifications...",
		flashStyle:    styles.StatusInfoStyle,
		loading:       true,
	}
}

func (m NotificationsModel) Init() tea.Cmd {
	return loadNotifications(m.apiClient)
}

func (m NotificationsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := m.width - 8
		if listWidth < 16 {
			listWidth = 16
		}

		listHeight := msg.Height - 10
		if listHeight < 8 {
			listHeight = 8
		}

		m.notifications.SetSize(listWidth, listHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if m.loading {
				return m, nil
			}
			m.loading = true
			m.flashMessage = "Refreshing..."
			m.flashStyle = styles.StatusInfoStyle
			return m, loadNotifications(m.apiClient)

		case "d":
			return m.deleteNotification()
		case "enter":
			// If selected notification is an invite, join directly
			if it, ok := m.notifications.SelectedItem().(alertItem); ok {
				if strings.EqualFold(it.notification.Type, "invite") {
					if m.joining {
						return m, nil
					}
					m.joining = true
					m.loading = true
					m.flashMessage = "Joining chatroom..."
					m.flashStyle = styles.StatusInfoStyle
					nid := uint(it.notification.Id)
					return m, joinInviteCmd(m.apiClient, it.notification.ChatroomId, nid)
				}
			}
			return m, nil
		case "esc", "q":
			return NewMainChatModel(m.username, m.userID, m.apiClient), nil
		}

	case notificationsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.flashMessage = msg.err.Error()
			m.flashStyle = styles.StatusErrorStyle
			return m, nil
		}
		m.notifications.SetItems(buildAlertItems(msg.resp.Notifications))
		m.flashMessage = fmt.Sprintf("%d notifications", len(msg.resp.Notifications))
		m.flashStyle = styles.StatusInfoStyle
		if len(msg.resp.Notifications) == 0 {
			m.flashMessage = "You're all caught up."
		}
		return m, nil
	case inviteJoinedMsg:
		m.joining = false
		m.loading = false
		if msg.err != nil {
			m.flashMessage = msg.err.Error()
			m.flashStyle = styles.StatusErrorStyle
			return m, nil
		}
		// Remove the notification from the UI list immediately (best-effort)
		if idx := m.findNotificationIndex(msg.notiID); idx >= 0 {
			m.notifications.RemoveItem(idx)
		}
		// Navigate into the joined chatroom if we could parse it; otherwise stay
		if msg.room.Id != 0 {
			cm := NewChatroomModel(m.username, m.userID, msg.room, m.apiClient)
			return cm, cm.Init()
		}
		m.flashMessage = "Joined"
		m.flashStyle = styles.StatusSuccessStyle
		return m, nil
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.notifications, cmd = m.notifications.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m NotificationsModel) View() string {
	header := styles.TitleStyle.Render("Notifications")
	subtitle := styles.SubtitleStyle.Render("Manage chatroom invites and recent alerts.")

	columns := renderPane("Activity", m.notifications, true, m.width-6, false)

	status := m.flashMessage
	statusStyle := m.flashStyle
	if status == "" {
		status = fmt.Sprintf("%d notifications", len(m.notifications.Items()))
		statusStyle = styles.StatusInfoStyle
	}

	helpItems := []string{
		styles.RenderKeyBinding("Enter", "Join invite"),
		styles.RenderKeyBinding("r", "Refresh"),
		styles.RenderKeyBinding("d", "Delete"),
		styles.RenderKeyBinding("Esc", "Back"),
	}
	help := strings.Join(helpItems, styles.HelpStyle.Render("  "))

	footerContent := statusStyle.Render(status) + "\n" + styles.HelpStyle.Render(help)
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

func (m NotificationsModel) deleteNotification() (tea.Model, tea.Cmd) {
	if n, ok := m.notifications.SelectedItem().(alertItem); ok {
		m.loading = true
		m.flashMessage = "Deleting notification..."
		m.flashStyle = styles.StatusInfoStyle

		id := uint(n.notification.Id)
		if err := m.apiClient.DeleteNotification(id); err != nil {
			m.flashMessage = fmt.Sprintf("Error deleting notification: %v", err)
			m.flashStyle = styles.StatusErrorStyle
			m.loading = false
			return m, nil
		}

		idx := m.notifications.Index()
		m.notifications.RemoveItem(idx)

		m.flashMessage = fmt.Sprintf("Notification %d deleted", id)
		m.flashStyle = styles.StatusSuccessStyle
		m.loading = false
	}
	return m, nil
}

func (m NotificationsModel) paneWidth() int {
	if m.width <= 0 {
		return 48
	}
	paneWidth := (m.width - 8) / 2
	if paneWidth < 28 {
		paneWidth = 28
	}
	return paneWidth
}

type notificationsLoadedMsg struct {
	resp appmodels.NotificationsResponse
	err  error
}

func loadNotifications(apiClient *client.APIClient) tea.Cmd {
	return func() tea.Msg {
		resp, err := apiClient.GetNotifications()
		return notificationsLoadedMsg{resp: resp, err: err}
	}
}

// Direct-join from an invite and return result
// moved below to avoid duplicates

// messages and command for direct-join from invite
type inviteJoinedMsg struct {
	room   appmodels.Chatroom
	err    error
	notiID uint
}

func joinInviteCmd(apiClient *client.APIClient, chatroomID uint, notificationID uint) tea.Cmd {
	return func() tea.Msg {
		res, err := apiClient.JoinChatroomVerbose(chatroomID)
		if err != nil {
			return inviteJoinedMsg{err: err, notiID: notificationID}
		}
		// Try to parse Chatroom from response for direct navigation
		var room appmodels.Chatroom
		if ch, ok := res["Chatroom"]; ok {
			if b, e := json.Marshal(ch); e == nil {
				_ = json.Unmarshal(b, &room)
			}
		}
		// Best-effort delete notification on success
		_ = apiClient.DeleteNotification(notificationID)
		return inviteJoinedMsg{room: room, notiID: notificationID}
	}
}

func (m NotificationsModel) findNotificationIndex(notiID uint) int {
	items := m.notifications.Items()
	for i, it := range items {
		if ai, ok := it.(alertItem); ok {
			if uint(ai.notification.Id) == notiID {
				return i
			}
		}
	}
	return -1
}

type alertItem struct {
	notification appmodels.Notification
}

func (a alertItem) Title() string       { return a.notification.Content }
func (a alertItem) FilterValue() string { return a.notification.Content }

type alertDelegate struct{}

func newAlertDelegate() alertDelegate { return alertDelegate{} }

func (d alertDelegate) Height() int                               { return 2 }
func (d alertDelegate) Spacing() int                              { return 1 }
func (d alertDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d alertDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(alertItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	titleStyle := styles.ListItemTitleStyle
	if isSelected {
		titleStyle = styles.ListItemTitleSelectedStyle
	}

	title := titleStyle.Render(item.notification.Content)
	meta := styles.ListItemMetaStyle.Render(fmt.Sprintf("received %s", formatRelativeTime(item.notification.CreatedAt)))

	pointer := "  "
	if isSelected {
		pointer = styles.KeyStyle.Render("> ")
	}

	fmt.Fprintf(w, "%s%s\n    %s", pointer, title, meta)
}

func buildAlertItems(alerts []appmodels.Notification) []list.Item {
	items := make([]list.Item, len(alerts))
	for i, alert := range alerts {
		// Compact title for different types; fall back to content
		title := alert.Content
		if alert.Type != "" {
			title = fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Type), alert.Content)
		}
		// Keep raw notification; rendering uses Content anyway
		_ = title
		items[i] = alertItem{notification: alert}
	}
	return items
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	if t.IsZero() {
		return "sometime"
	}
	if t.After(now) {
		return fmt.Sprintf("in %s", humanizeDuration(t.Sub(now)))
	}
	return fmt.Sprintf("%s ago", humanizeDuration(now.Sub(t)))
}

func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		secs := int(d.Seconds())
		if secs < 1 {
			secs = 1
		}
		return fmt.Sprintf("%ds", secs)
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}
