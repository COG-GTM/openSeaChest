"""Pydantic models matching the shared ingestion schema v1.0.0."""

from datetime import datetime
from typing import Optional

from pydantic import BaseModel, Field


class SmartAttribute(BaseModel):
    id: int
    name: str
    status_flags: Optional[str] = None
    nominal: Optional[int] = None
    worst: Optional[int] = None
    raw_hex: Optional[str] = None
    raw_value: Optional[int] = None


class DSTInfo(BaseModel):
    supported: Optional[bool] = None
    hours_since_last: Optional[float] = None
    status_result: Optional[str] = None
    test_run: Optional[str] = None


class SmartInfo(BaseModel):
    status: Optional[str] = None
    tripped: Optional[bool] = None
    attributes: Optional[list[SmartAttribute]] = None
    last_dst: Optional[DSTInfo] = None


class DeviceIdentity(BaseModel):
    model_number: str
    serial_number: str
    firmware_revision: str
    world_wide_name: Optional[str] = None
    form_factor_inches: Optional[float] = None
    rotation_rate_rpm: Optional[int] = None
    is_ssd: Optional[bool] = None
    logical_sector_size_bytes: Optional[int] = None
    physical_sector_size_bytes: Optional[int] = None


class CapacityInfo(BaseModel):
    capacity_bytes: Optional[int] = None
    capacity_display: Optional[str] = None
    max_lba: Optional[int] = None
    native_max_lba: Optional[int] = None


class TemperatureInfo(BaseModel):
    current_celsius: Optional[float] = None
    highest_celsius: Optional[float] = None
    lowest_celsius: Optional[float] = None


class PowerOnInfo(BaseModel):
    power_on_hours: Optional[float] = None
    power_on_hours_display: Optional[str] = None


class InterfaceInfo(BaseModel):
    protocol: Optional[str] = None
    max_speed_gbps: Optional[float] = None
    negotiated_speed_gbps: Optional[float] = None
    specifications: Optional[list[str]] = None
    features: Optional[list[str]] = None


class WorkloadInfo(BaseModel):
    annualized_workload_rate_tb_yr: Optional[float] = None
    total_bytes_read: Optional[int] = None
    total_bytes_written: Optional[int] = None
    percentage_used_endurance: Optional[float] = None


class FirmwareInfo(BaseModel):
    revision: Optional[str] = None
    download_support: Optional[list[str]] = None


class SecurityInfo(BaseModel):
    encryption_support: Optional[str] = None
    ata_security: Optional[str] = None


class DevicePayload(BaseModel):
    device_path: str
    device_type: str
    identity: DeviceIdentity
    capacity: Optional[CapacityInfo] = None
    temperature: Optional[TemperatureInfo] = None
    power_on: Optional[PowerOnInfo] = None
    interface: Optional[InterfaceInfo] = None
    workload: Optional[WorkloadInfo] = None
    smart: Optional[SmartInfo] = None
    firmware: Optional[FirmwareInfo] = None
    security: Optional[SecurityInfo] = None


class HostInfo(BaseModel):
    hostname: str
    os: Optional[str] = None
    agent_version: Optional[str] = None


class IngestionPayload(BaseModel):
    schema_version: str = Field(default="1.0.0")
    host: HostInfo
    collected_at: datetime
    devices: list[DevicePayload]


class IngestionResponse(BaseModel):
    status: str = "ok"
    host: str
    devices_upserted: int
    snapshots_created: int
