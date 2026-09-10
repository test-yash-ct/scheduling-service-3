package events

import (
	"context"
	"sync"
)

// Memory is a mutex-protected in-process outbox suitable for tests and
// single-process wiring until a durable dispatcher exists.
type Memory struct {
	mu     sync.Mutex
	events []Event
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Append(_ context.Context, event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *Memory) Events() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}
