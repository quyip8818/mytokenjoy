package notification

import (
	"sync"

	"github.com/google/uuid"
)

type SSEEvent struct {
	ID        string `json:"id"`
	EventType string `json:"eventType"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

// SSEHub manages per-user SSE connections for real-time notification push.
type SSEHub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID][]chan SSEEvent
}

// NewSSEHub creates a new hub for SSE notification broadcasting.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		subscribers: make(map[uuid.UUID][]chan SSEEvent),
	}
}

// Subscribe registers a new subscriber channel for a user. Returns a channel and an unsubscribe function.
func (h *SSEHub) Subscribe(userID uuid.UUID) (<-chan SSEEvent, func()) {
	ch := make(chan SSEEvent, 16)
	h.mu.Lock()
	h.subscribers[userID] = append(h.subscribers[userID], ch)
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.subscribers[userID]
		for i, sub := range subs {
			if sub == ch {
				h.subscribers[userID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
		if len(h.subscribers[userID]) == 0 {
			delete(h.subscribers, userID)
		}
	}

	return ch, unsubscribe
}

// Publish sends an event to all subscribers for a given user.
func (h *SSEHub) Publish(userID uuid.UUID, event SSEEvent) {
	h.mu.RLock()
	subs := h.subscribers[userID]
	h.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// drop if subscriber is slow — non-blocking
		}
	}
}

// ActiveSubscribers returns the number of active subscribers for a user.
func (h *SSEHub) ActiveSubscribers(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers[userID])
}
