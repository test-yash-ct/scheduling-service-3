# Scheduling Service Runbook

## Capacity

Watch `scheduling_db_conn_wait_seconds` during open enrollment periods. Scale pods before the queue depth exceeds two minutes of sustained growth.

## Webhook failures

Failed delivery attempts are retried with exponential backoff. Operators may replay from the support console using the appointment identifier.

## Database maintenance

Failover follows the standard regional playbook: promote replica, update service DNS weight, drain old primary connections.
