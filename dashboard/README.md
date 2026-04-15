# openSeaChest Fleet Health Dashboard

A fleet-wide device inventory and health monitoring system built on top of the [openSeaChest](https://github.com/Seagate/openSeaChest) storage diagnostic toolkit.

## High-Level Architecture

```
┌──────────────┐        ┌──────────────────┐        ┌────────────┐
│  osc-agent   │──POST──▶  Fleet Inventory │──SQL──▶│ PostgreSQL │
│  (per host)  │        │  REST API        │        │            │
└──────────────┘        └──────┬───────────┘        └────────────┘
                               │
                          GET /devices
                          GET /hosts
                               │
                        ┌──────▼───────┐
                        │  Dashboard   │
                        │  (future UI) │
                        └──────────────┘
```

The system has three main components:

1. **osc-agent** — a lightweight Go binary that runs on each host  
2. **Fleet Inventory API** — a central Go REST service  
3. **PostgreSQL** — persistent storage for hosts, devices, and snapshots

---

## Agent (`dashboard/agent/`)

The agent wraps the openSeaChest CLI tools to discover and report local storage devices.

### How It Works

1. Runs `openSeaChest_Basics --scan` to enumerate device handles (e.g. `/dev/sg0`).
2. For each handle, runs `openSeaChest_Basics -d <handle> -i` to collect device details.
3. Parses stdout to extract: model (`product_identification`), serial number, firmware revision (`product_revision`), capacity, interface type, and World Wide Name (WWN).
4. POSTs a JSON inventory report to the central API at `POST /api/v1/agents/{agent_id}/report`.

### Exit Code Handling

The agent maps all openSeaChest exit codes defined in `include/openseachest_util_options.h` (`eUtilExitCodes`). Notably:

| Code | Name | Agent Behaviour |
|------|------|-----------------|
| 0 | `UTIL_EXIT_NO_ERROR` | Success |
| 9 | `UTIL_EXIT_NEED_ELEVATED_PRIVILEGES` | Logs a clear error advising the user to run as root/administrator |
| 12 | `UTIL_EXIT_NO_DEVICE` | Treated as zero devices discovered |
| Others | Various | Logged with human-readable descriptions |

### Usage

```bash
# Run once and report to the API
./osc-agent --api-url http://fleet-api:8080

# Run continuously every 5 minutes
./osc-agent --api-url http://fleet-api:8080 --interval 5m

# Specify a custom path to the openSeaChest binary
./osc-agent --basics-path /opt/openSeaChest/bin/openSeaChest_Basics
```

### Building

```bash
cd dashboard/agent
go build -o osc-agent .
```

---

## API (`dashboard/api/`)

A Go REST API (stdlib `net/http`, no framework) that stores device inventory in PostgreSQL.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/agents/{agent_id}/report` | Receive device inventory from an agent |
| `GET` | `/api/v1/devices` | List devices (paginated, filterable) |
| `GET` | `/api/v1/devices/{serial}` | Get a single device by serial number |
| `GET` | `/api/v1/hosts` | List all known hosts |
| `GET` | `/healthz` | Health check |

#### Filters on `GET /api/v1/devices`

| Query Param | Description |
|-------------|-------------|
| `host` | Filter by hostname (substring, case-insensitive) |
| `model` | Filter by model (substring, case-insensitive) |
| `interface` | Filter by interface type (substring, case-insensitive) |
| `firmware_rev` | Filter by firmware revision (substring, case-insensitive) |
| `page` | Page number (default 1) |
| `per_page` | Results per page (default 20, max 100) |

#### Example Request/Response

```bash
# Agent reports devices
curl -X POST http://localhost:8080/api/v1/agents/host-01/report \
  -H 'Content-Type: application/json' \
  -d '{
    "agent_id": "host-01",
    "hostname": "storage-node-1",
    "devices": [
      {
        "device_handle": "/dev/sg0",
        "model": "ST8000NM000A",
        "serial_number": "ZR10ABCD",
        "firmware_rev": "SN06",
        "capacity_bytes": 8001563222016,
        "interface_type": "SATA",
        "wwn": "5000c500abcdef01"
      }
    ]
  }'

# List devices filtered by model
curl 'http://localhost:8080/api/v1/devices?model=ST8000&page=1&per_page=10'
```

### Building

```bash
cd dashboard/api
go build -o fleet-api .
```

### Running

```bash
# Set the PostgreSQL connection string
export DATABASE_URL="postgres://fleet:fleet@localhost:5432/fleet_inventory?sslmode=disable"
./fleet-api --addr :8080
```

---

## Database Schema (`dashboard/migrations/`)

Apply migrations with any PostgreSQL client:

```bash
psql "$DATABASE_URL" -f dashboard/migrations/001_inventory.sql
```

### Tables

#### `hosts`
| Column | Type | Description |
|--------|------|-------------|
| `id` | `BIGSERIAL` | Primary key |
| `agent_id` | `TEXT UNIQUE` | Unique agent identifier |
| `hostname` | `TEXT` | Host's reported hostname |
| `first_seen` | `TIMESTAMPTZ` | First registration time |
| `last_seen` | `TIMESTAMPTZ` | Most recent report time |

#### `devices`
| Column | Type | Description |
|--------|------|-------------|
| `id` | `BIGSERIAL` | Primary key |
| `host_id` | `BIGINT FK` | References `hosts(id)` |
| `serial_number` | `TEXT UNIQUE` | Drive serial number |
| `model` | `TEXT` | Product identification string |
| `firmware_rev` | `TEXT` | Firmware/product revision |
| `capacity_bytes` | `BIGINT` | Raw capacity in bytes |
| `interface_type` | `TEXT` | SATA, SAS, NVMe, USB |
| `wwn` | `TEXT` | World Wide Name |
| `first_seen` | `TIMESTAMPTZ` | First discovery time |
| `last_seen` | `TIMESTAMPTZ` | Most recent report time |

#### `device_snapshots`
| Column | Type | Description |
|--------|------|-------------|
| `id` | `BIGSERIAL` | Primary key |
| `device_id` | `BIGINT FK` | References `devices(id)` |
| `snapshot_time` | `TIMESTAMPTZ` | When the snapshot was taken |
| `raw_data` | `JSONB` | Full raw device info as JSON |

---

## Project Structure

```
dashboard/
├── agent/                    # Per-host agent binary
│   ├── main.go               # CLI entrypoint
│   ├── go.mod                # Go module definition
│   └── internal/
│       ├── models/
│       │   └── device.go     # Device & InventoryReport structs
│       ├── scanner/
│       │   ├── exitcodes.go  # openSeaChest exit code constants
│       │   └── scanner.go    # CLI wrapper & stdout parser
│       └── reporter/
│           └── reporter.go   # HTTP client for POSTing reports
├── api/                      # Central REST API
│   ├── main.go               # Server entrypoint & router
│   ├── go.mod                # Go module definition
│   └── internal/
│       ├── models/
│       │   └── models.go     # Request/response types
│       ├── handlers/
│       │   └── handlers.go   # HTTP handler implementations
│       ├── store/
│       │   └── store.go      # PostgreSQL data access layer
│       └── middleware/
│           └── middleware.go  # Logging & JSON middleware
├── migrations/
│   └── 001_inventory.sql     # Initial database schema
└── README.md                 # This file
```
