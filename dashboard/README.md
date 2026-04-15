# openSeaChest Fleet Health Dashboard — Firmware Compliance Auditing

This module adds firmware policy management, compliance evaluation, and firmware
update campaign tracking to the openSeaChest ecosystem. It is implemented as a
standalone Go HTTP service backed by SQLite.

## Architecture

```
┌──────────────┐        ┌────────────────────┐        ┌──────────────┐
│   REST API   │───────▶│ Compliance Engine   │───────▶│   SQLite DB  │
│  (chi router)│        │ (glob + exact match)│        │              │
└──────────────┘        └────────────────────┘        └──────────────┘
       │                         │
       │                         ▼
       │                ┌────────────────────┐
       │                │  Campaign Executor  │
       │                │  (exit code mapper) │
       │                └────────────────────┘
       │                         │
       ▼                         ▼
┌──────────────┐        ┌────────────────────┐
│   Agents     │◀───────│ openSeaChest_       │
│   (fleet)    │        │ Firmware CLI        │
└──────────────┘        └────────────────────┘
```

## Firmware Policy Management

A **firmware policy** declares that all devices whose model string matches a
glob pattern must be running a specific firmware revision.

| Field              | Description                                             |
|--------------------|---------------------------------------------------------|
| `name`             | Human-readable policy name                              |
| `model_pattern`    | Glob pattern matched against `product_identification`   |
| `required_firmware`| Exact firmware revision devices must be running         |
| `severity`         | `critical`, `warning`, or `info`                        |

### Glob Matching (model_pattern)

The compliance engine uses Go's `filepath.Match` for glob semantics, mirroring
the `MODEL_MATCH_FLAG` logic in `utils/C/openSeaChest/openSeaChest_NVMe.c`:

- `*` matches any sequence of characters
- `?` matches any single character
- `[abc]` matches character classes

Both the pattern and the device model are compared **case-insensitively**.

Example: pattern `ST500LM*` matches model `ST500LM001`.

### Firmware Comparison

The compliance engine performs an **exact string match** between
`required_firmware` and the device's `product_revision`, mirroring the
`FW_MATCH_FLAG` (`--onlyFW`) behaviour in `openSeaChest_NVMe.c`.

## Compliance Evaluation Logic

When the fleet compliance report is requested:

1. All firmware policies are loaded from the database.
2. All known devices (registered via agents) are loaded.
3. For each (policy, device) pair:
   - If `device.product_identification` does **not** match `policy.model_pattern` → skip (policy does not apply).
   - If it matches and `device.product_revision == policy.required_firmware` → **compliant**.
   - If it matches and firmware differs → **non-compliant**.
4. Results are aggregated into per-policy counts of compliant / non-compliant / unknown.
5. Individual compliance results are persisted for audit trail.

## Campaign Execution Flow

Firmware update campaigns automate the rollout of firmware updates to
non-compliant devices.

### Lifecycle

1. **Create campaign** — POST a campaign referencing a policy and a firmware
   file. All devices matching the policy that are non-compliant are enqueued
   with status `pending`.
2. **Agents execute** — Each agent runs the command:
   ```
   openSeaChest_Firmware --downloadFW <firmware_file> -d <device_handle>
   ```
3. **Report results** — Agents POST the exit code back to the API.
4. **Track progress** — GET the campaign status to see per-device results and
   aggregate counts.

### Exit Code Mapping

Exit codes are mapped from `openSeaChest_Firmware` return codes as documented
in `docs/man/man8/openSeaChest_Firmware.8` and the `eUtilExitCodes` enum in
`include/openseachest_util_options.h`:

| Exit Code | Meaning (from man page)              | Campaign Device Status |
|-----------|--------------------------------------|------------------------|
| 0         | No Error Found                       | `success`              |
| 32        | Firmware Download Complete           | `success`              |
| 33        | Deferred FW Download Complete        | `deferred` (reboot required) |
| 38        | Firmware Already up to date          | `skipped`              |
| 36        | Model matched, FW mismatched         | `wrong_fw`             |
| 3         | Operation Failure                    | `failed`               |
| Other     | Unknown error                        | `failed`               |

## API Endpoints

### Firmware Policies

| Method | Path                              | Description            |
|--------|-----------------------------------|------------------------|
| POST   | `/api/v1/firmware/policies`       | Create a policy        |
| GET    | `/api/v1/firmware/policies`       | List all policies      |
| PUT    | `/api/v1/firmware/policies/{id}`  | Update a policy        |

#### Create Policy — `POST /api/v1/firmware/policies`

```json
{
  "name": "ST500 FW Update",
  "model_pattern": "ST500LM*",
  "required_firmware": "LVM1",
  "severity": "critical"
}
```

#### Update Policy — `PUT /api/v1/firmware/policies/{id}`

```json
{
  "required_firmware": "LVM2",
  "severity": "warning"
}
```

### Compliance Report

| Method | Path                           | Description                  |
|--------|--------------------------------|------------------------------|
| GET    | `/api/v1/firmware/compliance`  | Fleet-wide compliance report |

Response:

```json
{
  "generated_at": "2025-01-15T10:30:00Z",
  "policies": [
    {
      "policy_id": 1,
      "policy_name": "ST500 FW Update",
      "severity": "critical",
      "compliant": 42,
      "non_compliant": 8,
      "unknown": 0,
      "total": 50
    }
  ]
}
```

### Firmware Campaigns

| Method | Path                                          | Description                    |
|--------|-----------------------------------------------|--------------------------------|
| POST   | `/api/v1/firmware/campaigns`                  | Create a campaign              |
| GET    | `/api/v1/firmware/campaigns/{id}`             | Campaign status with devices   |
| POST   | `/api/v1/firmware/campaigns/{id}/results`     | Report device exit code        |

#### Create Campaign — `POST /api/v1/firmware/campaigns`

```json
{
  "name": "Q1 ST500 Firmware Rollout",
  "policy_id": 1,
  "firmware_file": "/firmware/ST500_LVM2.bin"
}
```

#### Report Device Result — `POST /api/v1/firmware/campaigns/{id}/results`

```json
{
  "device_serial": "WBY1K234",
  "exit_code": 0
}
```

## Database Schema

See `migrations/001_firmware_compliance.sql` for the full schema. Tables:

- **firmware_policies** — policy definitions with glob patterns and severity
- **compliance_results** — per-device compliance evaluation audit log
- **firmware_campaigns** — campaign metadata and lifecycle status
- **campaign_devices** — per-device status within a campaign

## Running

```bash
cd dashboard
go build -o dashboard .
./dashboard
```

Environment variables:
- `DASHBOARD_DB_PATH` — SQLite database file path (default: `dashboard.db`)
- `DASHBOARD_ADDR` — Listen address (default: `:8080`)
