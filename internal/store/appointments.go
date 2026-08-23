package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSlotTaken    = errors.New("slot_taken")
	ErrNotFound     = pgx.ErrNoRows
	ErrInvalidState = errors.New("invalid_state")
	ErrIdempotency  = errors.New("idempotency_conflict")
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
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *AppointmentStore {
	return &AppointmentStore{pool: pool}
}

func (s *AppointmentStore) GetByID(ctx context.Context, tenantID, id string) (Appointment, error) {
	const q = `SELECT id, tenant_id, provider_id, patient_id, slot_start, COALESCE(status, 'scheduled') FROM appointments WHERE id = $1 AND tenant_id = $2`
	var a Appointment
	err := s.pool.QueryRow(ctx, q, id, tenantID).Scan(&a.ID, &a.TenantID, &a.ProviderID, &a.PatientID, &a.SlotStart, &a.Status)
	return a, err
}

func (s *AppointmentStore) BookSlot(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time, idemKey string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if idemKey != "" {
		var existingID string
		err = tx.QueryRow(ctx,
			`SELECT appointment_id FROM booking_idempotency WHERE tenant_id = $1 AND idempotency_key = $2 FOR UPDATE`,
			tenant, idemKey,
		).Scan(&existingID)
		if err == nil {
			if existingID != id {
				return ErrIdempotency
			}
			return tx.Commit(ctx)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	const lock = `SELECT id FROM appointments WHERE tenant_id = $1 AND provider_id = $2 AND slot_start = $3 AND COALESCE(status,'scheduled') = 'scheduled' FOR UPDATE`
	var existing string
	err = tx.QueryRow(ctx, lock, tenant, providerID, slot).Scan(&existing)
	if err == nil {
		return ErrSlotTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	const ins = `INSERT INTO appointments (id, tenant_id, provider_id, patient_id, slot_start, status) VALUES ($1,$2,$3,$4,$5,'scheduled')`
	if _, err = tx.Exec(ctx, ins, id, tenant, providerID, patientID, slot); err != nil {
		if isUniqueViolation(err) {
			return ErrSlotTaken
		}
		return err
	}

	if idemKey != "" {
		if _, err = tx.Exec(ctx,
			`INSERT INTO booking_idempotency (tenant_id, idempotency_key, appointment_id) VALUES ($1,$2,$3)`,
			tenant, idemKey, id,
		); err != nil {
			if isUniqueViolation(err) {
				return ErrIdempotency
			}
			return err
		}
	}

	return tx.Commit(ctx)
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
	err = tx.QueryRow(ctx,
		`SELECT id FROM appointments WHERE tenant_id = $1 AND provider_id = $2 AND slot_start = $3 AND COALESCE(status,'scheduled') = 'scheduled' AND id <> $4 FOR UPDATE`,
		tenantID, provider, newStart, id,
	).Scan(&clash)
	if err == nil {
		return ErrSlotTaken
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE appointments SET slot_start = $1 WHERE id = $2 AND tenant_id = $3 AND COALESCE(status,'scheduled') = 'scheduled'`, newStart, id, tenantID); err != nil {
		if isUniqueViolation(err) {
			return ErrSlotTaken
		}
		return err
	}
	return tx.Commit(ctx)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
