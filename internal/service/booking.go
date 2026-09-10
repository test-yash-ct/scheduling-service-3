package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/healthops/scheduling-service/internal/events"
	"github.com/healthops/scheduling-service/internal/obs"
)

// SlotBooker is the persistence slice needed to book an appointment.
type SlotBooker interface {
	BookSlot(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time) error
}

// Booking is a thin write-side use case so handlers do not own outbox emission.
type Booking struct {
	Slots  SlotBooker
	Outbox events.Outbox
}

func NewBooking(slots SlotBooker, outbox events.Outbox) *Booking {
	if outbox == nil {
		outbox = events.Nop{}
	}
	return &Booking{Slots: slots, Outbox: outbox}
}

func (b *Booking) Book(ctx context.Context, tenantID, providerID, patientID string, slot time.Time) (string, error) {
	id := uuid.NewString()
	if err := b.Slots.BookSlot(ctx, id, tenantID, providerID, patientID, slot); err != nil {
		return "", err
	}
	requestID := obs.RequestIDFromContext(ctx)
	ev := events.New(events.TypeAppointmentBooked, tenantID, requestID, map[string]any{
		"appointment_id": id,
		"patient_id":     patientID,
		"provider_id":    providerID,
	})
	if err := b.Outbox.Append(ctx, ev); err != nil {
		log.Printf("outbox append failed request_id=%s event_type=%s", requestID, ev.EventType)
	}
	return id, nil
}
