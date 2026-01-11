package ws

import (
	"encoding/json"
	"slices"
	"strings"
)

type typingEventQueue []string

func NewQueue() typingEventQueue {
	return typingEventQueue{}
}

func (q *typingEventQueue) add(username string) {
	username = strings.TrimSpace(username)
	if username == "" {
		return
	}
	if slices.Contains(*q, username) {
		return
	}
	*q = append(*q, username)
}

func (q *typingEventQueue) remove(username string) {
	username = strings.TrimSpace(username)
	if username == "" {
		return
	}
	for i, existing := range *q {
		if existing == username {
			*q = append((*q)[:i], (*q)[i+1:]...)
			return
		}
	}
}

func (q typingEventQueue) marshalUsernames() ([]byte, error) {
	list := make([]string, len(q))
	copy(list, q)
	return json.Marshal(list)
}
