package app

import (
	"sync"
	"uuid"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*Subscription
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uuid.UUID]*Subscription)}
}

func (h *Hub) Subscribe(user, workspace uuid.UUID) (subscription *Subscription, unsubscribe func()) {
	s := &Subscription{
		ID:          uuid.New(),
		UserID:      user,
		WorkspaceID: workspace,
		Messages:    make(chan Message, 64),
	}

	h.mu.Lock()
	h.clients[s.ID] = s
	h.mu.Unlock()

	return s, func() {
		h.mu.Lock()
		if _, ok := h.clients[s.ID]; ok {
			delete(h.clients, s.ID)
			close(s.Messages)
		}
		h.mu.Unlock()
	}
}

func (h *Hub) Publish(user, workspace uuid.UUID, m Message) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, s := range h.clients {
		if user != uuid.Nil() && s.UserID != user {
			continue
		}
		if workspace != uuid.Nil() && s.WorkspaceID != workspace {
			continue
		}
		select {
		case s.Messages <- m:
			count++
		default:
		}
	}

	return count
}
