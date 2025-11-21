package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Wal-20/cli-chat-app/internal/models"
	"github.com/gorilla/websocket"
)

// Hub manages websocket rooms keyed by chatroom ID.
type Hub struct {
	mu    sync.RWMutex
	rooms map[uint]*Room
}

var hub = &Hub{rooms: make(map[uint]*Room)}

func getHub() *Hub { return hub }

// Room maintains active clients and broadcasts messages to them.
type Room struct {
	id         uint
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

func newRoom(id uint) *Room {
	r := &Room{
		id:         id,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
	go r.run()
	return r
}

func (r *Room) run() {
	for {
		select {
		case c := <-r.register:
			r.clients[c] = true
		case c := <-r.unregister:
			if _, ok := r.clients[c]; ok {
				delete(r.clients, c)
				close(c.send)
			}
		case msg := <-r.broadcast:
			for c := range r.clients {
				select {
				case c.send <- msg:
				default:
					// slow client; drop
				}
			}
		}
	}
}

// Client represents a single websocket connection.
type Client struct {
	room *Room
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(1 << 20)
	c.conn.SetCloseHandler(func(code int, text string) error { return nil })
	for {
		// Handle incoming events from this client (e.g., typing, joined).
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var evt models.WsEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			// Ignore malformed input; keep the connection alive.
			continue
		}

		switch evt.Type {
		case "typing", "joined", "left":
			// Fan out ephemeral status events to everyone in the room.
			c.room.broadcast <- data
		default:
			// Ignore other incoming event types for now.
		}
	}
}

func (c *Client) writePump() {
	defer func() { _ = c.conn.Close() }()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// GetRoom returns existing room or creates a new one.
func GetRoom(id uint) *Room {
	h := getHub()
	h.mu.RLock()
	r := h.rooms[id]
	h.mu.RUnlock()
	if r != nil {
		return r
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if r = h.rooms[id]; r == nil {
		r = newRoom(id)
		h.rooms[id] = r
	}
	return r
}

// BroadcastMessage sends the given message to all websocket clients in a room.
func BroadcastMessage(roomID uint, message models.WsEvent) {
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
