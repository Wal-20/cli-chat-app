package ws

import (
	"encoding/json"
	"log"
	"strings"
)

// Room maintains active clients and broadcasts messages to them.
type Room struct {
	id             uint
	clients        map[*Client]bool
	registerChan   chan *Client
	unregisterChan chan *Client
	broadcastChan  chan []byte
	typingEventQ   typingEventQueue
	typingChan     chan typingUpdate  // carries typing start/stop updates into the room loop for serialized processing.
	typingByConn   map[*Client]string // stores the current "typing as username" state per websocket client connection.
	typingCounts   map[string]int     // reference-counts active typing connections per username (handles multi-tab/devices).
}

// newRoom constructs a room and starts its event loop goroutine.
func newRoom(id uint) *Room {
	r := &Room{
		id:             id,
		clients:        make(map[*Client]bool),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		broadcastChan:  make(chan []byte, 256),
		typingEventQ:   NewTypingQueue(),
		typingChan:     make(chan typingUpdate, 256),
		typingByConn:   make(map[*Client]string),
		typingCounts:   make(map[string]int),
	}
	go r.run()
	return r
}

// run is the room event loop; it serializes all room state changes and fans out broadcasts.
func (r *Room) run() {
	for {
		select {
		case c := <-r.registerChan:
			r.clients[c] = true
		case c := <-r.unregisterChan:
			if _, ok := r.clients[c]; ok {
				changed := r.clearTypingForClient(c)
				delete(r.clients, c)
				close(c.send)
				if changed {
					r.broadcastTypingQueue()
				}
			}
		case msg := <-r.broadcastChan:
			r.broadcastToClients(msg)
		case u := <-r.typingChan:
			if changed := r.applyTypingUpdate(u); changed {
				r.broadcastTypingQueue()
			}
		}
	}
}

// typingUpdate is a room-local event describing a typing start/stop for a specific connection.
type typingUpdate struct {
	client   *Client
	username string
	isTyping bool
}

// queueTypingUpdate posts a typing update to the room loop for serialized processing.
func (r *Room) queueTypingUpdate(client *Client, username string, isTyping bool) {
	r.typingChan <- typingUpdate{client: client, username: username, isTyping: isTyping}
}

// applyTypingUpdate applies a queued typing update and returns true if the visible typing queue changed.
func (r *Room) applyTypingUpdate(u typingUpdate) bool {
	if u.client == nil {
		return false
	}

	if !u.isTyping {
		return r.clearTypingForClient(u.client)
	}

	if _, ok := r.clients[u.client]; !ok {
		// Ignore typing starts for disconnected/unregistered clients.
		return false
	}

	username := strings.TrimSpace(u.username)
	if username == "" {
		return false
	}

	prevUsername, wasTyping := r.typingByConn[u.client]
	if wasTyping && prevUsername == username {
		return false
	}

	changed := false
	if wasTyping {
		changed = r.decrementTyping(prevUsername) || changed
	}

	r.typingByConn[u.client] = username
	changed = r.incrementTyping(username) || changed
	return changed
}

// clearTypingForClient clears any typing state associated with the given connection.
func (r *Room) clearTypingForClient(client *Client) bool {
	prevUsername, ok := r.typingByConn[client]
	if !ok {
		return false
	}
	delete(r.typingByConn, client)
	return r.decrementTyping(prevUsername) // this is called to account for multiple connections for the same username, it decrements typingCounts for the username
}

// incrementTyping increments the active typing count for a username and adds it to the queue on first entry, returns true if first entry for the username
func (r *Room) incrementTyping(username string) bool {
	username = strings.TrimSpace(username)
	if username == "" {
		return false
	}
	count := r.typingCounts[username]
	r.typingCounts[username] = count + 1
	if count == 0 {
		r.typingEventQ.add(username)
		return true
	}
	return false
}

// decrementTyping decrements the active typing count for a username and removes it from the queue on last exit.
func (r *Room) decrementTyping(username string) bool {
	username = strings.TrimSpace(username)
	if username == "" {
		return false
	}
	count := r.typingCounts[username]
	if count <= 1 {
		delete(r.typingCounts, username)
		r.typingEventQ.remove(username)
		return true
	}
	r.typingCounts[username] = count - 1
	return false
}

// broadcastToClients sends a raw websocket message to all connected clients, dropping slow consumers.
func (r *Room) broadcastToClients(msg []byte) {
	for c := range r.clients {
		select {
		case c.send <- msg:
		default:
			// slow client; drop
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

// broadcastTypingQueue emits the current room typing queue as a "typing_queue" websocket event.
func (r *Room) broadcastTypingQueue() {
	payload, err := r.typingEventQ.marshalUsernames()
	if err != nil {
		log.Printf("ws typing queue marshal error: %v", err)
		return
	}
	evt := WsEvent{
		Type: "typing_queue",
		Data: json.RawMessage(payload),
	}
	b, err := json.Marshal(evt)
	if err != nil {
		log.Printf("ws typing queue event marshal error: %v", err)
		return
	}
	r.broadcastToClients(b)
}
