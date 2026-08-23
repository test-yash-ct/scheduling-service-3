CREATE TABLE IF NOT EXISTS appointments (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    patient_id TEXT NOT NULL,
    slot_start TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled'
);

CREATE INDEX IF NOT EXISTS idx_appointments_tenant_provider_slot
    ON appointments (tenant_id, provider_id, slot_start)
    WHERE status = 'scheduled';

CREATE UNIQUE INDEX IF NOT EXISTS uq_appointments_active_slot
    ON appointments (tenant_id, provider_id, slot_start)
    WHERE status = 'scheduled';

CREATE TABLE IF NOT EXISTS booking_idempotency (
    tenant_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    appointment_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_booking_idempotency_appointment
    ON booking_idempotency (appointment_id);
