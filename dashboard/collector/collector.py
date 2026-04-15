"""Main orchestrator: scan -> collect info -> build payload."""

from __future__ import annotations

import logging
import platform
import socket
from datetime import datetime, timezone
from typing import Optional

from dashboard.collector import __version__
from dashboard.collector.cli_runner import CLIRunner, CLIError
from dashboard.collector.parsers.scan_parser import parse_scan_output
from dashboard.collector.parsers.device_info_parser import parse_device_info
from dashboard.collector.parsers.smart_parser import (
    parse_smart_check,
    parse_smart_attributes,
)
from dashboard.collector.schema import SCHEMA_VERSION, validate_payload

logger = logging.getLogger(__name__)


class Collector:
    """Orchestrates a full collection run across all discovered devices.

    Args:
        cli: A CLIRunner instance. If None, a default one is created.
    """

    def __init__(self, cli: Optional[CLIRunner] = None):
        self.cli = cli or CLIRunner()

    def collect(self) -> dict:
        """Run a full collection and return the ingestion payload.

        Steps:
            1. Scan for devices
            2. For each device, collect deviceInfo
            3. For each device, collect SMART check and attributes
            4. Build and return the full payload

        Returns:
            A dict conforming to the ingestion schema.
        """
        # Scan for devices
        scan_output = self.cli.scan()
        discovered = parse_scan_output(scan_output)
        logger.info("Discovered %d device(s)", len(discovered))

        devices: list[dict] = []
        for entry in discovered:
            device_path = entry["device_path"]
            try:
                device = self._collect_device(device_path)
                devices.append(device)
            except CLIError:
                logger.warning(
                    "Failed to collect data for %s, skipping",
                    device_path,
                    exc_info=True,
                )

        payload = self._build_payload(devices)

        errors = validate_payload(payload)
        if errors:
            logger.warning(
                "Payload validation warnings: %s", "; ".join(errors)
            )

        return payload

    def _collect_device(self, device_path: str) -> dict:
        """Collect all information for a single device.

        Args:
            device_path: OS device handle, e.g. '/dev/sg1'.

        Returns:
            A device dict matching the ingestion schema.
        """
        # Get device info
        info_output = self.cli.device_info(device_path)
        device = parse_device_info(info_output, device_path)

        # Get SMART status
        try:
            smart_check_output = self.cli.smart_check(device_path)
            smart_status = parse_smart_check(smart_check_output)
            device["smart"]["status"] = smart_status["status"]
            device["smart"]["tripped"] = smart_status["tripped"]
        except CLIError:
            logger.warning(
                "SMART check failed for %s", device_path, exc_info=True
            )

        # Get SMART attributes
        try:
            smart_attrs_output = self.cli.smart_attributes(device_path)
            attrs = parse_smart_attributes(smart_attrs_output)
            device["smart"]["attributes"] = attrs
        except CLIError:
            logger.warning(
                "SMART attributes failed for %s",
                device_path,
                exc_info=True,
            )

        return device

    def _build_payload(self, devices: list[dict]) -> dict:
        """Wrap device data into the full ingestion payload.

        Args:
            devices: List of device dicts.

        Returns:
            The complete payload dict.
        """
        os_info = f"{platform.system()} {platform.release()}"

        return {
            "schema_version": SCHEMA_VERSION,
            "host": {
                "hostname": socket.gethostname(),
                "os": os_info,
                "agent_version": __version__,
            },
            "collected_at": datetime.now(timezone.utc).isoformat(),
            "devices": devices,
        }
