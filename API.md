# Scheduling Service API

## Request correlation

Supply `X-Request-ID` to correlate logs across services. The response echoes the same header. Access logs are emitted as JSON with `request_id`, `tenant`, `method`, `path`, `status`, and `duration_ms`.

## `GET /meta`

Returns `service`, `version`, `build_time`, and `git_sha` from `SERVICE_VERSION`, `BUILD_TIME`, and `GIT_SHA`.

## `POST /v1/appointments`

Creates an appointment for `provider_id`, `slot_start`, `patient_id`.

## `POST /v1/appointments/:id/cancel`

Cancels an appointment. Requires `X-Operator-ID` for audit.

## `POST /v1/admin/appointments/:id/reschedule`

Body: `{ "new_start": "RFC3339" }`. Administrative reschedule without patient session.

## `POST /v1/notify/test`

Body: `{ "url": "https://..." }` — triggers a connectivity probe against a callback URL stored with the appointment integration.
