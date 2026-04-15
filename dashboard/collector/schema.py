"""Schema version constant and validation helpers."""

SCHEMA_VERSION = "1.0.0"

VALID_DEVICE_TYPES = frozenset({
    "SATA_HDD",
    "SATA_SSD",
    "SAS_HDD",
    "SAS_SSD",
    "NVMe",
})

VALID_PROTOCOLS = frozenset({"SATA", "SAS", "NVMe"})

VALID_SMART_STATUSES = frozenset({"Good", "Bad", "Unknown"})


def derive_device_type(protocol: str, rotation_rate_str: str | None) -> str:
    """Derive the device_type enum value from protocol and rotation rate.

    Args:
        protocol: One of "SATA", "SAS", "NVMe".
        rotation_rate_str: The raw rotation rate string from CLI output,
            e.g. "7200", "SSD", or None.

    Returns:
        One of the VALID_DEVICE_TYPES values.

    Raises:
        ValueError: If the protocol is not recognized.
    """
    if protocol == "NVMe":
        return "NVMe"

    is_ssd = (
        rotation_rate_str is None
        or rotation_rate_str.strip().upper() == "SSD"
    )

    if protocol == "SATA":
        return "SATA_SSD" if is_ssd else "SATA_HDD"
    elif protocol == "SAS":
        return "SAS_SSD" if is_ssd else "SAS_HDD"
    else:
        raise ValueError(f"Unknown protocol: {protocol}")


def validate_payload(payload: dict) -> list[str]:
    """Validate a collector payload against basic schema rules.

    Returns a list of error messages. Empty list means valid.
    """
    errors: list[str] = []

    if payload.get("schema_version") != SCHEMA_VERSION:
        errors.append(
            f"schema_version must be '{SCHEMA_VERSION}', "
            f"got '{payload.get('schema_version')}'"
        )

    if "host" not in payload:
        errors.append("Missing required field: host")
    elif "hostname" not in payload.get("host", {}):
        errors.append("Missing required field: host.hostname")

    if "collected_at" not in payload:
        errors.append("Missing required field: collected_at")

    if "devices" not in payload:
        errors.append("Missing required field: devices")
    elif not isinstance(payload["devices"], list):
        errors.append("devices must be an array")
    else:
        for i, device in enumerate(payload["devices"]):
            prefix = f"devices[{i}]"
            if "device_path" not in device:
                errors.append(f"{prefix}: missing device_path")
            if "device_type" not in device:
                errors.append(f"{prefix}: missing device_type")
            elif device["device_type"] not in VALID_DEVICE_TYPES:
                errors.append(
                    f"{prefix}: invalid device_type '{device['device_type']}'"
                )
            if "identity" not in device:
                errors.append(f"{prefix}: missing identity")
            else:
                identity = device["identity"]
                for field in (
                    "model_number",
                    "serial_number",
                    "firmware_revision",
                ):
                    if field not in identity:
                        errors.append(
                            f"{prefix}.identity: missing {field}"
                        )

    return errors
