package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

// Client represents a single websocket connection.
type Client struct {
	room *Room
	conn *websocket.Conn
	send chan []byte
}

type wsUserStatusPayload struct {
	Username string `json:"username"`
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

		var evt WsEvent
		if err := json.Unmarshal(data, &evt); err != nil {
			// Ignore malformed input; keep the connection alive.
			continue
		}

		switch evt.Type {
		case "typing":
			var payload wsUserStatusPayload
			if err := json.Unmarshal(evt.Data, &payload); err != nil || payload.Username == "" {
				continue
			}
			c.room.typingEventQ.add(payload.Username)
			c.room.broadcastTypingQueue()
		case "stoppedTyping":
			var payload wsUserStatusPayload
			if err := json.Unmarshal(evt.Data, &payload); err != nil || payload.Username == "" {
				continue
			}
			c.room.typingEventQ.remove(payload.Username)
			c.room.broadcastTypingQueue()
		case "joined", "left":
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
