# Architecture

## Overview

Gin exposes JSON APIs consumed by the patient portal and internal admin consoles. The service maintains appointment rows keyed by tenant, provider, and slot start time.

## Components

- **cmd/server** — router wiring, graceful shutdown, obs middleware.
- **internal/handlers** — appointment and admin HTTP adapters.
- **internal/store** — PostgreSQL access with request-scoped logging.
- **internal/obs** — request ID propagation, JSON access logs, `/meta`.
- **internal/config** — env-driven settings including service metadata.

## Booking pipeline

1. Client requests a slot from the availability cache.
2. obs middleware assigns `X-Request-ID` and threads tenant context when present.
3. Handler calls the store `Book` path which inserts a provisional row.
4. Notification worker optionally invokes external webhooks with appointment metadata.

## Admin paths

Administrative routes live under `/v1/admin` and are intended for trusted operator networks. Session cookies are issued by the corporate SSO reverse proxy in full deployments.

## Observability

Structured logs include `request_id` at the handler and store layers. `/meta` exposes build provenance for deploy verification.
