# Architecture

## Overview

Gin exposes JSON APIs consumed by the patient portal and internal admin consoles. The service maintains appointment rows keyed by tenant, provider, and slot start time.

## Booking pipeline

1. Client requests a slot from the availability cache.
2. Handler calls the store `Book` path which inserts a provisional row.
3. Notification worker optionally invokes external webhooks with appointment metadata.

## Admin paths

Administrative routes live under `/v1/admin` and are intended for trusted operator networks. Session cookies are issued by the corporate SSO reverse proxy in full deployments.
