"""Firmware compliance schemas."""

from typing import Optional

from pydantic import BaseModel


class FirmwareDeviceSummary(BaseModel):
    serial_number: str
    model_number: Optional[str] = None
    firmware_revision: Optional[str] = None
    hostname: Optional[str] = None


class FirmwareComplianceResponse(BaseModel):
    target_firmware: str
    total_devices: int
    compliant_count: int
    non_compliant_count: int
    compliant: list[FirmwareDeviceSummary]
    non_compliant: list[FirmwareDeviceSummary]
