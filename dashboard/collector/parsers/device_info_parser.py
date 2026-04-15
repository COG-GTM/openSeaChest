"""Parse output of ``openSeaChest_Basics --deviceInfo``."""

from __future__ import annotations

import re
from typing import Any


def _not_reported(value: str | None) -> Any:
    """Return None if the value is 'Not Reported' or empty."""
    if value is None:
        return None
    stripped = value.strip()
    if stripped in ("Not Reported", ""):
        return None
    return stripped


def _parse_float(value: str | None) -> float | None:
    """Parse a float value, returning None for non-numeric strings."""
    cleaned = _not_reported(value)
    if cleaned is None:
        return None
    try:
        return float(cleaned)
    except (ValueError, TypeError):
        return None


def _parse_int(value: str | None) -> int | None:
    """Parse an integer value, returning None for non-numeric strings."""
    cleaned = _not_reported(value)
    if cleaned is None:
        return None
    try:
        return int(cleaned)
    except (ValueError, TypeError):
        return None


def _parse_capacity_line(key: str, value: str) -> tuple[int | None, str | None]:
    """Parse a Drive Capacity line, returning (bytes, display_string).

    Key examples:
        'Drive Capacity (TB/TiB)'
        'Drive Capacity (GB/GiB)'
    """
    display = _not_reported(value)
    if display is None:
        return None, None

    # Determine the unit from the key
    unit_match = re.search(r"\((\w+)/", key)
    unit = unit_match.group(1) if unit_match else "GB"

    # Extract the first numeric value
    parts = value.split("/")
    try:
        numeric_value = float(parts[0].strip())
    except (ValueError, TypeError):
        return None, display

    if unit.upper() == "TB":
        capacity_bytes = int(numeric_value * 1_000_000_000_000)
        display_str = f"{numeric_value:.2f} TB"
    else:
        capacity_bytes = int(numeric_value * 1_000_000_000)
        display_str = f"{numeric_value:.2f} GB"

    return capacity_bytes, display_str


def _detect_protocol(specs: list[str], features: list[str], ata_security: str | None) -> str:
    """Detect whether the device uses SATA, SAS, or NVMe protocol.

    Heuristics:
    - If any spec contains "SATA" or "ACS" or "ATA", it's SATA
    - If any spec contains "SPC" or "SBC", it's SAS
    - If any spec contains "NVMe", it's NVMe
    - If features include "TRIM" or "NCQ" with ATA security supported, it's SATA
    - If features include "EPC" or "Format Unit", it's likely SAS
    """
    all_specs_str = " ".join(specs).upper()
    all_features_str = " ".join(features).upper()

    if "NVME" in all_specs_str or "NVME" in all_features_str:
        return "NVMe"

    if any(kw in all_specs_str for kw in ("SATA", "ACS-", "ATA8", "ATA/")):
        return "SATA"

    if any(kw in all_specs_str for kw in ("SPC", "SBC")):
        return "SAS"

    # Fallback heuristics
    if ata_security and ata_security != "Not Supported":
        return "SATA"

    return "SATA"


def _parse_power_on_hours(text: str) -> tuple[float | None, str | None]:
    """Parse Power On Hours and Power On Time lines.

    Examples:
        'Power On Hours: 97.00'
        'Power On Time:  4 days 1 hour'
    """
    hours_val = _parse_float(text)
    return hours_val, None


def _parse_interface_speed_block(lines: list[str], start_idx: int) -> tuple[dict, int]:
    """Parse the Interface speed block, handling both SATA and SAS formats.

    SATA format:
        Interface speed:
            Max Speed (Gb/s): 6.0
            Negotiated Speed (Gb/s): 6.0

    SAS format:
        Interface speed:
            Port 0 (Current Port)
                Max Speed (GB/s): 6.0
                Negotiated Speed (Gb/s): 3.0
            Port 1
                Max Speed (GB/s): 6.0
                Negotiated Speed (Gb/s): Not Reported

    Returns:
        A tuple of (result_dict, end_line_index) where end_line_index is the
        last line consumed by this block (exclusive).
    """
    result: dict = {
        "max_speed_gbps": None,
        "negotiated_speed_gbps": None,
    }

    # Only consume lines that belong to the interface speed block:
    # - "Port N" lines (SAS multi-port)
    # - "Max Speed" lines
    # - "Negotiated Speed" lines
    # - blank lines between them
    # Stop at any other content line.
    i = start_idx + 1
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()

        if not stripped:
            i += 1
            continue

        # Only these patterns belong to the interface speed block
        is_speed_line = (
            "Max Speed" in stripped
            or "Negotiated Speed" in stripped
            or stripped.startswith("Port ")
        )

        if not is_speed_line:
            break

        if "Max Speed" in stripped:
            val = stripped.split(":", 1)[-1].strip()
            parsed = _parse_float(val)
            if parsed is not None and result["max_speed_gbps"] is None:
                result["max_speed_gbps"] = parsed

        if "Negotiated Speed" in stripped:
            val = stripped.split(":", 1)[-1].strip()
            parsed = _parse_float(val)
            if parsed is not None and result["negotiated_speed_gbps"] is None:
                result["negotiated_speed_gbps"] = parsed

        i += 1

    return result, i


def parse_device_info(text: str, device_path: str = "") -> dict:
    """Parse the ``--deviceInfo`` output for a single device.

    Args:
        text: The full stdout text from ``openSeaChest_Basics --deviceInfo``.
        device_path: The OS device path, e.g. '/dev/sg1'.

    Returns:
        A dict conforming to the device portion of the ingestion schema.
    """
    lines = text.splitlines()

    # Key-value store for simple fields
    kv: dict[str, str] = {}
    # Multi-value lists
    specs: list[str] = []
    features: list[str] = []
    # Temperature sub-fields
    temp: dict[str, float | None] = {
        "current_celsius": None,
        "highest_celsius": None,
        "lowest_celsius": None,
    }
    # DST info
    dst: dict[str, Any] = {
        "supported": True,
        "hours_since_last": None,
        "status_result": None,
        "test_run": None,
    }
    # Interface speed
    interface_speed: dict = {
        "max_speed_gbps": None,
        "negotiated_speed_gbps": None,
    }

    in_specs = False
    in_features = False
    in_temperature = False
    in_dst = False
    interface_speed_end = -1  # line index where interface speed block ends

    for i, line in enumerate(lines):
        stripped = line.strip()

        if not stripped or stripped.startswith("="):
            in_specs = False
            in_features = False
            continue

        # Detect section headers
        if stripped == "Specifications Supported:":
            in_specs = True
            in_features = False
            in_temperature = False
            in_dst = False
            in_interface_speed = False
            continue

        if stripped == "Features Supported:":
            in_features = True
            in_specs = False
            in_temperature = False
            in_dst = False
            in_interface_speed = False
            continue

        if stripped == "Temperature Data:":
            in_temperature = True
            in_specs = False
            in_features = False
            in_dst = False
            in_interface_speed = False
            continue

        if stripped == "Last DST information:":
            in_dst = True
            in_temperature = False
            in_specs = False
            in_features = False
            continue

        if stripped == "Interface speed:" or stripped.startswith("Interface speed:"):
            in_temperature = False
            in_specs = False
            in_features = False
            in_dst = False
            interface_speed, interface_speed_end = _parse_interface_speed_block(lines, i)
            continue

        # Collect section items
        if in_specs:
            if ":" not in stripped:
                specs.append(stripped)
                continue
            else:
                in_specs = False

        if in_features:
            if ":" not in stripped:
                features.append(stripped)
                continue
            else:
                in_features = False

        if in_temperature:
            if "Current Temperature" in stripped:
                val = stripped.split(":", 1)[-1].strip()
                temp["current_celsius"] = _parse_float(val)
                continue
            elif "Highest Temperature" in stripped:
                val = stripped.split(":", 1)[-1].strip()
                temp["highest_celsius"] = _parse_float(val)
                continue
            elif "Lowest Temperature" in stripped:
                val = stripped.split(":", 1)[-1].strip()
                temp["lowest_celsius"] = _parse_float(val)
                continue
            elif ":" in stripped and "Humidity" not in stripped:
                in_temperature = False

        if in_dst:
            if "Not supported" in stripped:
                dst["supported"] = False
                in_dst = False
                continue
            elif "Time since last DST" in stripped:
                val = stripped.split(":", 1)[-1].strip()
                dst["hours_since_last"] = _parse_float(val)
                continue
            elif "DST Status/Result" in stripped:
                val = stripped.split(":", 1)[-1].strip()
                dst["status_result"] = _not_reported(val)
                continue
            elif "DST Test run" in stripped:
                val = stripped.split(":", 1)[-1].strip()
                dst["test_run"] = _not_reported(val)
                continue

        if i < interface_speed_end:
            # Already parsed in the block parser
            continue

        # Parse key: value lines
        if ":" in stripped:
            key, _, value = stripped.partition(":")
            key = key.strip()
            value = value.strip()

            # Handle capacity specially
            if key.startswith("Drive Capacity"):
                kv["_capacity_key"] = key
                kv["Drive Capacity"] = value
            else:
                kv[key] = value

    # Build the structured output
    rotation_rate_str = _not_reported(kv.get("Rotation Rate (RPM)"))

    is_ssd = (
        rotation_rate_str is not None and rotation_rate_str.upper() == "SSD"
    ) or rotation_rate_str is None

    rotation_rate_rpm: int | None = None
    if rotation_rate_str and rotation_rate_str.upper() != "SSD":
        try:
            rotation_rate_rpm = int(rotation_rate_str)
        except ValueError:
            rotation_rate_rpm = None

    # Detect protocol
    ata_security = _not_reported(kv.get("ATA Security Information"))
    protocol = _detect_protocol(specs, features, ata_security)

    # Build capacity info
    capacity_key = kv.get("_capacity_key", "Drive Capacity (GB/GiB)")
    capacity_value = kv.get("Drive Capacity")
    capacity_bytes, capacity_display = _parse_capacity_line(
        capacity_key, capacity_value
    ) if capacity_value else (None, None)

    # Power on
    power_on_hours = _parse_float(kv.get("Power On Hours"))
    power_on_display = _not_reported(kv.get("Power On Time"))

    # Workload
    workload_rate = _parse_float(kv.get("Annualized Workload Rate (TB/yr)"))

    total_read_str = kv.get("Total Bytes Read (GB)") or kv.get("Total Bytes Read (TB)")
    total_written_str = kv.get("Total Bytes Written (GB)") or kv.get("Total Bytes Written (TB)")

    total_read_bytes: int | None = None
    if total_read_str and _not_reported(total_read_str):
        val = _parse_float(total_read_str)
        if val is not None:
            if "Total Bytes Read (TB)" in kv:
                total_read_bytes = int(val * 1_000_000_000_000)
            else:
                total_read_bytes = int(val * 1_000_000_000)

    total_written_bytes: int | None = None
    if total_written_str and _not_reported(total_written_str):
        val = _parse_float(total_written_str)
        if val is not None:
            if "Total Bytes Written (TB)" in kv:
                total_written_bytes = int(val * 1_000_000_000_000)
            else:
                total_written_bytes = int(val * 1_000_000_000)

    endurance = _parse_float(
        kv.get("Percentage Used Endurance Indicator (%)")
    )

    # Firmware
    fw_revision = _not_reported(kv.get("Firmware Revision")) or ""
    fw_download_str = _not_reported(kv.get("Firmware Download Support"))
    fw_download: list[str] = []
    if fw_download_str:
        fw_download = [s.strip() for s in fw_download_str.split(",")]

    # Security
    encryption = _not_reported(kv.get("Encryption Support")) or "Not Supported"

    # Form factor
    form_factor = _parse_float(kv.get("Form Factor (inch)"))

    # Device type from schema
    from dashboard.collector.schema import derive_device_type
    raw_rotation = _not_reported(kv.get("Rotation Rate (RPM)"))
    device_type = derive_device_type(protocol, raw_rotation)

    device: dict[str, Any] = {
        "device_path": device_path,
        "device_type": device_type,
        "identity": {
            "model_number": _not_reported(kv.get("Model Number")) or "",
            "serial_number": _not_reported(kv.get("Serial Number")) or "",
            "firmware_revision": fw_revision,
            "world_wide_name": _not_reported(kv.get("World Wide Name")),
            "form_factor_inches": form_factor,
            "rotation_rate_rpm": rotation_rate_rpm,
            "is_ssd": is_ssd,
            "logical_sector_size_bytes": _parse_int(
                kv.get("Logical Sector Size (B)")
            ),
            "physical_sector_size_bytes": _parse_int(
                kv.get("Physical Sector Size (B)")
            ),
        },
        "capacity": {
            "capacity_bytes": capacity_bytes,
            "capacity_display": capacity_display,
            "max_lba": _parse_int(kv.get("MaxLBA")),
            "native_max_lba": _parse_int(kv.get("Native MaxLBA")),
        },
        "temperature": temp,
        "power_on": {
            "power_on_hours": power_on_hours,
            "power_on_hours_display": power_on_display,
        },
        "interface": {
            "protocol": protocol,
            "max_speed_gbps": interface_speed["max_speed_gbps"],
            "negotiated_speed_gbps": interface_speed["negotiated_speed_gbps"],
            "specifications": specs,
            "features": features,
        },
        "workload": {
            "annualized_workload_rate_tb_yr": workload_rate,
            "total_bytes_read": total_read_bytes,
            "total_bytes_written": total_written_bytes,
            "percentage_used_endurance": endurance,
        },
        "smart": {
            "status": _not_reported(kv.get("SMART Status")) or "Unknown",
            "tripped": kv.get("SMART Status", "").strip() == "Bad",
            "attributes": [],
            "last_dst": dst,
        },
        "firmware": {
            "revision": fw_revision,
            "download_support": fw_download,
        },
        "security": {
            "encryption_support": encryption,
            "ata_security": ata_security,
        },
    }

    return device
