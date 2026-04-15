# OpenSeaChest Fleet Health Dashboard — Alerting & Notification Engine

This module implements the alerting rules engine and notification delivery system for the OpenSeaChest Fleet Health Dashboard. It monitors storage device telemetry and triggers alerts based on configurable rules, with support for multiple notification channels.

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────────┐
│  Device Agents   │────▶│  Rule Engine  │────▶│  Notification Svc   │
│  (telemetry)     │     │  (evaluate)   │     │  (webhook / email)  │
└─────────────────┘     └──────┬───────┘     └─────────────────────┘
                               │
                        ┌──────▼───────┐
                        │    Store     │
                        │  (alerts,   │
                        │   rules,    │
                        │   history)  │
                        └─────────────┘
```

### Components

- **Rule Engine** (`api/services/rule_engine.go`): Evaluates incoming device reports against all enabled alert rules. Supports six built-in rule types with pluggable configuration.
- **Store** (`api/services/store.go`): In-memory data store for alert rules, alerts, alert history, and notification channels. In production, this is backed by PostgreSQL (see migration).
- **Notification Service** (`api/services/notification.go`): Dispatches alert notifications via configured channels (webhook with retry, email via SMTP stub).
- **HTTP Handlers** (`api/handlers/`): REST API handlers for rule CRUD, alert management, and rule evaluation.

## Built-in Rule Types

| Rule Type | Description | Configuration |
|-----------|-------------|---------------|
| `smart_trip` | SMART self-test failure detected. Maps to openSeaChest `UTIL_EXIT_OPERATION_FAILURE` exit code. | No additional config required. |
| `smart_warning` | SMART check returns warning or in-progress status. | No additional config required. |
| `temperature_threshold` | Device temperature exceeds configured threshold. Includes debounce logic. | `threshold_celsius` (float), `consecutive_count` (int, default: 2) |
| `firmware_non_compliance` | Device firmware doesn't match required policy version. | `required_firmware` (string) |
| `agent_stale` | Agent hasn't reported within the configured interval. | `stale_interval_seconds` (int) |
| `reallocated_sector_growth` | Reallocated sector count is increasing beyond threshold. | `growth_threshold` (int, default: 1) |

### Rule Configuration Examples

**Temperature Threshold:**
```json
{
  "threshold_celsius": 55.0,
  "consecutive_count": 2
}
```

**Firmware Non-Compliance:**
```json
{
  "required_firmware": "SN06"
}
```

**Agent Stale:**
```json
{
  "stale_interval_seconds": 300
}
```

**Reallocated Sector Growth:**
```json
{
  "growth_threshold": 5
}
```

## Alert State Machine

Alerts follow a strict state machine with the following transitions:

```
     condition triggers
 OK ──────────────────▶ FIRING
                          │
              user ack    │    condition clears
                ▼         │         │
          ACKNOWLEDGED    │         │
              │           │         │
              ▼           ▼         ▼
           RESOLVED ◀────────────────
                │
                │  can re-fire
                ▼
             FIRING
```

### States

| State | Description |
|-------|-------------|
| `ok` | No active condition. Initial state. |
| `firing` | Alert condition has been triggered and is active. |
| `acknowledged` | A user has acknowledged the firing alert. |
| `resolved` | The condition has cleared or the alert was manually resolved. |

### Valid Transitions

- `OK` → `FIRING`: Condition triggers
- `FIRING` → `ACKNOWLEDGED`: User acknowledges
- `FIRING` → `RESOLVED`: Condition clears or manual resolve
- `ACKNOWLEDGED` → `RESOLVED`: Condition clears or manual resolve
- `RESOLVED` → `FIRING`: Condition re-triggers after resolution

## Notification Channels

### Webhook

Sends the alert payload as a JSON POST request to a configured URL.

**Features:**
- Retry with exponential backoff: max 3 retries, backoff intervals of 1s, 2s, 4s
- Custom headers support
- Client errors (4xx) are not retried
- Server errors (5xx) and network errors trigger retries

**Configuration:**
```json
{
  "url": "https://example.com/webhook",
  "headers": {
    "Authorization": "Bearer <token>"
  },
  "secret": "optional-signing-secret"
}
```

### Email (SMTP Stub)

Sends formatted alert emails via SMTP. This is a stub implementation that connects to the configured SMTP server.

**Configuration:**
```json
{
  "smtp_host": "smtp.example.com",
  "smtp_port": 587,
  "from_address": "alerts@example.com",
  "to_addresses": ["ops@example.com", "oncall@example.com"],
  "username": "alerts@example.com",
  "password": "secret",
  "use_tls": true
}
```

### Per-Rule Channel Assignment

Notification channels are configurable per alert rule via the `notification_channels` field, which accepts an array of channel IDs.

## Alert Deduplication

If the same condition fires on the same device and an active alert already exists, the engine updates the `last_seen` timestamp instead of creating a duplicate alert.

**Deduplication key:** `(rule_id, device_serial)`

This is enforced at both the application level and the database level (via a partial unique index on the `alerts` table).

## Temperature Alert Debounce

Temperature alerts include debounce logic to prevent transient spikes from generating noise:

- Only fires after **2 consecutive intervals** above the threshold (configurable via `consecutive_count`)
- If temperature drops below the threshold at any point, the consecutive counter resets
- This ensures only sustained temperature issues generate alerts

## API Endpoints

### Alert Rule CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/alerts/rules` | Create a new alert rule |
| `GET` | `/api/v1/alerts/rules` | List all alert rules |
| `GET` | `/api/v1/alerts/rules/{id}` | Get a specific rule |
| `PUT` | `/api/v1/alerts/rules/{id}` | Update an existing rule |
| `DELETE` | `/api/v1/alerts/rules/{id}` | Delete a rule |

### Alert Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/alerts/active` | List all currently active (FIRING or ACKNOWLEDGED) alerts |
| `GET` | `/api/v1/alerts/history` | List historical alerts with pagination (`?page=1&page_size=50`) |
| `POST` | `/api/v1/alerts/{id}/acknowledge` | Acknowledge a firing alert |
| `POST` | `/api/v1/alerts/{id}/resolve` | Manually resolve an alert |
| `POST` | `/api/v1/alerts/evaluate` | Submit a device report for rule evaluation (testing) |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Server health check |

## Usage

### Create an Alert Rule

```bash
curl -X POST http://localhost:8080/api/v1/alerts/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Temperature Alert",
    "rule_type": "temperature_threshold",
    "config": {"threshold_celsius": 55.0, "consecutive_count": 2},
    "severity": "warning",
    "enabled": true,
    "notification_channels": ["<channel-id>"]
  }'
```

### Submit a Device Report for Evaluation

```bash
curl -X POST http://localhost:8080/api/v1/alerts/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "device_serial": "WD-ABC123",
    "smart_status": "passed",
    "temperature_celsius": 58.0,
    "current_firmware": "SN06",
    "last_reported_at": "2026-01-15T10:00:00Z",
    "reallocated_sector_count": 5,
    "prev_reallocated_count": 3
  }'
```

### Acknowledge an Alert

```bash
curl -X POST http://localhost:8080/api/v1/alerts/<alert-id>/acknowledge \
  -H "Content-Type: application/json" \
  -d '{"acknowledged_by": "admin@example.com"}'
```

### Resolve an Alert

```bash
curl -X POST http://localhost:8080/api/v1/alerts/<alert-id>/resolve \
  -H "Content-Type: application/json" \
  -d '{"resolved_by": "admin@example.com"}'
```

## Database Schema

The database schema is defined in `migrations/001_alerting.sql` and includes:

- **`alert_rules`** — Rule definitions with JSONB configuration
- **`alerts`** — Alert instances with state tracking and deduplication
- **`alert_history`** — Audit trail of state transitions
- **`notification_channels`** — Notification delivery channel configurations

See the migration file for full schema details, indexes, and constraints.

## Building

```bash
cd dashboard
go build -o bin/alerting-server ./cmd/server
```

## Running

```bash
# Default port 8080
./bin/alerting-server

# Custom port
PORT=9090 ./bin/alerting-server
```

## openSeaChest Exit Code Mapping

The alert rule types map to openSeaChest CLI exit codes defined in `include/openseachest_util_options.h`:

| Exit Code | Constant | Alert Rule Type |
|-----------|----------|-----------------|
| 0 | `UTIL_EXIT_NO_ERROR` | No alert (OK state) |
| 3 | `UTIL_EXIT_OPERATION_FAILURE` | `smart_trip` (self-test failure) |
| 4 | `UTIL_EXIT_OPERATION_NOT_SUPPORTED` | N/A |
| 5 | `UTIL_EXIT_OPERATION_ABORTED` | `smart_warning` (check interrupted) |
