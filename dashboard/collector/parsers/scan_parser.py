"""Parse output of ``openSeaChest_Basics --scan``."""

from __future__ import annotations

import re


def parse_scan_output(text: str) -> list[dict]:
    """Parse the ``--scan`` output and return a list of device handles.

    Expected output lines look like::

        /dev/sg0 - SEAGATE  ST4000DX001-1CE168 - CC44 - 4.00 TB
        /dev/sg1 - SEAGATE  ST120FP0021 - B770 - 120.03 GB

    or on Windows::

        PD0 - SEAGATE  ST4000DX001-1CE168 - CC44 - 4.00 TB

    Returns:
        A list of dicts, each with at least ``device_path`` and optionally
        ``model``, ``firmware``, ``capacity``.
    """
    devices: list[dict] = []

    # Match lines that start with a device path (Linux /dev/* or Windows PD*)
    device_pattern = re.compile(
        r"^\s*(/dev/\S+|PD\d+)\s+-\s+(.*)$"
    )

    for line in text.splitlines():
        match = device_pattern.match(line)
        if not match:
            continue

        device_path = match.group(1).strip()
        remainder = match.group(2).strip()

        device: dict = {"device_path": device_path}

        # Try to parse the remainder: "VENDOR MODEL - FW_REV - CAPACITY"
        parts = [p.strip() for p in remainder.split(" - ")]
        if len(parts) >= 1:
            device["model"] = parts[0]
        if len(parts) >= 2:
            device["firmware"] = parts[1]
        if len(parts) >= 3:
            device["capacity"] = parts[2]

        devices.append(device)

    return devices
