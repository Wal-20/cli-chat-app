package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	stdpath "path"

	"github.com/Wal-20/cli-chat-app/internal/api/ws"
	"github.com/gorilla/websocket"
)

// SubscribeChatroom opens a websocket to receive live messages for a chatroom.
// It returns:
//   - a channel of incoming events,
//   - a cancel function to close the stream,
//   - and a send function to push events (e.g., typing / presence) to the server.
func (c *APIClient) SubscribeChatroom(chatroomID uint) (<-chan ws.WsEvent, func(), func(ws.WsEvent) error, error) {
	if c.baseURL == "" {
		return nil, nil, nil, fmt.Errorf("client not initialized")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, nil, nil, err
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	u.Path = stdpath.Join(u.Path, fmt.Sprintf("chatrooms/%d/ws", chatroomID))

	// Prepare headers
	header := http.Header{}
	if c.accessToken != "" {
		header.Set("Authorization", "Bearer "+c.accessToken)
	}

	// Dial WS
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		// Attempt automatic token refresh
		if resp != nil && resp.StatusCode == 401 && c.refreshToken != "" {
			if rerr := c.refreshTokens(); rerr == nil {
				header = http.Header{}
				if c.accessToken != "" {
					header.Set("Authorization", "Bearer "+c.accessToken)
				}
				conn, resp, err = websocket.DefaultDialer.Dial(u.String(), header)
			}
		}
		// Still failed
		if err != nil {
			if resp != nil {
				return nil, nil, nil, fmt.Errorf("ws dial failed: %s", resp.Status)
			}
			return nil, nil, nil, fmt.Errorf("ws dial error: %w", err)
		}
	}

	// read from ws server in goroutine
	ch := make(chan ws.WsEvent, 32)
	go func() {
		defer close(ch)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var evt ws.WsEvent
			if err := json.Unmarshal(data, &evt); err == nil {
				ch <- evt
			}
		}
	}()

	cancel := func() {
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
		)
		_ = conn.Close()
	}

	// return a function to be used in client to write to ws server
	send := func(evt ws.WsEvent) error {
		data, err := json.Marshal(evt)
		if err != nil {
			return err
		}
		return conn.WriteMessage(websocket.TextMessage, data)
	}

	return ch, cancel, send, nil
}
