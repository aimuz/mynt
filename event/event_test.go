package event

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBus_Subscribe(t *testing.T) {
	bus := NewBus()

	ch := bus.Subscribe("test.event")
	defer bus.Unsubscribe("test.event", ch)

	bus.Publish(Event{Type: "test.event", Data: "test"})

	select {
	case e := <-ch:
		require.Equal(t, "test.event", e.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event not received")
	}
}

func TestBus_Subscribe_MultipleSubscribers(t *testing.T) {
	bus := NewBus()

	ch1 := bus.Subscribe("test.event")
	ch2 := bus.Subscribe("test.event")
	defer bus.Unsubscribe("test.event", ch1)
	defer bus.Unsubscribe("test.event", ch2)

	bus.Publish(Event{Type: "test.event"})

	received := 0
	timeout := time.After(100 * time.Millisecond)

	for received < 2 {
		select {
		case <-ch1:
			received++
		case <-ch2:
			received++
		case <-timeout:
			t.Fatal("Did not receive events on both channels")
		}
	}

	require.Equal(t, 2, received)
}

func TestBus_Subscribe_DifferentTypes(t *testing.T) {
	bus := NewBus()

	ch1 := bus.Subscribe("type1")
	ch2 := bus.Subscribe("type2")
	defer bus.Unsubscribe("type1", ch1)
	defer bus.Unsubscribe("type2", ch2)

	bus.Publish(Event{Type: "type1"})

	select {
	case <-ch1:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("type1 event not received")
	}

	select {
	case <-ch2:
		t.Fatal("type2 channel should not receive type1 event")
	case <-time.After(50 * time.Millisecond):
		// OK - timeout expected
	}
}

func TestBus_PatternMatching(t *testing.T) {
	bus := NewBus()

	tests := []struct {
		name        string
		pattern     string
		eventType   string
		shouldMatch bool
	}{
		{"exact match", "disk.added", "disk.added", true},
		{"prefix match", "disk.*", "disk.added", true},
		{"prefix match 2", "disk.*", "disk.removed", true},
		{"prefix no match", "disk.*", "pool.created", false},
		{"wildcard", "*", "any.event", true},
		{"no match", "disk.added", "pool.created", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := bus.Subscribe(tt.pattern)
			defer bus.Unsubscribe(tt.pattern, ch)

			bus.Publish(Event{Type: tt.eventType})

			select {
			case <-ch:
				if !tt.shouldMatch {
					t.Fatal("Event received but should not match")
				}
			case <-time.After(50 * time.Millisecond):
				if tt.shouldMatch {
					t.Fatal("Event not received but should match")
				}
			}
		})
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	bus := NewBus()

	ch := bus.Subscribe("test.event")

	// Publish first event
	bus.Publish(Event{Type: "test.event"})

	select {
	case <-ch:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("First event not received")
	}

	// Unsubscribe
	bus.Unsubscribe("test.event", ch)

	// Publish second event - should not be received (channel closed)
	bus.Publish(Event{Type: "test.event"})

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("Received event after unsubscribe")
		}
		// Channel closed - expected
	case <-time.After(50 * time.Millisecond):
		// Timeout also acceptable
	}
}

func TestBus_ConcurrentOperations(t *testing.T) {
	bus := NewBus()
	wg := sync.WaitGroup{}

	// Subscribe from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := bus.Subscribe("test.event")
			bus.Unsubscribe("test.event", ch)
		}()
	}

	// Publish from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(Event{Type: "test.event"})
		}()
	}

	wg.Wait()
	// Test passes if no race conditions detected
}

type mockPersister struct {
	events []Event
	mu     sync.Mutex
}

func (m *mockPersister) Save(e Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}

func TestBus_WithPersister(t *testing.T) {
	bus := NewBus()
	persister := &mockPersister{}
	bus.SetPersister(persister)

	bus.Publish(Event{
		Type: "test.event",
		Data: "test_data",
	})

	time.Sleep(50 * time.Millisecond)

	persister.mu.Lock()
	require.Len(t, persister.events, 1)
	require.Equal(t, "test.event", persister.events[0].Type)
	persister.mu.Unlock()
}

func TestEvent_AutoTimestamp(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe("test")
	defer bus.Unsubscribe("test", ch)

	bus.Publish(Event{Type: "test"})

	select {
	case e := <-ch:
		require.False(t, e.Time.IsZero())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event not received")
	}
}
