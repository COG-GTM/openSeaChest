# OpenSeaChest Fleet Dashboard - Data Collection Agent

The Data Collection Agent shells out to openSeaChest CLI tools, parses their
unstructured stdout text, and produces structured JSON conforming to the
[Dashboard Ingestion Schema v1.0.0](../../docs/dashboard/ingestion-schema.json).

## Package Structure

```
dashboard/collector/
  __init__.py          # Package init, __version__
  cli_runner.py        # Subprocess wrapper for openSeaChest CLIs
  collector.py         # Main orchestrator: scan -> collect -> build payload
  schema.py            # Schema version constant, validation helper
  parsers/
    __init__.py
    scan_parser.py         # Parse --scan output to list device handles
    device_info_parser.py  # Parse --deviceInfo output
    smart_parser.py        # Parse --smartCheck and --smartAttributes output
  tests/
    __init__.py
    test_parsers.py    # Unit tests for all parsers
    test_collector.py  # Unit tests for the collector orchestrator
    fixtures/          # Captured stdout text files for testing
```

## Supported Device Types

| Type       | Protocol | Description                     |
|------------|----------|---------------------------------|
| SATA_HDD   | SATA     | SATA hard disk drive            |
| SATA_SSD   | SATA     | SATA solid-state drive          |
| SAS_HDD    | SAS      | SAS hard disk drive             |
| SAS_SSD    | SAS      | SAS solid-state drive           |
| NVMe       | NVMe     | NVMe solid-state drive          |

Device type is automatically derived from the interface protocol and the
reported rotation rate. If the rotation rate is `"SSD"` or not reported
(NVMe), the device is classified as an SSD.

## Usage

### As a Library

```python
from dashboard.collector.collector import Collector

# Use defaults (expects openSeaChest_Basics and openSeaChest_SMART on PATH)
collector = Collector()
payload = collector.collect()

# payload is a dict matching the ingestion schema
print(payload["schema_version"])  # "1.0.0"
for device in payload["devices"]:
    print(f"{device['device_path']}: {device['device_type']}")
```

### With Custom CLI Paths

```python
from dashboard.collector.cli_runner import CLIRunner
from dashboard.collector.collector import Collector

cli = CLIRunner(
    basics_path="/opt/openSeaChest/openSeaChest_Basics",
    smart_path="/opt/openSeaChest/openSeaChest_SMART",
    timeout=300,
)
collector = Collector(cli=cli)
payload = collector.collect()
```

### Using Individual Parsers

```python
from dashboard.collector.parsers import (
    parse_scan_output,
    parse_device_info,
    parse_smart_attributes,
)

# Parse scan output
with open("scan_output.txt") as f:
    devices = parse_scan_output(f.read())

# Parse device info
with open("device_info.txt") as f:
    device = parse_device_info(f.read(), "/dev/sg0")

# Parse SMART attributes
with open("smart_attrs.txt") as f:
    attrs = parse_smart_attributes(f.read())
```

### Validating a Payload

```python
from dashboard.collector.schema import validate_payload

errors = validate_payload(payload)
if errors:
    print("Validation errors:", errors)
else:
    print("Payload is valid")
```

## Running Tests

From the repository root:

```bash
python -m pytest dashboard/collector/tests/ -v
```

Or with unittest:

```bash
python -m unittest discover -s dashboard/collector/tests -v
```

## Output Schema

The collector produces JSON matching the shared ingestion schema
(`docs/dashboard/ingestion-schema.json`). Key fields:

- `schema_version`: Always `"1.0.0"`
- `host`: Hostname, OS, and agent version
- `collected_at`: ISO-8601 UTC timestamp
- `devices[]`: Array of device objects with identity, capacity, temperature,
  power-on hours, interface details, workload stats, SMART data, firmware
  info, and security status

## Parsing Rules

- Fields are extracted from tab/space-indented `Key: Value` lines
- `"Not Reported"` values are mapped to `null`
- SMART attribute hex values (nominal, worst) are converted to decimal
- SAS interface speed uses Port 0 (Current Port) values
- `Rotation Rate (RPM): SSD` sets `is_ssd=true`, `rotation_rate_rpm=null`
- Capacity bytes are computed from the decimal (TB/GB) value reported
