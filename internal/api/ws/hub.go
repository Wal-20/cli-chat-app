package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages websocket rooms keyed by chatroom ID.
type Hub struct {
	mu    sync.RWMutex
	rooms map[uint]*Room
}

var hub = &Hub{rooms: make(map[uint]*Room)}

// getHub returns the process-wide websocket hub singleton.
func getHub() *Hub { return hub }

// BroadcastMessage sends a server-originated websocket event to all clients currently
// connected to the given room ID.
func BroadcastMessage(roomID uint, message WsEvent) {
	b, err := json.Marshal(message)
	if err != nil {
		log.Printf("ws marshal error: %v", err)
		return
	}
	GetRoom(roomID).broadcast <- b
}

// UpdateTyping enqueues a typing status update to be applied by the room's single
// goroutine (the room loop), which serializes updates and avoids concurrent writes
// to the typing queue.
func (h *Hub) UpdateTyping(roomID uint, client *Client, username string, isTyping bool) {
	GetRoom(roomID).queueTypingUpdate(client, username, isTyping)
}

// upgrader handles the HTTP->WebSocket upgrade; CheckOrigin is permissive because
// auth/authorization is handled at a higher layer.
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// ServeChatroomWS upgrades the incoming HTTP request to a websocket connection,
// resolves the chatroom ID from the request path, registers the new client with
// the room, and starts the read/write pumps.
func ServeChatroomWS(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "missing chatroom id", http.StatusBadRequest)
		return
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid chatroom id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	room := GetRoom(uint(id64))
	client := &Client{room: room, conn: conn, send: make(chan []byte, 256)}
	room.register <- client

	go client.writePump()
	client.readPump()
}
