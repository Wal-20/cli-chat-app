package models

import (
    "strconv"
    "strings"
    "github.com/Wal-20/cli-chat-app/internal/models"
    "github.com/Wal-20/cli-chat-app/internal/tui/client"
    "github.com/Wal-20/cli-chat-app/internal/tui/styles"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
)

type CreateChatroomModel struct {
    apiClient *client.APIClient
    username  string
    userID    uint

    title     textinput.Model
    maxUsers  textinput.Model
    isPublic  bool

    submitting bool
    statusMessage string
}

func NewCreateChatroomModel(username string, userID uint, apiClient *client.APIClient) CreateChatroomModel {
    title := textinput.New()
    title.Prompt = "Title: "
    title.PromptStyle = styles.InputPromptFocusedStyle
    title.TextStyle = styles.InputTextFocusedStyle
    title.Focus()

    max := textinput.New()
    max.Prompt = "Max users: "
    max.PromptStyle = styles.InputPromptStyle
    max.TextStyle = styles.InputTextStyle
    max.Placeholder = "10"

    return CreateChatroomModel{
        apiClient: apiClient,
        username:  username,
        userID:    userID,
        title:     title,
        maxUsers:  max,
        isPublic:  true,
    }
}

func (m CreateChatroomModel) Init() tea.Cmd { return textinput.Blink }

func (m CreateChatroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case createdChatroomMsg:
        return m.UpdateCreated(msg)
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            // go back to main
            return NewMainChatModel(m.username, m.userID, m.apiClient), nil
        case "tab":
            if m.title.Focused() {
                m.title.Blur(); m.maxUsers.Focus()
            } else { m.maxUsers.Blur(); m.title.Focus() }
            return m, nil
        case " ":
            // toggle public/private when not focused on inputs
            if !m.title.Focused() && !m.maxUsers.Focused() {
                m.isPublic = !m.isPublic
                return m, nil
            }
        case "enter":
            if m.submitting { return m, nil }
            t := strings.TrimSpace(m.title.Value())
            if t == "" { m.statusMessage = "Title is required"; return m, nil }
            maxCount := 10
            if s := strings.TrimSpace(m.maxUsers.Value()); s != "" {
                if v, err := strconv.Atoi(s); err == nil && v > 0 { maxCount = v }
            }
            m.submitting = true
            return m, createChatroomCmd(m.apiClient, m.username, t, maxCount, m.isPublic)
        }
    }
    var cmd tea.Cmd
    if m.title.Focused() { m.title, cmd = m.title.Update(msg) } else { m.maxUsers, cmd = m.maxUsers.Update(msg) }
    return m, cmd
}

func (m CreateChatroomModel) View() string {
    toggle := "Public"
    if !m.isPublic { toggle = "Private" }
    btn := styles.RenderButton("Create", true)
    status := m.statusMessage
    return styles.AppStyle.Render(strings.Join([]string{
        styles.CardTitleStyle.Render("New Chatroom"),
        styles.InputFieldFocusedStyle.Render(m.title.View()),
        styles.InputFieldStyle.Render(m.maxUsers.View()),
        styles.StatusInfoStyle.Render("Space to toggle: "+toggle),
        btn,
        status,
        styles.HelpStyle.Render(strings.Join([]string{
            styles.RenderKeyBinding("Tab", "Next field"),
            styles.RenderKeyBinding("Space", "Toggle public"),
            styles.RenderKeyBinding("Enter", "Create"),
            styles.RenderKeyBinding("Esc", "Back"),
        }, styles.HelpStyle.Render("  "))),
    }, "\n\n"))
}

type createdChatroomMsg struct { chatroom models.Chatroom; err error }

func createChatroomCmd(api *client.APIClient, username, title string, maxUsers int, isPublic bool) tea.Cmd {
    return func() tea.Msg {
        room, err := api.CreateChatroom(title, maxUsers, isPublic)
        if err != nil { return createdChatroomMsg{err: err} }
        return createdChatroomMsg{chatroom: room}
    }
}

// Handle result messages in Update
func (m CreateChatroomModel) UpdateCreated(msg createdChatroomMsg) (tea.Model, tea.Cmd) {
    if msg.err != nil {
        m.submitting = false
        m.statusMessage = msg.err.Error()
        return m, nil
    }
    // jump right into the new room
    return NewChatroomModel(m.username, m.userID, msg.chatroom, m.apiClient), nil
}
