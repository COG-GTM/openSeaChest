# openSeaChest Fleet Health Dashboard

Health monitoring and trend analysis for storage devices managed by [openSeaChest](https://github.com/Seagate/openSeaChest) CLI utilities.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   REST API Server                        │
│  GET /api/v1/devices/{serial}/health                    │
│  GET /api/v1/devices/{serial}/health/history            │
│  GET /api/v1/fleet/health/summary                       │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              TimescaleDB (health_metrics)                │
│  Hypertable partitioned by time (7-day chunks)          │
└────────────────────┬────────────────────────────────────┘
                     ▲
                     │
┌────────────────────┴────────────────────────────────────┐
│                  Health Agent                            │
│  ┌──────────────┐ ┌───────────────┐ ┌────────────────┐  │
│  │ SMART Check  │ │ SMART Attrs   │ │ FARM Log       │  │
│  │ Collector    │ │ Collector     │ │ Collector      │  │
│  └──────┬───────┘ └───────┬───────┘ └───────┬────────┘  │
│         │                 │                  │           │
│         ▼                 ▼                  ▼           │
│  openSeaChest_SMART  openSeaChest_SMART  openSeaChest_  │
│  --smartCheck        --smartAttributes   Logs --farm    │
│                      hybrid              --logMode pipe │
└─────────────────────────────────────────────────────────┘
```

The dashboard is **entirely additive** — it does not modify any existing openSeaChest source code. It shells out to the compiled openSeaChest binaries and parses their stdout output.

## Directory Structure

```
dashboard/
├── main.go                 # API server entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency checksums
├── README.md               # This file
├── agent/
│   ├── collector.go        # Orchestrates all collectors for a device
│   ├── smart_check.go      # Wraps openSeaChest_SMART --smartCheck
│   ├── smart_attributes.go # Wraps openSeaChest_SMART --smartAttributes hybrid
│   └── farm_collector.go   # Wraps openSeaChest_Logs --farm; includes FARM JSON parser
├── api/
│   ├── handlers.go         # HTTP request handlers for REST endpoints
│   ├── server.go           # Server configuration and graceful shutdown
│   └── store.go            # TimescaleDB data access layer
├── models/
│   └── health.go           # Shared data structures and types
└── migrations/
    └── 001_health_metrics.sql  # TimescaleDB hypertable schema
```

## Health Data Collection

### How the Agent Works

The agent collects health data by executing openSeaChest CLI tools as subprocesses and parsing their output. It does **not** link against C libraries — it shells out to the compiled binaries and reads stdout.

### SMART Check (`agent/smart_check.go`)

Runs `openSeaChest_SMART --smartCheck` and maps the process exit code to a health status:

| Exit Code | Constant (from `eUtilExitCodes`) | Health Status |
|-----------|----------------------------------|---------------|
| 0         | `UTIL_EXIT_NO_ERROR`             | `healthy`     |
| 5         | `UTIL_EXIT_OPERATION_ABORTED`    | `warning`     |
| 3         | `UTIL_EXIT_OPERATION_FAILURE`    | `critical`    |
| Other     | —                                | `unknown`     |

The exit codes are defined in `include/openseachest_util_options.h`.

### SMART Attributes (`agent/smart_attributes.go`)

Runs `openSeaChest_SMART --smartAttributes hybrid` and parses the tabular output. Each SMART attribute produces two metrics:

- `smart_attr_{name}_raw` — the raw attribute value
- `smart_attr_{name}_normalized` — the normalized value (0–100 scale)

### FARM Log Collection (`agent/farm_collector.go`)

Runs `openSeaChest_Logs --farm --logMode pipe` which outputs JSON to stdout. The JSON structure matches `example/fromPipe.json` in the repository.

## FARM Log Parsing Details

The FARM (Field Accessible Reliability Metrics) log parser extracts these key fields from the JSON output:

| JSON Section | Field | Metric Name | Unit |
|---|---|---|---|
| Drive Information | `Power on Hour` | `farm_power_on_hours` | hours |
| Drive Information | `Power Cycle count` | `farm_power_cycle_count` | count |
| Environment Information | `Current Temperature (Celsius)` | `farm_current_temperature` | celsius |
| Environment Information | `Highest Temperature` | `farm_highest_temperature` | celsius |
| Environment Information | `Lowest Temperature` | `farm_lowest_temperature` | celsius |
| Error Information | `Unrecoverable Read Errors` | `farm_unrecoverable_read_errors` | count |
| Error Information | `Unrecoverable Write Errors` | `farm_unrecoverable_write_errors` | count |
| Workload | `Rated Workload Percentage` | `farm_rated_workload_percentage` | percent |
| Reliability Information | `Helium Pressure Threshold Tripped` | `farm_helium_pressure_threshold_tripped` | boolean |
| Actuator Information | `Head Load Events` | `farm_head_load_events` | count |
| Actuator Information | `Number of Reallocated Sector Reclamations` | `farm_reallocated_sector_reclamations` | count |
| Error Information | `Number of Mechanical Start Failures` | `farm_mechanical_start_failures` | count |

**Note on FARM JSON quirks:** Some fields in the `fromPipe.json` output contain concatenated string artifacts (e.g., `"Current Temperature (Celsius)": "Environment Information From Farm Log copy 037"` where `37` is the actual value). The parser handles this by extracting leading numeric values from strings.

## API Endpoints

### `GET /api/v1/devices/{serial}/health`

Returns the current health summary for a specific device.

**Response:**
```json
{
  "device_serial": "ZRS004C2",
  "status": "healthy",
  "last_checked": "2025-01-15T10:30:00Z",
  "metrics": [
    {
      "device_serial": "ZRS004C2",
      "timestamp": "2025-01-15T10:30:00Z",
      "metric_name": "smart_check_status",
      "metric_value": 0,
      "unit": "status",
      "source": "smart_check"
    }
  ]
}
```

### `GET /api/v1/devices/{serial}/health/history?from=&to=&metric=`

Returns time-series health metrics for a device within a date range.

**Query Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `from`    | Yes      | Start time in RFC3339 format (e.g., `2025-01-01T00:00:00Z`) |
| `to`      | Yes      | End time in RFC3339 format |
| `metric`  | No       | Filter to a specific metric name (e.g., `farm_current_temperature`) |

**Response:**
```json
{
  "device_serial": "ZRS004C2",
  "from": "2025-01-01T00:00:00Z",
  "to": "2025-01-15T00:00:00Z",
  "metric": "farm_current_temperature",
  "data_points": [
    {
      "device_serial": "ZRS004C2",
      "timestamp": "2025-01-01T10:00:00Z",
      "metric_name": "farm_current_temperature",
      "metric_value": 37,
      "unit": "celsius",
      "source": "farm_log"
    }
  ]
}
```

### `GET /api/v1/fleet/health/summary`

Returns fleet-wide rollup showing aggregate health counts.

**Response:**
```json
{
  "total_devices": 50,
  "healthy_count": 42,
  "warning_count": 5,
  "critical_count": 2,
  "unknown_count": 1,
  "timestamp": "2025-01-15T10:30:00Z",
  "devices": [
    {
      "device_serial": "ZRS004C2",
      "status": "healthy",
      "last_checked": "0001-01-01T00:00:00Z"
    }
  ]
}
```

### `GET /healthz`

Simple health check endpoint for the API server itself. Returns `200 OK` with body `ok`.

## TimescaleDB Schema

### `health_metrics` Hypertable

The core table stores all health observations as time-series data:

```sql
CREATE TABLE health_metrics (
    device_serial TEXT         NOT NULL,
    time          TIMESTAMPTZ  NOT NULL,
    metric_name   TEXT         NOT NULL,
    metric_value  DOUBLE PRECISION NOT NULL,
    unit          TEXT         DEFAULT '',
    source        TEXT         NOT NULL DEFAULT 'unknown'
);
```

Converted to a hypertable with 7-day chunk intervals for optimal time-range query performance.

### Indexes

| Index | Columns | Purpose |
|-------|---------|---------|
| `idx_health_metrics_device_time` | `(device_serial, time DESC)` | Device health history lookups |
| `idx_health_metrics_metric_name_time` | `(metric_name, time DESC)` | Fleet-wide metric queries |
| `idx_health_metrics_device_metric_time` | `(device_serial, metric_name, time DESC)` | Filtered metric history per device |
| `idx_health_metrics_source_time` | `(source, time DESC)` | Source-based metric filtering |

### Time-Series Query Patterns

**Latest metrics for a device:**
```sql
SELECT DISTINCT ON (metric_name)
    device_serial, time, metric_name, metric_value, unit, source
FROM health_metrics
WHERE device_serial = 'ZRS004C2'
ORDER BY metric_name, time DESC;
```

**Temperature trend over the last 7 days:**
```sql
SELECT time, metric_value
FROM health_metrics
WHERE device_serial = 'ZRS004C2'
  AND metric_name = 'farm_current_temperature'
  AND time >= NOW() - INTERVAL '7 days'
ORDER BY time ASC;
```

**Fleet-wide health status rollup:**
```sql
SELECT device_serial, metric_value
FROM (
    SELECT DISTINCT ON (device_serial)
        device_serial, metric_value, time
    FROM health_metrics
    WHERE metric_name = 'smart_check_status'
    ORDER BY device_serial, time DESC
) latest;
```

## Configuration

The API server reads configuration from environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DASHBOARD_LISTEN_ADDR` | `:8080` | HTTP listen address |
| `DASHBOARD_DATABASE_URL` | `postgres://dashboard:dashboard@localhost:5432/openseachest_health?sslmode=disable` | TimescaleDB connection string |

## Getting Started

### Prerequisites

- Go 1.21+
- TimescaleDB (PostgreSQL with TimescaleDB extension)
- openSeaChest CLI binaries installed and in `$PATH`

### Setup

1. **Run the database migration:**
   ```bash
   psql -d openseachest_health -f dashboard/migrations/001_health_metrics.sql
   ```

2. **Build and run the API server:**
   ```bash
   cd dashboard
   go build -o dashboard-server .
   ./dashboard-server
   ```

3. **Verify the server is running:**
   ```bash
   curl http://localhost:8080/healthz
   # ok
   ```

### Running the Agent

The agent can be used programmatically from Go code:

```go
import "github.com/COG-GTM/openSeaChest/dashboard/agent"

target := agent.DeviceTarget{
    Handle: "/dev/sg0",
    Serial: "ZRS004C2",
}

result := agent.CollectAll(target)
// result.Metrics contains all collected HealthMetric entries
// result.SMARTCheck contains the SMART check result with health status
```
