package events

import "context"

// Outbox is an in-process append-only log for write-side domain events.
// A future dispatcher can drain this without changing producers.
type Outbox interface {
	Append(ctx context.Context, event Event) error
}

// Nop is a no-op Outbox used when events are not observed.
type Nop struct{}

func (Nop) Append(context.Context, Event) error { return nil }
