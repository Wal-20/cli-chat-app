package handlers

import (
    "net/http"

    "github.com/Wal-20/cli-chat-app/internal/api/ws"
)

// ChatroomWebSocket handles websocket upgrade for a chatroom.
func ChatroomWebSocket(w http.ResponseWriter, r *http.Request) {
    ws.ServeChatroomWS(w, r)
}

