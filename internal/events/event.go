package events

import (
	"time"

	"github.com/google/uuid"
)

// Event is the shared integration-event envelope. Payload must contain
// identifiers only (no names, contact details, notes, or other PHI).
type Event struct {
	EventID    string         `json:"event_id"`
	EventType  string         `json:"event_type"`
	TenantID   string         `json:"tenant_id"`
	OccurredAt string         `json:"occurred_at"`
	RequestID  string         `json:"request_id"`
	Payload    map[string]any `json:"payload"`
}

func New(eventType, tenantID, requestID string, payload map[string]any) Event {
	copied := make(map[string]any, len(payload))
	for k, v := range payload {
		copied[k] = v
	}
	return Event{
		EventID:    uuid.NewString(),
		EventType:  eventType,
		TenantID:   tenantID,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		RequestID:  requestID,
		Payload:    copied,
	}
}
