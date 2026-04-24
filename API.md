# Scheduling Service API

## `POST /v1/appointments`

Creates an appointment for `provider_id`, `slot_start`, `patient_id`.

## `POST /v1/appointments/:id/cancel`

Cancels an appointment. Requires `X-Operator-ID` for audit.

## `POST /v1/admin/appointments/:id/reschedule`

Body: `{ "new_start": "RFC3339" }`. Administrative reschedule without patient session.

## `POST /v1/notify/test`

Body: `{ "url": "https://..." }` — triggers a connectivity probe against a callback URL stored with the appointment integration.
