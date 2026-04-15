# Fleet Health Dashboard — Data Schema

This document defines the shared PostgreSQL + TimescaleDB schema used by all Fleet Health Dashboard subsystems. All child work items (WI1–WI6) should reference these table definitions for consistency.

## Database: `fleet_dashboard`

### Extension Requirements

```sql
CREATE EXTENSION IF NOT EXISTS timescaledb;
```

---

## Core Tables (WI2: REST API & Data Store)

### `devices`

Primary device inventory table. Updated on each report ingestion.

| Column           | Type                     | Constraints                | Description                              |
|------------------|--------------------------|----------------------------|------------------------------------------|
| serial_number    | VARCHAR(64)              | PRIMARY KEY                | Unique device serial number              |
| host             | VARCHAR(255)             | NOT NULL, INDEX            | Hostname of the agent that reports this device |
| model            | VARCHAR(255)             | NOT NULL, INDEX            | Device model string                      |
| fw_revision      | VARCHAR(64)              |                            | Current firmware revision                |
| wwn              | VARCHAR(64)              |                            | World Wide Name                          |
| capacity_bytes   | BIGINT                   |                            | Device capacity in bytes                 |
| interface_type   | VARCHAR(16)              | CHECK (IN 'SATA','SAS','NVMe','USB','Unknown') | Storage interface type |
| smart_status     | VARCHAR(16)              | CHECK (IN 'PASS','FAIL','WARNING','UNKNOWN') | Latest SMART status     |
| smart_attributes | JSONB                    |                            | Raw SMART attribute key-value pairs      |
| farm_data        | JSONB                    |                            | Latest FARM reliability metrics JSON     |
| fw_download_support | VARCHAR(16)           |                            | Firmware download support mode           |
| first_seen       | TIMESTAMPTZ              | NOT NULL, DEFAULT NOW()    | First time device was reported           |
| last_seen        | TIMESTAMPTZ              | NOT NULL, DEFAULT NOW()    | Most recent report timestamp             |

**Indexes:**
- `idx_devices_host` ON (host)
- `idx_devices_model` ON (model)
- `idx_devices_smart_status` ON (smart_status)
- `idx_devices_interface_type` ON (interface_type)
- `idx_devices_fw_revision` ON (fw_revision)

---

### `device_metrics` (TimescaleDB Hypertable)

Time-series metrics table for device telemetry. Converted to a TimescaleDB hypertable partitioned by `time`.

| Column                    | Type          | Constraints             | Description                              |
|---------------------------|---------------|-------------------------|------------------------------------------|
| time                      | TIMESTAMPTZ   | NOT NULL                | Metric collection timestamp              |
| serial_number             | VARCHAR(64)   | NOT NULL, FK → devices  | Device serial number                     |
| temperature_c             | REAL          |                         | Drive temperature in Celsius             |
| power_on_hours            | INTEGER       |                         | Cumulative power-on hours                |
| unrecoverable_read_errors | INTEGER       |                         | Unrecoverable read error count           |
| unrecoverable_write_errors| INTEGER       |                         | Unrecoverable write error count          |
| reallocated_sectors       | INTEGER       |                         | Reallocated sector count                 |
| workload_tb_yr            | REAL          |                         | Annualized workload in TB/year           |

**Hypertable config:**
```sql
SELECT create_hypertable('device_metrics', 'time');
```

**Indexes:**
- `idx_metrics_serial_time` ON (serial_number, time DESC)

**Retention:** 1 year default (configurable)

---

### `hosts`

Derived/cached host summary table. Updated on each report ingestion.

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| host           | VARCHAR(255)  | PRIMARY KEY             | Hostname                                 |
| device_count   | INTEGER       | NOT NULL, DEFAULT 0     | Number of devices on this host           |
| last_check_in  | TIMESTAMPTZ   | NOT NULL                | Last report timestamp from this host     |
| poll_interval_s| INTEGER       | DEFAULT 900             | Configured poll interval in seconds      |

---

## Compliance Tables (WI3: Firmware Compliance)

### `policies`

Firmware compliance policy definitions.

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| id             | SERIAL        | PRIMARY KEY             | Auto-increment policy ID                 |
| model_pattern  | VARCHAR(255)  | NOT NULL                | Glob pattern matching device model numbers |
| approved_fw    | JSONB         | NOT NULL                | JSON array of approved firmware revision strings |
| severity       | VARCHAR(16)   | NOT NULL, CHECK (IN 'critical','warning','info') | Policy violation severity |
| created_at     | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | Policy creation timestamp                |
| updated_at     | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | Last modification timestamp              |

---

### `device_compliance`

Per-device compliance evaluation results. Updated when reports are ingested.

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| serial_number  | VARCHAR(64)   | NOT NULL, FK → devices  | Device serial number                     |
| policy_id      | INTEGER       | NOT NULL, FK → policies | Matched policy ID                        |
| is_compliant   | BOOLEAN       | NOT NULL                | Whether device firmware matches an approved version |
| current_fw     | VARCHAR(64)   |                         | Device's current firmware at evaluation time |
| evaluated_at   | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | When compliance was last evaluated       |

**Primary Key:** (serial_number, policy_id)

**Indexes:**
- `idx_compliance_noncompliant` ON (is_compliant) WHERE is_compliant = FALSE

---

## Alert Tables (WI4: Alerting Engine)

### `alert_rules`

Alert rule definitions (both built-in and custom).

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| id             | SERIAL        | PRIMARY KEY             | Auto-increment rule ID                   |
| name           | VARCHAR(255)  | NOT NULL, UNIQUE        | Human-readable rule name                 |
| rule_type      | VARCHAR(32)   | NOT NULL                | One of: smart_tripped, smart_warning, temperature_exceeded, firmware_non_compliant, agent_stale, error_rate_spike, custom_threshold |
| severity       | VARCHAR(16)   | NOT NULL, CHECK (IN 'critical','warning','info') | Alert severity |
| config         | JSONB         |                         | Rule-specific config (e.g. `{"threshold": 60}`) |
| enabled        | BOOLEAN       | NOT NULL, DEFAULT TRUE  | Whether the rule is active               |
| created_at     | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | Rule creation timestamp                  |

---

### `alerts`

Fired alert instances.

| Column           | Type          | Constraints             | Description                              |
|------------------|---------------|-------------------------|------------------------------------------|
| id               | SERIAL        | PRIMARY KEY             | Auto-increment alert ID                  |
| serial_number    | VARCHAR(64)   | FK → devices            | Device that triggered the alert (NULL for host-level alerts) |
| host             | VARCHAR(255)  |                         | Host associated with the alert           |
| rule_id          | INTEGER       | NOT NULL, FK → alert_rules | Rule that fired this alert            |
| severity         | VARCHAR(16)   | NOT NULL                | Severity at time of firing               |
| status           | VARCHAR(16)   | NOT NULL, DEFAULT 'active', CHECK (IN 'active','resolved','acknowledged','snoozed') | Alert status |
| message          | TEXT          |                         | Human-readable alert description         |
| fired_at         | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | When the alert was created               |
| resolved_at      | TIMESTAMPTZ   |                         | When the alert was auto-resolved         |
| acknowledged_at  | TIMESTAMPTZ   |                         | When a user acknowledged the alert       |
| snoozed_until    | TIMESTAMPTZ   |                         | Snooze expiration time                   |

**Indexes:**
- `idx_alerts_status` ON (status)
- `idx_alerts_serial` ON (serial_number)
- `idx_alerts_rule_serial` ON (rule_id, serial_number) — for deduplication lookups
- `idx_alerts_fired_at` ON (fired_at) — for retention cleanup

**Retention:** 90 days (delete alerts older than 90 days)

**Deduplication:** UNIQUE constraint on (serial_number, rule_id) WHERE status = 'active' to prevent duplicate active alerts.

---

### `notification_channels`

Notification dispatch configuration.

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| id             | SERIAL        | PRIMARY KEY             | Auto-increment channel ID                |
| name           | VARCHAR(255)  | NOT NULL                | Channel display name                     |
| channel_type   | VARCHAR(16)   | NOT NULL, CHECK (IN 'webhook','email','syslog') | Dispatch method |
| config         | JSONB         | NOT NULL                | Channel-specific config (see OpenAPI spec for shapes) |
| enabled        | BOOLEAN       | NOT NULL, DEFAULT TRUE  | Whether channel is active                |

---

## Analytics Tables (WI6: Trend Analysis & Predictive Health)

### `device_health_scores`

Computed health scores per device. Recomputed every 15 minutes for devices with new data.

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| serial_number  | VARCHAR(64)   | PRIMARY KEY, FK → devices | Device serial number                   |
| score          | REAL          | NOT NULL, CHECK (0 <= score <= 100) | Composite health score (0–100) |
| trend          | VARCHAR(16)   | NOT NULL, CHECK (IN 'improving','stable','declining') | Score trend direction |
| factors        | JSONB         | NOT NULL                | Breakdown of contributing factor scores and weights |
| computed_at    | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | When this score was last computed        |

**Indexes:**
- `idx_health_score` ON (score)
- `idx_health_trend` ON (trend)

---

## Health Score Algorithm (WI6 Reference)

The composite health score (0–100) is computed as a weighted sum of four component scores:

| Factor                | Weight | Description                                              |
|-----------------------|--------|----------------------------------------------------------|
| SMART Status          | 0.30   | PASS=100, WARNING=50, FAIL=0, UNKNOWN=50                |
| Error Rate Trends     | 0.25   | Based on 7d/30d rate of change for read/write errors     |
| Temperature Deviation | 0.15   | Deviation from fleet median temperature (lower = better) |
| FARM Reliability      | 0.30   | Helium pressure trips, disc slip recals, ECC counts      |

Weights are configurable via a config file. The `factors` JSONB column stores per-factor scores.

**Trend detection:** A device is "declining" if the 7-day rolling average score decreased by ≥ 5 points. "Improving" if it increased by ≥ 5 points. Otherwise "stable".

**Sparse data handling:** If a device has missed reports (no data in last 2x poll interval), the previous score is retained without penalty. Only actual metric changes trigger re-scoring.

---

## Entity Relationship Summary

```
devices (serial_number PK)
  ├── device_metrics (serial_number FK, time) — hypertable
  ├── device_compliance (serial_number FK, policy_id FK)
  ├── alerts (serial_number FK, rule_id FK)
  └── device_health_scores (serial_number PK/FK)

policies (id PK)
  └── device_compliance (policy_id FK)

alert_rules (id PK)
  └── alerts (rule_id FK)

hosts (host PK)
  └── referenced by devices.host (logical, not FK enforced)
```

---

## API Key Table (shared)

### `api_keys`

| Column         | Type          | Constraints             | Description                              |
|----------------|---------------|-------------------------|------------------------------------------|
| key_hash       | VARCHAR(128)  | PRIMARY KEY             | SHA-256 hash of the API key              |
| name           | VARCHAR(255)  | NOT NULL                | Human-readable key name                  |
| created_at     | TIMESTAMPTZ   | NOT NULL, DEFAULT NOW() | Key creation timestamp                   |
| last_used_at   | TIMESTAMPTZ   |                         | Last time this key was used              |
| is_active      | BOOLEAN       | NOT NULL, DEFAULT TRUE  | Whether the key is active                |
