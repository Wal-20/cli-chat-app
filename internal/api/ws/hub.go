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

func getHub() *Hub { return hub }

// BroadcastMessage sends the given message to all websocket clients in a room.
func BroadcastMessage(roomID uint, message WsEvent) {
	b, err := json.Marshal(message)
	if err != nil {
		log.Printf("ws marshal error: %v", err)
		return
	}
	GetRoom(roomID).broadcast <- b
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// ServeChatroomWS upgrades a connection and joins the specified chatroom.
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
