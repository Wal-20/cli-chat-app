package ws

import (
	"encoding/json"
	"log"
)

// Room maintains active clients and broadcasts messages to them.
type Room struct {
	id           uint
	clients      map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	broadcast    chan []byte
	typingEventQ typingEventQueue
}

func newRoom(id uint) *Room {
	r := &Room{
		id:           id,
		clients:      make(map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan []byte, 256),
		typingEventQ: NewQueue(),
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
	r.broadcast <- b
}
