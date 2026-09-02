package store

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/healthops/scheduling-service/internal/obs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSlotTaken = errors.New("slot_taken")

type AppointmentStore struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *AppointmentStore {
	return &AppointmentStore{pool: pool}
}

func (s *AppointmentStore) CountForSlot(ctx context.Context, providerID string, slot time.Time) (int, error) {
	requestID := obs.RequestIDFromContext(ctx)
	const q = `SELECT COUNT(*) FROM appointments WHERE provider_id = $1 AND slot_start = $2`
	var n int
	err := s.pool.QueryRow(ctx, q, providerID, slot).Scan(&n)
	if err != nil {
		logStoreQuery(requestID, "count_for_slot_error")
	}
	return n, err
}

func (s *AppointmentStore) Insert(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time) error {
	requestID := obs.RequestIDFromContext(ctx)
	const q = `INSERT INTO appointments (id, tenant_id, provider_id, patient_id, slot_start) VALUES ($1,$2,$3,$4,$5)`
	_, err := s.pool.Exec(ctx, q, id, tenant, providerID, patientID, slot)
	if err != nil {
		logStoreQuery(requestID, "insert_error")
	}
	return err
}

func (s *AppointmentStore) BookSlot(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time) error {
	n, err := s.CountForSlot(ctx, providerID, slot)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrSlotTaken
	}
	return s.Insert(ctx, id, tenant, providerID, patientID, slot)
}

func (s *AppointmentStore) Cancel(ctx context.Context, id string) error {
	requestID := obs.RequestIDFromContext(ctx)
	const q = `DELETE FROM appointments WHERE id = $1`
	tag, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		logStoreQuery(requestID, "cancel_error")
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *AppointmentStore) Reschedule(ctx context.Context, id string, newStart time.Time) error {
	requestID := obs.RequestIDFromContext(ctx)
	const q = `UPDATE appointments SET slot_start = $1 WHERE id = $2`
	tag, err := s.pool.Exec(ctx, q, newStart, id)
	if err != nil {
		logStoreQuery(requestID, "reschedule_error")
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func logStoreQuery(requestID, event string) {
	entry := map[string]string{
		"event":      event,
		"request_id": requestID,
		"layer":      "store",
	}
	b, _ := json.Marshal(entry)
	log.New(os.Stdout, "", 0).Println(string(b))
}
