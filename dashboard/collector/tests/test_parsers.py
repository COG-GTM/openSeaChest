"""Unit tests for parsers."""

from __future__ import annotations

import os
import unittest

from dashboard.collector.parsers.scan_parser import parse_scan_output
from dashboard.collector.parsers.device_info_parser import parse_device_info
from dashboard.collector.parsers.smart_parser import (
    parse_smart_check,
    parse_smart_attributes,
)

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def _load_fixture(name: str) -> str:
    with open(os.path.join(FIXTURES_DIR, name), "r") as f:
        return f.read()


class TestScanParser(unittest.TestCase):
    def test_parse_scan_output(self):
        text = _load_fixture("scan_output.txt")
        devices = parse_scan_output(text)
        self.assertEqual(len(devices), 4)
        self.assertEqual(devices[0]["device_path"], "/dev/sg0")
        self.assertEqual(devices[1]["device_path"], "/dev/sg1")
        self.assertEqual(devices[2]["device_path"], "/dev/sg2")
        self.assertEqual(devices[3]["device_path"], "/dev/sg3")

    def test_parse_scan_with_model_info(self):
        text = _load_fixture("scan_output.txt")
        devices = parse_scan_output(text)
        # First device should have model info
        self.assertIn("model", devices[0])
        self.assertIn("capacity", devices[0])

    def test_parse_empty_scan(self):
        text = """===============================================================================
 openSeaChest_Basics - Seagate drive utilities
===============================================================================
No drives found.
"""
        devices = parse_scan_output(text)
        self.assertEqual(len(devices), 0)

    def test_parse_windows_paths(self):
        text = """===============================================================================
PD0 - SEAGATE  ST4000DX001 - CC44 - 4.00 TB
PD1 - SEAGATE  ST120FP0021 - B770 - 120.03 GB
"""
        devices = parse_scan_output(text)
        self.assertEqual(len(devices), 2)
        self.assertEqual(devices[0]["device_path"], "PD0")
        self.assertEqual(devices[1]["device_path"], "PD1")


class TestDeviceInfoParser(unittest.TestCase):
    def test_sata_hdd(self):
        text = _load_fixture("sata_hdd_info.txt")
        device = parse_device_info(text, "/dev/sg0")

        self.assertEqual(device["device_path"], "/dev/sg0")
        self.assertEqual(device["device_type"], "SATA_HDD")

        identity = device["identity"]
        self.assertEqual(identity["model_number"], "ST4000DX001-1CE168")
        self.assertEqual(identity["serial_number"], "ZQ3034X7R")
        self.assertEqual(identity["firmware_revision"], "CC44")
        self.assertEqual(identity["world_wide_name"], "500Q0C5007A5FCF19")
        self.assertEqual(identity["form_factor_inches"], 3.5)
        self.assertEqual(identity["rotation_rate_rpm"], 5900)
        self.assertFalse(identity["is_ssd"])
        self.assertEqual(identity["logical_sector_size_bytes"], 512)
        self.assertEqual(identity["physical_sector_size_bytes"], 4096)

        # Capacity
        self.assertIsNotNone(device["capacity"]["capacity_bytes"])
        self.assertEqual(device["capacity"]["max_lba"], 7814037167)
        self.assertEqual(device["capacity"]["native_max_lba"], 7814037167)

        # Temperature
        self.assertEqual(device["temperature"]["current_celsius"], 25.0)
        self.assertEqual(device["temperature"]["highest_celsius"], 40.0)
        self.assertEqual(device["temperature"]["lowest_celsius"], 18.0)

        # Power on
        self.assertEqual(device["power_on"]["power_on_hours"], 97.0)
        self.assertEqual(device["power_on"]["power_on_hours_display"], "4 days 1 hour")

        # Interface
        self.assertEqual(device["interface"]["protocol"], "SATA")
        self.assertEqual(device["interface"]["max_speed_gbps"], 6.0)
        self.assertEqual(device["interface"]["negotiated_speed_gbps"], 6.0)
        self.assertIn("ACS-2", device["interface"]["specifications"])
        self.assertIn("SATA 3.1", device["interface"]["specifications"])
        self.assertIn("NCQ", device["interface"]["features"])
        self.assertIn("SMART", device["interface"]["features"])

        # Workload
        self.assertEqual(
            device["workload"]["annualized_workload_rate_tb_yr"], 2.51
        )
        self.assertIsNotNone(device["workload"]["total_bytes_read"])
        self.assertIsNotNone(device["workload"]["total_bytes_written"])
        self.assertIsNone(device["workload"]["percentage_used_endurance"])

        # SMART
        self.assertEqual(device["smart"]["status"], "Good")
        self.assertFalse(device["smart"]["tripped"])

        # DST
        self.assertTrue(device["smart"]["last_dst"]["supported"])
        self.assertEqual(device["smart"]["last_dst"]["hours_since_last"], 0.0)
        self.assertEqual(device["smart"]["last_dst"]["status_result"], "0x0")
        self.assertEqual(device["smart"]["last_dst"]["test_run"], "0x1")

        # Firmware
        self.assertEqual(device["firmware"]["revision"], "CC44")
        self.assertEqual(
            device["firmware"]["download_support"],
            ["Immediate", "Segmented"],
        )

        # Security
        self.assertEqual(
            device["security"]["encryption_support"], "Not Supported"
        )
        self.assertEqual(
            device["security"]["ata_security"], "Supported, Frozen"
        )

    def test_sata_ssd(self):
        text = _load_fixture("sata_ssd_info.txt")
        device = parse_device_info(text, "/dev/sg1")

        self.assertEqual(device["device_type"], "SATA_SSD")
        self.assertTrue(device["identity"]["is_ssd"])
        self.assertIsNone(device["identity"]["rotation_rate_rpm"])
        self.assertEqual(device["identity"]["model_number"], "ST120FP0021")
        self.assertEqual(device["identity"]["form_factor_inches"], 2.5)

        # Capacity in GB
        self.assertIsNotNone(device["capacity"]["capacity_bytes"])
        capacity_gb = device["capacity"]["capacity_bytes"] / 1e9
        self.assertAlmostEqual(capacity_gb, 120.03, places=0)

        # DST not supported
        self.assertFalse(device["smart"]["last_dst"]["supported"])

        # SSD endurance indicator
        self.assertEqual(
            device["workload"]["percentage_used_endurance"], 0.0
        )

        # Total bytes read in TB
        self.assertIsNotNone(device["workload"]["total_bytes_read"])
        read_tb = device["workload"]["total_bytes_read"] / 1e12
        self.assertAlmostEqual(read_tb, 2.35, places=1)

        # Interface speed
        self.assertEqual(device["interface"]["max_speed_gbps"], 6.0)
        self.assertEqual(device["interface"]["negotiated_speed_gbps"], 3.0)

        self.assertIn("TRIM", device["interface"]["features"])

    def test_sas_hdd(self):
        text = _load_fixture("sas_hdd_info.txt")
        device = parse_device_info(text, "/dev/sg2")

        self.assertEqual(device["device_type"], "SAS_HDD")
        self.assertFalse(device["identity"]["is_ssd"])
        self.assertEqual(device["identity"]["rotation_rate_rpm"], 7200)
        self.assertEqual(device["interface"]["protocol"], "SAS")
        self.assertIn("SPC-4", device["interface"]["specifications"])

        # SAS interface speed (Port 0)
        self.assertEqual(device["interface"]["max_speed_gbps"], 6.0)
        self.assertEqual(device["interface"]["negotiated_speed_gbps"], 3.0)

        # Encryption
        self.assertEqual(
            device["security"]["encryption_support"], "Self Encrypting"
        )
        self.assertEqual(device["security"]["ata_security"], "Not Supported")

        # Native MaxLBA is "Not Reported" -> None
        self.assertIsNone(device["capacity"]["native_max_lba"])

        # DST supported
        self.assertTrue(device["smart"]["last_dst"]["supported"])
        self.assertEqual(
            device["smart"]["last_dst"]["hours_since_last"], 548.23
        )

        self.assertIn("EPC", device["interface"]["features"])
        self.assertIn("TCG", device["interface"]["features"])

    def test_sas_ssd(self):
        text = _load_fixture("sas_ssd_info.txt")
        device = parse_device_info(text, "/dev/sg3")

        self.assertEqual(device["device_type"], "SAS_SSD")
        self.assertTrue(device["identity"]["is_ssd"])
        self.assertIsNone(device["identity"]["rotation_rate_rpm"])
        self.assertEqual(device["interface"]["protocol"], "SAS")

        # SAS SSD interface speed (Port 0)
        self.assertEqual(device["interface"]["max_speed_gbps"], 12.0)
        self.assertEqual(device["interface"]["negotiated_speed_gbps"], 3.0)

        # Endurance
        self.assertEqual(
            device["workload"]["percentage_used_endurance"], 1.0
        )

        # Firmware download
        self.assertEqual(
            device["firmware"]["download_support"],
            ["Immediate", "Segmented", "Deferred"],
        )

    def test_nvme(self):
        text = _load_fixture("nvme_info.txt")
        device = parse_device_info(text, "/dev/nvme0")

        self.assertEqual(device["device_type"], "NVMe")
        self.assertTrue(device["identity"]["is_ssd"])
        self.assertIsNone(device["identity"]["rotation_rate_rpm"])
        self.assertEqual(device["interface"]["protocol"], "NVMe")
        self.assertIn("NVMe 1.4", device["interface"]["specifications"])

        # NVMe capacity
        self.assertIsNotNone(device["capacity"]["capacity_bytes"])

        # Temperature
        self.assertEqual(device["temperature"]["current_celsius"], 38.0)
        self.assertEqual(device["temperature"]["highest_celsius"], 52.0)
        self.assertEqual(device["temperature"]["lowest_celsius"], 22.0)

        # DST not supported for NVMe
        self.assertFalse(device["smart"]["last_dst"]["supported"])

        # Endurance
        self.assertEqual(
            device["workload"]["percentage_used_endurance"], 2.0
        )

        # Total bytes in TB
        self.assertIsNotNone(device["workload"]["total_bytes_read"])
        read_tb = device["workload"]["total_bytes_read"] / 1e12
        self.assertAlmostEqual(read_tb, 15.3, places=0)


class TestSmartParser(unittest.TestCase):
    def test_parse_smart_check_good(self):
        text = """===============================================================================
        SMART Status: Good
"""
        result = parse_smart_check(text)
        self.assertEqual(result["status"], "Good")
        self.assertFalse(result["tripped"])

    def test_parse_smart_check_bad(self):
        text = """===============================================================================
        SMART Status: Bad
"""
        result = parse_smart_check(text)
        self.assertEqual(result["status"], "Bad")
        self.assertTrue(result["tripped"])

    def test_parse_smart_check_missing(self):
        text = """===============================================================================
No SMART data available.
"""
        result = parse_smart_check(text)
        self.assertEqual(result["status"], "Unknown")
        self.assertFalse(result["tripped"])

    def test_parse_smart_attributes(self):
        text = _load_fixture("smart_attributes.txt")
        attrs = parse_smart_attributes(text)

        self.assertEqual(len(attrs), 21)

        # Check first attribute: Read Error Rate
        attr0 = attrs[0]
        self.assertEqual(attr0["id"], 1)
        self.assertEqual(attr0["name"], "Read Error Rate")
        self.assertEqual(attr0["status_flags"], "0x000F")
        self.assertEqual(attr0["nominal"], 0x75)  # 117
        self.assertEqual(attr0["worst"], 0x63)  # 99
        self.assertEqual(attr0["raw_hex"], "0x0000000969AFB4")
        self.assertEqual(attr0["raw_value"], 0x0000000969AFB4)

        # Check hex to decimal conversion
        self.assertEqual(attr0["nominal"], 117)
        self.assertEqual(attr0["worst"], 99)

        # Check attribute with ID 188 (Command Timeout)
        attr_188 = next(a for a in attrs if a["id"] == 188)
        self.assertEqual(attr_188["name"], "Command Timeout")
        self.assertEqual(attr_188["worst"], 0xFD)  # 253

        # Check attribute with multi-word name containing hyphen
        attr_199 = next(a for a in attrs if a["id"] == 199)
        self.assertEqual(attr_199["name"], "Ultra DMA CRC Error")
        self.assertEqual(attr_199["nominal"], 0xC8)  # 200
        self.assertEqual(attr_199["worst"], 0xC8)  # 200

    def test_parse_smart_attributes_empty(self):
        text = """===============================================================================
No SMART attributes available.
"""
        attrs = parse_smart_attributes(text)
        self.assertEqual(len(attrs), 0)


if __name__ == "__main__":
    unittest.main()
