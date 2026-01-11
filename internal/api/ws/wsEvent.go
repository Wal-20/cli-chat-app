package ws

import (
	"encoding/json"
)

type WsEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
