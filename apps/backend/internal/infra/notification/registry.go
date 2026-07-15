package notification

import (
	"log/slog"
	"sync"
)

// Registry holds all registered notification channels.
type Registry struct {
	mu       sync.RWMutex
	channels map[string]Channel
	logger   *slog.Logger
}

// NewRegistry creates a new empty channel registry.
func NewRegistry(logger *slog.Logger) *Registry {
	return &Registry{
		channels: make(map[string]Channel),
		logger:   logger,
	}
}

// Register adds a channel to the registry.
func (r *Registry) Register(ch Channel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels[ch.Name()] = ch
	r.logger.Info("notification channel registered",
		"channel", ch.Name(),
		"configured", ch.IsConfigured())
}

// Get returns a channel by name.
func (r *Registry) Get(name string) (Channel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ch, ok := r.channels[name]
	return ch, ok
}

// Configured returns all channels that are fully configured and ready to send.
func (r *Registry) Configured() []Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Channel
	for _, ch := range r.channels {
		if ch.IsConfigured() {
			result = append(result, ch)
		}
	}
	return result
}

// ConfiguredNames returns the names of all configured channels.
func (r *Registry) ConfiguredNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for _, ch := range r.channels {
		if ch.IsConfigured() {
			names = append(names, ch.Name())
		}
	}
	return names
}

// All returns all registered channels regardless of configuration state.
func (r *Registry) All() []Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Channel, 0, len(r.channels))
	for _, ch := range r.channels {
		result = append(result, ch)
	}
	return result
}
