"""Parse output of ``openSeaChest_SMART --smartCheck`` and ``--smartAttributes``."""

from __future__ import annotations

import re


def parse_smart_check(text: str) -> dict:
    """Parse the ``--smartCheck`` output and return SMART status info.

    Looks for lines like:
        SMART Status: Good
        SMART Status: Bad

    Returns:
        A dict with ``status`` ("Good", "Bad", or "Unknown") and
        ``tripped`` (bool).
    """
    status = "Unknown"
    tripped = False

    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("SMART Status:"):
            value = stripped.split(":", 1)[-1].strip()
            if value in ("Good", "Bad"):
                status = value
                tripped = value == "Bad"
            break

    return {"status": status, "tripped": tripped}


def parse_smart_attributes(text: str) -> list[dict]:
    """Parse the ``--smartAttributes raw`` output.

    Expected table format::

        Attribute Name:                   Status: Nominal: Worst:     Raw (hex):
          1 Read Error Rate               0x000F    0x75    0x63      0x0000000969AFB4

    Each attribute line starts with an ID number (1-255) followed by name,
    hex status flags, hex nominal, hex worst, and hex raw value.

    Returns:
        A list of dicts matching the ``smart_attribute`` schema definition.
    """
    attributes: list[dict] = []

    # Pattern: <spaces><id> <name> <status_hex> <nominal_hex> <worst_hex> <raw_hex>
    # The name can contain multiple words and special characters
    attr_pattern = re.compile(
        r"^\s*(\d{1,3})\s+"       # ID (1-255)
        r"(.+?)\s+"               # Name (multi-word, greedy but minimal)
        r"(0x[0-9A-Fa-f]+)\s+"   # Status flags (hex)
        r"(0x[0-9A-Fa-f]+)\s+"   # Nominal (hex)
        r"(0x[0-9A-Fa-f]+)\s+"   # Worst (hex)
        r"(0x[0-9A-Fa-f]+)"      # Raw value (hex)
    )

    for line in text.splitlines():
        match = attr_pattern.match(line)
        if not match:
            continue

        attr_id = int(match.group(1))
        name = match.group(2).strip()
        status_flags = match.group(3)
        nominal_hex = match.group(4)
        worst_hex = match.group(5)
        raw_hex = match.group(6)

        # Convert hex values to decimal
        try:
            nominal = int(nominal_hex, 16)
        except ValueError:
            nominal = None

        try:
            worst = int(worst_hex, 16)
        except ValueError:
            worst = None

        try:
            raw_value = int(raw_hex, 16)
        except ValueError:
            raw_value = None

        attributes.append({
            "id": attr_id,
            "name": name,
            "status_flags": status_flags,
            "nominal": nominal,
            "worst": worst,
            "raw_hex": raw_hex,
            "raw_value": raw_value,
        })

    return attributes
