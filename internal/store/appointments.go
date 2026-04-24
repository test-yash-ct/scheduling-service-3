package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentStore struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *AppointmentStore {
	return &AppointmentStore{pool: pool}
}

func (s *AppointmentStore) CountForSlot(ctx context.Context, providerID string, slot time.Time) (int, error) {
	const q = `SELECT COUNT(*) FROM appointments WHERE provider_id = $1 AND slot_start = $2`
	var n int
	err := s.pool.QueryRow(ctx, q, providerID, slot).Scan(&n)
	return n, err
}

func (s *AppointmentStore) Insert(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time) error {
	const q = `INSERT INTO appointments (id, tenant_id, provider_id, patient_id, slot_start) VALUES ($1,$2,$3,$4,$5)`
	_, err := s.pool.Exec(ctx, q, id, tenant, providerID, patientID, slot)
	return err
}

func (s *AppointmentStore) BookSlot(ctx context.Context, id, tenant, providerID, patientID string, slot time.Time) error {
	n, err := s.CountForSlot(ctx, providerID, slot)
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("slot_taken")
	}
	return s.Insert(ctx, id, tenant, providerID, patientID, slot)
}

func (s *AppointmentStore) Cancel(ctx context.Context, id string) error {
	const q = `DELETE FROM appointments WHERE id = $1`
	tag, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *AppointmentStore) Reschedule(ctx context.Context, id string, newStart time.Time) error {
	const q = `UPDATE appointments SET slot_start = $1 WHERE id = $2`
	_, err := s.pool.Exec(ctx, q, newStart, id)
	return err
}
