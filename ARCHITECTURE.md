# Architecture

## Overview

Gin exposes JSON APIs consumed by the patient portal and internal admin consoles. The service maintains appointment rows keyed by tenant, provider, and slot start time.

## Components

- **cmd/server** — router wiring, graceful shutdown, obs middleware.
- **internal/handlers** — appointment and admin HTTP adapters.
- **internal/store** — PostgreSQL access with request-scoped logging.
- **internal/obs** — request ID propagation, JSON access logs, `/meta`.
- **internal/config** — env-driven settings including service metadata.
- **internal/service** — thin booking use case that emits `appointment.booked` after a successful store write.
- **internal/events** — typed integration-event envelope and in-process outbox.

## Booking pipeline

1. Client requests a slot from the availability cache.
2. obs middleware assigns `X-Request-ID` and threads tenant context when present.
3. Handler calls `service.Booking.Book`, which inserts via the store then appends `appointment.booked` to the in-process outbox (payload: `appointment_id`, `patient_id`, `provider_id` only).
4. Notification worker optionally invokes external webhooks with appointment metadata.

## Integration events

Services emit typed domain events after successful writes instead of expecting peers to scrape HTTP APIs.

Envelope:

```json
{"event_id":"uuid","event_type":"...","tenant_id":"...","occurred_at":"RFC3339","request_id":"...","payload":{}}
```

Transport today is an in-process `Outbox` (`Memory`). A durable dispatcher can be added later without changing producers. Events are not appended when `BookSlot` fails (including slot-taken).

## Admin paths

Administrative routes live under `/v1/admin` and are intended for trusted operator networks. Session cookies are issued by the corporate SSO reverse proxy in full deployments.

## Observability

Structured logs include `request_id` at the handler and store layers. `/meta` exposes build provenance for deploy verification.
