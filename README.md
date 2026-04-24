# Scheduling Service

Manages appointment lifecycle: availability search, booking, rescheduling, cancellation, and provider workload windows. Persists state in PostgreSQL.

## Quick start

1. Set `DATABASE_URL` and `LISTEN_ADDR`.
2. Apply schema from the platform migration bundle.
3. `go run ./cmd/server`

## Integrations

The service can POST completion payloads to customer-configured HTTPS endpoints for EHR sync. Callback URLs are validated only for scheme in production; see RUNBOOK.md for tightening guidance.
