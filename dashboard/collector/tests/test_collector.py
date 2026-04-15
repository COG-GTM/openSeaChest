"""Unit tests for the Collector orchestrator."""

from __future__ import annotations

import os
import unittest
from unittest.mock import MagicMock

from dashboard.collector.collector import Collector
from dashboard.collector.cli_runner import CLIRunner, CLIError
from dashboard.collector.schema import SCHEMA_VERSION, validate_payload

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def _load_fixture(name: str) -> str:
    with open(os.path.join(FIXTURES_DIR, name), "r") as f:
        return f.read()


class TestCollector(unittest.TestCase):
    def setUp(self):
        """Set up a mock CLI runner."""
        self.mock_cli = MagicMock(spec=CLIRunner)
        self.collector = Collector(cli=self.mock_cli)

        # Configure scan to return two devices
        self.mock_cli.scan.return_value = _load_fixture("scan_output.txt")

        # Map device_info calls to the appropriate fixture
        def device_info_side_effect(device_path):
            fixture_map = {
                "/dev/sg0": "sata_hdd_info.txt",
                "/dev/sg1": "sata_ssd_info.txt",
                "/dev/sg2": "sas_hdd_info.txt",
                "/dev/sg3": "sas_ssd_info.txt",
            }
            fixture_name = fixture_map.get(device_path)
            if fixture_name:
                return _load_fixture(fixture_name)
            raise CLIError(["cmd"], 1, "Device not found")

        self.mock_cli.device_info.side_effect = device_info_side_effect

        # SMART check returns good status
        self.mock_cli.smart_check.return_value = (
            "        SMART Status: Good\n"
        )

        # SMART attributes returns fixture
        self.mock_cli.smart_attributes.return_value = _load_fixture(
            "smart_attributes.txt"
        )

    def test_collect_produces_valid_payload(self):
        payload = self.collector.collect()

        # Check top-level fields
        self.assertEqual(payload["schema_version"], SCHEMA_VERSION)
        self.assertIn("host", payload)
        self.assertIn("hostname", payload["host"])
        self.assertIn("os", payload["host"])
        self.assertIn("agent_version", payload["host"])
        self.assertIn("collected_at", payload)
        self.assertIn("devices", payload)

        # Should have 4 devices
        self.assertEqual(len(payload["devices"]), 4)

        # Validate passes
        errors = validate_payload(payload)
        self.assertEqual(errors, [])

    def test_collect_device_types(self):
        payload = self.collector.collect()
        devices = payload["devices"]

        types = [d["device_type"] for d in devices]
        self.assertIn("SATA_HDD", types)
        self.assertIn("SATA_SSD", types)
        self.assertIn("SAS_HDD", types)
        self.assertIn("SAS_SSD", types)

    def test_collect_smart_attributes_populated(self):
        payload = self.collector.collect()

        # SATA HDD should have SMART attributes
        sata_hdd = payload["devices"][0]
        self.assertEqual(len(sata_hdd["smart"]["attributes"]), 21)

    def test_collect_handles_smart_failure(self):
        """If SMART commands fail for a device, the device is still included."""
        self.mock_cli.smart_check.side_effect = CLIError(
            ["cmd"], 1, "SMART not supported"
        )
        self.mock_cli.smart_attributes.side_effect = CLIError(
            ["cmd"], 1, "SMART not supported"
        )

        payload = self.collector.collect()
        self.assertEqual(len(payload["devices"]), 4)

        # SMART status should still be from deviceInfo parsing
        for device in payload["devices"]:
            self.assertIn("smart", device)

    def test_collect_handles_device_info_failure(self):
        """If deviceInfo fails for one device, others are still collected."""
        original_side_effect = self.mock_cli.device_info.side_effect

        def failing_side_effect(device_path):
            if device_path == "/dev/sg1":
                raise CLIError(["cmd"], 1, "Device busy")
            return original_side_effect(device_path)

        self.mock_cli.device_info.side_effect = failing_side_effect

        payload = self.collector.collect()
        # Should have 3 devices (sg1 failed)
        self.assertEqual(len(payload["devices"]), 3)

    def test_collect_schema_version(self):
        payload = self.collector.collect()
        self.assertEqual(payload["schema_version"], "1.0.0")

    def test_collect_host_info(self):
        payload = self.collector.collect()
        host = payload["host"]
        self.assertIsInstance(host["hostname"], str)
        self.assertIsInstance(host["os"], str)
        self.assertEqual(host["agent_version"], "1.0.0")

    def test_collect_collected_at_is_iso8601(self):
        payload = self.collector.collect()
        collected_at = payload["collected_at"]
        # Should be a valid ISO-8601 string with timezone info
        self.assertIn("T", collected_at)
        # Should end with +00:00 or Z
        self.assertTrue(
            collected_at.endswith("+00:00") or collected_at.endswith("Z")
        )


class TestSchemaValidation(unittest.TestCase):
    def test_valid_payload(self):
        payload = {
            "schema_version": "1.0.0",
            "host": {"hostname": "test-host"},
            "collected_at": "2024-01-01T00:00:00+00:00",
            "devices": [
                {
                    "device_path": "/dev/sg0",
                    "device_type": "SATA_HDD",
                    "identity": {
                        "model_number": "ST4000DX001",
                        "serial_number": "ABC123",
                        "firmware_revision": "CC44",
                    },
                }
            ],
        }
        errors = validate_payload(payload)
        self.assertEqual(errors, [])

    def test_missing_schema_version(self):
        payload = {
            "host": {"hostname": "test"},
            "collected_at": "2024-01-01T00:00:00+00:00",
            "devices": [],
        }
        errors = validate_payload(payload)
        self.assertTrue(any("schema_version" in e for e in errors))

    def test_invalid_device_type(self):
        payload = {
            "schema_version": "1.0.0",
            "host": {"hostname": "test"},
            "collected_at": "2024-01-01T00:00:00+00:00",
            "devices": [
                {
                    "device_path": "/dev/sg0",
                    "device_type": "INVALID",
                    "identity": {
                        "model_number": "M",
                        "serial_number": "S",
                        "firmware_revision": "F",
                    },
                }
            ],
        }
        errors = validate_payload(payload)
        self.assertTrue(any("device_type" in e for e in errors))

    def test_missing_identity_fields(self):
        payload = {
            "schema_version": "1.0.0",
            "host": {"hostname": "test"},
            "collected_at": "2024-01-01T00:00:00+00:00",
            "devices": [
                {
                    "device_path": "/dev/sg0",
                    "device_type": "SATA_HDD",
                    "identity": {},
                }
            ],
        }
        errors = validate_payload(payload)
        self.assertTrue(any("model_number" in e for e in errors))
        self.assertTrue(any("serial_number" in e for e in errors))
        self.assertTrue(any("firmware_revision" in e for e in errors))


if __name__ == "__main__":
    unittest.main()
