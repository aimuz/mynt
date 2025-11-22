// Package event provides a unified event system for the NAS.
// It replaces the previous dual-system of EventService and NotificationBus.
package event

import (
	"strings"
	"sync"
	"time"
)

// Event represents a system event.
type Event struct {
	Type string    // Event type (e.g., "disk.added", "pool.degraded")
	Time time.Time // When the event occurred
	Data any       // Event-specific data
}

// Event type constants
const (
	DiskAdded        = "disk.added"
	DiskRemoved      = "disk.removed"
	SmartFailed      = "smart.failed"
	PoolDegraded     = "pool.degraded"
	PoolOnline       = "pool.online"
	DatasetCreated   = "dataset.created"
	DatasetDestroyed = "dataset.destroyed"
)

// Persist is an optional interface that can be implemented to persist events.
type Persister interface {
	Save(evt Event) error
}

// Bus is the central event distribution hub.
// It allows components to publish events and subscribe to patterns.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event // pattern -> channels
	persister   Persister               // optional persistence
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]chan Event),
	}
}

// SetPersister sets an optional persister for events.
func (b *Bus) SetPersister(p Persister) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.persister = p
}

// Publish sends an event to all matching subscribers.
// Events are sent asynchronously and non-blocking.
func (b *Bus) Publish(evt Event) {
	if evt.Time.IsZero() {
		evt.Time = time.Now()
	}

	// Persist event if persister is set
	if b.persister != nil {
		go b.persister.Save(evt) // Non-blocking
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for pattern, channels := range b.subscribers {
		if matchPattern(pattern, evt.Type) {
			for _, ch := range channels {
				select {
				case ch <- evt:
				default:
					// Drop event if subscriber is too slow
				}
			}
		}
	}
}

// Subscribe creates a subscription for events matching the pattern.
// Pattern can be:
//   - Exact match: "disk.added"
//   - Prefix match: "disk.*" matches all disk events
//   - All events: "*"
//
// The returned channel receives matching events.
// The caller must call Unsubscribe when done to prevent leaks.
func (b *Bus) Subscribe(pattern string) <-chan Event {
	ch := make(chan Event, 10) // Buffer to prevent blocking

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[pattern] = append(b.subscribers[pattern], ch)
	return ch
}

// Unsubscribe removes a subscription.
func (b *Bus) Unsubscribe(pattern string, ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	channels := b.subscribers[pattern]
	for i, subscriber := range channels {
		if subscriber == ch {
			// Remove from slice
			b.subscribers[pattern] = append(channels[:i], channels[i+1:]...)
			close(subscriber)

			// Clean up empty pattern
			if len(b.subscribers[pattern]) == 0 {
				delete(b.subscribers, pattern)
			}
			return
		}
	}
}

// matchPattern checks if an event type matches a subscription pattern.
func matchPattern(pattern, eventType string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == eventType {
		return true
	}
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return strings.HasPrefix(eventType, prefix+".")
	}
	return false
}
