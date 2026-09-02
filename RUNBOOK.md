# Scheduling Service Runbook

## Deployment checklist

- Set `SERVICE_VERSION`, `GIT_SHA`, and `BUILD_TIME` from CI before rollout.
- Confirm ingress forwards `X-Request-ID` from upstream gateways.

## Environment variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `SERVICE_VERSION` | Version on `/meta` | `dev` |
| `GIT_SHA` | Commit SHA on `/meta` | `unknown` |
| `BUILD_TIME` | Build timestamp on `/meta` | `unknown` |

## Capacity

Watch `scheduling_db_conn_wait_seconds` during open enrollment periods. Scale pods before the queue depth exceeds two minutes of sustained growth.

## Webhook failures

Failed delivery attempts are retried with exponential backoff. Operators may replay from the support console using the appointment identifier.

## Database maintenance

Failover follows the standard regional playbook: promote replica, update service DNS weight, drain old primary connections.
