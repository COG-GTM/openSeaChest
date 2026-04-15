"""Response schemas for device endpoints."""

from datetime import datetime
from typing import Optional

from pydantic import BaseModel


class DeviceResponse(BaseModel):
    id: int
    host_id: int
    hostname: Optional[str] = None
    serial_number: str
    model_number: Optional[str] = None
    firmware_revision: Optional[str] = None
    world_wide_name: Optional[str] = None
    device_type: Optional[str] = None
    device_path: Optional[str] = None
    form_factor_inches: Optional[float] = None
    rotation_rate_rpm: Optional[int] = None
    is_ssd: Optional[bool] = None
    logical_sector_size_bytes: Optional[int] = None
    physical_sector_size_bytes: Optional[int] = None
    capacity_bytes: Optional[int] = None
    max_lba: Optional[int] = None
    protocol: Optional[str] = None
    max_speed_gbps: Optional[float] = None
    negotiated_speed_gbps: Optional[float] = None
    encryption_support: Optional[str] = None
    ata_security: Optional[str] = None
    first_seen: datetime
    last_seen: datetime

    model_config = {"from_attributes": True}


class DeviceListResponse(BaseModel):
    total: int
    offset: int
    limit: int
    devices: list[DeviceResponse]


class HealthSnapshotResponse(BaseModel):
    id: int
    device_id: int
    collected_at: datetime
    temperature_celsius: Optional[float] = None
    highest_temp_celsius: Optional[float] = None
    lowest_temp_celsius: Optional[float] = None
    power_on_hours: Optional[float] = None
    smart_status: Optional[str] = None
    smart_tripped: Optional[bool] = None
    smart_attributes_json: Optional[dict | list] = None
    annualized_workload_rate: Optional[float] = None
    total_bytes_read: Optional[int] = None
    total_bytes_written: Optional[int] = None
    percentage_used_endurance: Optional[float] = None

    model_config = {"from_attributes": True}


class HealthHistoryResponse(BaseModel):
    serial_number: str
    total: int
    snapshots: list[HealthSnapshotResponse]
