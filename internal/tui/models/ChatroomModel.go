package models

import (
    "encoding/json"
    "fmt"
    "sort"
    "strconv"
    "strings"
    "time"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/Wal-20/cli-chat-app/internal/tui/client"
	"github.com/Wal-20/cli-chat-app/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sendMessageResultMsg struct {
	message models.MessageWithUser
	err     error
}

type ChatroomModel struct {
    apiClient          *client.APIClient
    username           string
    userID             uint
    chatroom           models.Chatroom
    messages           []models.MessageWithUser
    users              []models.UserChatroom
    input              textarea.Model
    viewport           viewport.Model
    width              int
    height             int
    messageColumnWidth int
    sidebarWidth       int
    showSidebar        bool
    sending            bool
    flashMessage       string
    flashStyle         lipgloss.Style
    wsChan             <-chan models.MessageWithUser
    wsCancel           func()
    // invite
    inviting           bool
    inviteInput        textinput.Model
}

func NewChatroomModel(username string, userID uint, chatroom models.Chatroom, apiClient *client.APIClient) ChatroomModel {
	input := textarea.New()
	input.Placeholder = "Write a message..."
	input.Prompt = ""
	input.CharLimit = 500
	input.SetWidth(80)
	input.SetHeight(4)
	input.ShowLineNumbers = false
	input.Focus()

	input.FocusedStyle.Base = styles.InputAreaStyle
	input.FocusedStyle.CursorLine = styles.InputAreaStyle.Copy()
	input.FocusedStyle.Placeholder = styles.InputPlaceholderStyle
	input.FocusedStyle.Text = styles.InputTextFocusedStyle

	input.BlurredStyle.Base = styles.InputAreaStyle
	input.BlurredStyle.CursorLine = styles.InputAreaStyle.Copy()
	input.BlurredStyle.Placeholder = styles.InputPlaceholderStyle
	input.BlurredStyle.Text = styles.InputTextStyle

	input.Cursor.Style = styles.KeyStyle

	messages, err := apiClient.GetMessages(chatroom.Id)
	if err != nil {
		panic(err)
	}

	users, err := apiClient.GetUsersByChatroom(chatroom.Id)
	if err != nil {
		panic(err)
	}

	vp := viewport.New(80, 20)
	vp.Style = styles.ConversationWrapperStyle
	vp.MouseWheelEnabled = true

    inv := textinput.New()
    inv.Prompt = "Invite (id or name): "
    inv.Placeholder = "e.g., 42 or alice"
    model := ChatroomModel{
        apiClient:    apiClient,
        username:     username,
        userID:       userID,
        chatroom:     chatroom,
        messages:     messages,
        users:        users,
        input:        input,
        viewport:     vp,
        sidebarWidth: styles.SidebarStyle.GetWidth(),
        showSidebar:  true,
        flashStyle:   styles.StatusInfoStyle,
        inviteInput:  inv,
    }

    // Attempt to subscribe to websocket updates for this room
    if ch, cancel, err := apiClient.SubscribeChatroom(chatroom.Id); err == nil {
        model.wsChan = ch
        model.wsCancel = cancel
    } else {
        model.flashMessage = fmt.Sprintf("Live updates unavailable: %s", err.Error())
        model.flashStyle = styles.StatusErrorStyle
    }

	model.refreshViewportContent(false)
	return model
}

func (m ChatroomModel) Init() tea.Cmd {
    if m.wsChan != nil {
        return tea.Batch(textarea.Blink, listenWS(m.wsChan))
    }
    return textarea.Blink
}

func (m ChatroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.applyWindowSize(msg.Width, msg.Height)
        m.refreshViewportContent(true)
        return m, nil

    case tea.KeyMsg:
        switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
        case "esc":
            if m.inviting {
                m.inviting = false
                m.inviteInput.Blur()
                return m, nil
            }
            if m.wsCancel != nil { m.wsCancel() }
            // Ensure next main screen pulls fresh memberships
            m.apiClient.InvalidateUserChatrooms()
            return NewMainChatModel(m.username, m.userID, m.apiClient), nil
        case "ctrl+u":
            m.viewport.HalfViewUp()
            return m, nil
        case "ctrl+d":
            m.viewport.HalfViewDown()
            return m, nil
        case "ctrl+i":
            if !m.inviting {
                if !m.currentUserIsAdmin() {
                    m.flashMessage = "Only admins can invite users"
                    m.flashStyle = styles.StatusErrorStyle
                    return m, nil
                }
                m.inviting = true
                m.inviteInput.SetValue("")
                m.inviteInput.Focus()
                m.flashMessage = ""
                return m, nil
            }
        case "enter":
            if m.inviting {
                ident := strings.TrimSpace(m.inviteInput.Value())
                if ident == "" {
                    m.flashMessage = "Enter user id or username"
                    m.flashStyle = styles.StatusErrorStyle
                    return m, nil
                }
                m.inviting = false
                return m, inviteUserCmd(m.apiClient, m.chatroom.Id, ident)
            }
            if m.sending {
                return m, nil
            }
            content := strings.TrimSpace(m.input.Value())
            if content == "" {
                m.flashMessage = "Nothing to send yet."
                m.flashStyle = styles.StatusErrorStyle
                return m, nil
            }

			m.sending = true
			m.input.Reset()
			m.flashMessage = "Sending message..."
			m.flashStyle = styles.StatusInfoStyle

			chatroomID := m.chatroom.Id
			return m, tea.Batch(sendMessage(m.apiClient, chatroomID, m.username, content))
		}

    case sendMessageResultMsg:
		m.sending = false
        if msg.err != nil {
            m.flashMessage = fmt.Sprintf("Failed to send: %s", msg.err.Error())
            m.flashStyle = styles.StatusErrorStyle
            return m, nil
        }

		if msg.message.Username == "" {
			msg.message.Username = m.username
		}

		m.messages = append(m.messages, msg.message)
		m.ensureParticipant(msg.message.Username)
		m.flashMessage = fmt.Sprintf("Delivered at %s", time.Now().Format("15:04:05"))
		m.flashStyle = styles.StatusSuccessStyle
        m.refreshViewportContent(false)
        return m, nil
    case wsMessageMsg:
        // Append incoming message, ensure participant, keep scroll position
        if len(m.messages) > 0 {
            last := m.messages[len(m.messages)-1]
            if strings.EqualFold(last.Username, msg.message.Username) && last.Content == msg.message.Content && last.CreatedAt.Equal(msg.message.CreatedAt) {
                return m, listenWS(m.wsChan)
            }
        }
        m.messages = append(m.messages, msg.message)
        m.ensureParticipant(msg.message.Username)
        m.refreshViewportContent(true)
        // continue listening for the next websocket message
        return m, listenWS(m.wsChan)
    case wsClosedMsg:
        m.flashMessage = "Live updates disconnected"
        m.flashStyle = styles.StatusErrorStyle
        return m, nil
    }

	var cmds []tea.Cmd

	var viewportCmd tea.Cmd
	m.viewport, viewportCmd = m.viewport.Update(msg)
	cmds = append(cmds, viewportCmd)

	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
		return m, tea.Batch(cmds...)
	}

    if m.inviting {
        var ic tea.Cmd
        m.inviteInput, ic = m.inviteInput.Update(msg)
        cmds = append(cmds, ic)
    } else {
        var inputCmd tea.Cmd
        m.input, inputCmd = m.input.Update(msg)
        cmds = append(cmds, inputCmd)
    }

    return m, tea.Batch(cmds...)
}

// tea message types for websocket events
type wsMessageMsg struct{ message models.MessageWithUser }
type wsClosedMsg struct{}

func listenWS(ch <-chan models.MessageWithUser) tea.Cmd {
    return func() tea.Msg {
        msg, ok := <-ch
        if !ok {
            return wsClosedMsg{}
        }
        return wsMessageMsg{message: msg}
    }
}

func (m ChatroomModel) View() string {
	header := styles.TitleStyle.Render(m.chatroom.Title)

	visibility := "private"
	if m.chatroom.IsPublic {
		visibility = "public"
	}
	summary := styles.SubtitleStyle.Render(fmt.Sprintf("%s | %d of %d participants", visibility, len(m.users), m.chatroom.MaxUserCount))

	conversation := m.viewport.View()

	var sidebar string
	if m.showSidebar {
		sidebar = m.renderSidebar()
	}

	var mainRow string
	if sidebar != "" {
		mainRow = lipgloss.JoinHorizontal(lipgloss.Top, conversation, sidebar)
	} else {
		mainRow = conversation
	}

	inputWidth := m.messageColumnWidth
	if inputWidth <= 0 {
		inputWidth = m.viewport.Width + styles.ConversationWrapperStyle.GetHorizontalFrameSize()
		if inputWidth <= 0 {
			inputWidth = 80
		}
	}
    inputView := styles.InputAreaStyle.Copy().Width(inputWidth).Render(m.input.View())
    if m.inviting {
        inviteView := styles.InputFieldFocusedStyle.Render(m.inviteInput.View())
        inputView = inviteView + "\n" + inputView
    }

	info := fmt.Sprintf("%d messages | %d participants", len(m.messages), len(m.users))
	statusStyle := styles.StatusInfoStyle
	if m.flashMessage != "" {
		info = m.flashMessage
		statusStyle = m.flashStyle
	}

	helpItems := []string{
		styles.RenderKeyBinding("Esc", "Back"),
		styles.RenderKeyBinding("Enter", "Send"),
		styles.RenderKeyBinding("Shift+Enter", "New line"),
		styles.RenderKeyBinding("Ctrl+i", "Invite User"),
		styles.RenderKeyBinding("Ctrl+c", "Quit"),
		styles.RenderKeyBinding("Ctrl+U/Ctrl+D", "Scroll"),
	}
	help := strings.Join(helpItems, styles.HelpStyle.Render("  "))

    footerContent := statusStyle.Render(info) + "\n" + styles.HelpStyle.Render(help)
    footer := styles.StatusBarStyle.Render(footerContent)

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		summary,
		"",
		mainRow,
		"",
		inputView,
		"",
		footer,
	)

	if m.width > 0 && m.height > 0 {
		app := styles.AppStyle.Copy().Width(m.width).Height(m.height)
		return app.Render(layout)
	}

	return styles.AppStyle.Render(layout)
}

func (m *ChatroomModel) applyWindowSize(width, height int) {
	m.width = width
	m.height = height

	sidebarWidth := styles.SidebarStyle.GetWidth()
	if width < 100 {
		sidebarWidth = 0
	}

	messageWidth := width - sidebarWidth - 6
	if messageWidth < 40 {
		messageWidth = width - 6
	}
	if messageWidth < 32 {
		messageWidth = 32
	}

	messageHeight := height - 10
	if messageHeight < 8 {
		messageHeight = 8
	}

	contentWidth := messageWidth - styles.ConversationWrapperStyle.GetHorizontalFrameSize()
	if contentWidth < 24 {
		contentWidth = 24
	}

	inputWidth := messageWidth - styles.InputAreaStyle.GetHorizontalFrameSize()
	if inputWidth < 24 {
		inputWidth = 24
	}

	m.messageColumnWidth = messageWidth
	m.sidebarWidth = sidebarWidth
	m.showSidebar = sidebarWidth > 0

	m.viewport.Width = contentWidth
	m.viewport.Height = messageHeight
	m.viewport.Style = styles.ConversationWrapperStyle.Copy().Width(messageWidth)

	m.input.SetWidth(inputWidth)
	if height < 24 {
		m.input.SetHeight(3)
	} else {
		m.input.SetHeight(5)
	}
}

func (m *ChatroomModel) refreshViewportContent(keepPosition bool) {
	wasAtBottom := m.viewport.AtBottom()
	previousOffset := m.viewport.YOffset

	m.viewport.SetContent(m.renderMessages())

	if keepPosition && !wasAtBottom {
		m.viewport.YOffset = previousOffset
	} else {
		m.viewport.GotoBottom()
	}
}

func (m ChatroomModel) renderMessages() string {
	if len(m.messages) == 0 {
		return styles.MutedTextStyle.Render("No messages yet. Say hi to get things started!")
	}

	contentWidth := m.viewport.Width
	if contentWidth <= 0 {
		contentWidth = 60
	}

	var sections []string
	var prevDate string

	for _, message := range m.messages {
		// Simple date separator when date changes
		dateKey := message.CreatedAt.Format("2006-01-02")
		if dateKey != prevDate {
			sections = append(sections, styles.DateDividerStyle.Width(contentWidth).Render(formatDateSeparator(message.CreatedAt)))
			prevDate = dateKey
		}

		isSelf := strings.EqualFold(message.Username, m.username)
		authorStyle := styles.MessageAuthorStyle
		if isSelf {
			authorStyle = styles.MessageAuthorSelfStyle
		}

		// Inline role tag, lightly muted
		role := getUserRole(m.users, message.Username, m.username)
		roleTag := ""
		switch role {
		case "Owner":
			roleTag = styles.MutedTextStyle.Render(" [owner]")
		case "Admin":
			roleTag = styles.MutedTextStyle.Render(" [admin]")
		case "You":
			roleTag = styles.MutedTextStyle.Render(" [you]")
		}

		author := authorStyle.Render(message.Username) + roleTag
		timestamp := styles.MessageTimestampStyle.Render(message.CreatedAt.Format("15:04"))

		// Single-line structure: "author HH:MM: content"; wrapped to viewport width
		line := fmt.Sprintf("%s %s: %s", author, timestamp, message.Content)
		wrapped := wrapText(line, contentWidth)
		sections = append(sections, styles.MessageContainerStyle.Render(wrapped))
	}

	return strings.Join(sections, "\n")
}

func (m ChatroomModel) renderSidebar() string {
	users := append([]models.UserChatroom(nil), m.users...)
	sort.Slice(users, func(i, j int) bool {
		ri, rj := sidebarRank(users[i], m.username), sidebarRank(users[j], m.username)
		if ri == rj {
			return strings.ToLower(users[i].Name) < strings.ToLower(users[j].Name)
		}
		return ri < rj
	})

	lines := []string{styles.SidebarTitleStyle.Render("Members")}
	for _, user := range users {
		line := styles.ParticipantLineStyle.Render(user.Name)
		badges := []string{}
		if user.IsOwner {
			badges = append(badges, styles.ParticipantBadgeOwnerStyle.Render(" owner"))
		}
		if user.IsAdmin {
			badges = append(badges, styles.ParticipantBadgeAdminStyle.Render(" admin"))
		}
		if strings.EqualFold(user.Name, m.username) {
			badges = append(badges, styles.ParticipantBadgeYouStyle.Render(" you"))
		}
		if len(badges) > 0 {
			line = lipgloss.JoinHorizontal(lipgloss.Left, line, strings.Join(badges, " "))
		}
		lines = append(lines, line)
	}

	if len(lines) == 1 {
		lines = append(lines, styles.MutedTextStyle.Render("No other participants yet."))
	}

	return styles.SidebarStyle.Render(strings.Join(lines, "\n"))
}

func (m *ChatroomModel) ensureParticipant(username string) {
	for _, user := range m.users {
		if strings.EqualFold(user.Name, username) {
			return
		}
	}

	m.users = append(m.users, models.UserChatroom{Name: username})
}

func (m ChatroomModel) currentUserIsAdmin() bool {
    for _, u := range m.users {
        if u.UserID == m.userID {
            return u.IsAdmin || u.IsOwner
        }
    }
    return false
}

func sendMessage(apiClient *client.APIClient, chatroomID uint, username, content string) tea.Cmd {
	return func() tea.Msg {
		result, err := apiClient.SendMessage(strconv.FormatUint(uint64(chatroomID), 10), content)
		if err != nil {
			return sendMessageResultMsg{err: err}
		}

		var message models.MessageWithUser
		payload, ok := result["Message"]
		if ok {
			raw, _ := json.Marshal(payload)
			_ = json.Unmarshal(raw, &message)
		}

		if message.Username == "" {
			message.Username = username
		}
		if message.Content == "" {
			message.Content = content
		}
		if message.CreatedAt.IsZero() {
			message.CreatedAt = time.Now()
		}

		return sendMessageResultMsg{message: message}
	}
}

// invite user by id or username
func inviteUserCmd(apiClient *client.APIClient, chatroomID uint, ident string) tea.Cmd {
    return func() tea.Msg {
        err := apiClient.InviteUser(strconv.FormatUint(uint64(chatroomID), 10), ident)
        // Reuse flash mechanism via sendMessageResultMsg with err only
        if err != nil {
            return sendMessageResultMsg{err: fmt.Errorf("invite failed: %w", err)}
        }
        return sendMessageResultMsg{message: models.MessageWithUser{Content: "", Username: ""}}
    }
}

func getUserRole(users []models.UserChatroom, username string, currentUsername string) string {
	if strings.EqualFold(username, currentUsername) {
		return "You"
	}
	for _, user := range users {
		if !strings.EqualFold(user.Name, username) {
			continue
		}
		if user.IsOwner {
			return "Owner"
		}
		if user.IsAdmin {
			return "Admin"
		}
		return "Member"
	}
	return "Guest"
}

func renderRoleBadge(role string) string {
	switch role {
	case "Owner":
		return styles.ParticipantBadgeOwnerStyle.Render("owner")
	case "Admin":
		return styles.ParticipantBadgeAdminStyle.Render("admin")
	case "You":
		return styles.ParticipantBadgeYouStyle.Render("you")
	default:
		return ""
	}
}

func sidebarRank(user models.UserChatroom, current string) int {
	switch {
	case strings.EqualFold(user.Name, current):
		return 0
	case user.IsOwner:
		return 1
	case user.IsAdmin:
		return 2
	default:
		return 3
	}
}

func formatDateSeparator(t time.Time) string {
	today := time.Now().Format("2006-01-02")
	msgDate := t.Format("2006-01-02")
	if msgDate == today {
		return "— Today —"
	}
	return "— " + t.Format("Mon Jan 02") + " —"
}

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var line strings.Builder
	lineLen := 0

	for _, word := range words {
		wordLen := len(word)
		if lineLen > 0 && lineLen+1+wordLen > width {
			lines = append(lines, line.String())
			line.Reset()
			lineLen = 0
		}
		if lineLen > 0 {
			line.WriteByte(' ')
			lineLen++
		}
		if wordLen > width {
			for start := 0; start < wordLen; start += width {
				end := start + width
				if end > wordLen {
					end = wordLen
				}
				chunk := word[start:end]
				if lineLen > 0 {
					lines = append(lines, line.String())
					line.Reset()
					lineLen = 0
				}
				if len(chunk) == width {
					lines = append(lines, chunk)
				} else {
					line.WriteString(chunk)
					lineLen += len(chunk)
				}
			}
			continue
		}
		line.WriteString(word)
		lineLen += wordLen
	}

	if lineLen > 0 {
		lines = append(lines, line.String())
	}

	if len(lines) == 0 {
		return text
	}

	return strings.Join(lines, "\n")
}
