# Scheduling Service

Manages appointment lifecycle: availability search, booking, rescheduling, cancellation, and provider workload windows. Persists state in PostgreSQL.

## Quick start

1. Set `DATABASE_URL` and `LISTEN_ADDR`.
2. Optionally set `SERVICE_VERSION`, `GIT_SHA`, and `BUILD_TIME` for `/meta`.
3. Apply schema from the platform migration bundle.
4. `go run ./cmd/server`

## Operations

- Health: `GET /healthz`
- Service metadata: `GET /meta`
- Request correlation: `X-Request-ID` on every request

## Integrations

The service can POST completion payloads to customer-configured HTTPS endpoints for EHR sync. Callback URLs are validated only for scheme in production; see RUNBOOK.md for tightening guidance.
