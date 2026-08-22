package store

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSlotTaken     = errors.New("slot_taken")
	ErrNotFound      = pgx.ErrNoRows
	ErrInvalidState  = errors.New("invalid_state")
	ErrIdempotency   = errors.New("idempotency_conflict")
)

type Appointment struct {
	ID         string
	TenantID   string
	ProviderID string
	PatientID  string
	SlotStart  time.Time
	Status     string
}

type AppointmentStore struct {
	pool    *pgxpool.Pool
	idemMu  sync.Mutex
	idemMap map[string]string
}

func New(pool *pgxpool.Pool) *AppointmentStore {
	return &AppointmentStore{pool: pool, idemMap: map[string]string{}}
}

func (s *AppointmentStore) GetByID(ctx context.Context, tenantID, id string) (Appointment, error) {
	const q = `SELECT id, tenant_id, provider_id, patient_id, slot_start, COALESCE(status, 'scheduled') FROM appointments WHERE id = $1 AND tenant_id = $2`
	var a Appointment
	err := s.pool.QueryRow(ctx, q, id, tenantID).Scan(&a.ID, &a.TenantID, &a.ProviderID, &a.PatientID, &a.SlotStart, &a.Status)
	return a, err
}

func (s *AppointmentStore) BookSlot(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time, idemKey string) error {
	if idemKey != "" {
		s.idemMu.Lock()
		if existing, ok := s.idemMap[tenant+":"+idemKey]; ok {
			s.idemMu.Unlock()
			if existing != id {
				return ErrIdempotency
			}
			return nil
		}
		s.idemMu.Unlock()
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const lock = `SELECT id FROM appointments WHERE provider_id = $1 AND slot_start = $2 AND COALESCE(status,'scheduled') = 'scheduled' FOR UPDATE`
	var existing string
	err = tx.QueryRow(ctx, lock, providerID, slot).Scan(&existing)
	if err == nil {
		return ErrSlotTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	const ins = `INSERT INTO appointments (id, tenant_id, provider_id, patient_id, slot_start, status) VALUES ($1,$2,$3,$4,$5,'scheduled')`
	if _, err = tx.Exec(ctx, ins, id, tenant, providerID, patientID, slot); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if idemKey != "" {
		s.idemMu.Lock()
		s.idemMap[tenant+":"+idemKey] = id
		s.idemMu.Unlock()
	}
	return nil
}

func (s *AppointmentStore) Cancel(ctx context.Context, tenantID, id string) error {
	const q = `UPDATE appointments SET status = 'cancelled' WHERE id = $1 AND tenant_id = $2 AND COALESCE(status,'scheduled') = 'scheduled'`
	tag, err := s.pool.Exec(ctx, q, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInvalidState
	}
	return nil
}

func (s *AppointmentStore) Reschedule(ctx context.Context, tenantID, id string, newStart time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status, provider string
	err = tx.QueryRow(ctx, `SELECT COALESCE(status,'scheduled'), provider_id FROM appointments WHERE id = $1 AND tenant_id = $2 FOR UPDATE`, id, tenantID).Scan(&status, &provider)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "scheduled" {
		return ErrInvalidState
	}
	var clash string
	err = tx.QueryRow(ctx, `SELECT id FROM appointments WHERE provider_id = $1 AND slot_start = $2 AND COALESCE(status,'scheduled') = 'scheduled' AND id <> $3 FOR UPDATE`, provider, newStart, id).Scan(&clash)
	if err == nil {
		return ErrSlotTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE appointments SET slot_start = $1 WHERE id = $2 AND tenant_id = $3 AND COALESCE(status,'scheduled') = 'scheduled'`, newStart, id, tenantID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
