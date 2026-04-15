"""Device listing and health history endpoints."""

from datetime import datetime
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models.device import Device
from ..models.host import Host
from ..models.health_snapshot import HealthSnapshot
from ..schemas.device import (
    DeviceResponse,
    DeviceListResponse,
    HealthSnapshotResponse,
    HealthHistoryResponse,
)

router = APIRouter()


def _device_to_response(device: Device, db: Session) -> DeviceResponse:
    host = db.query(Host).filter(Host.id == device.host_id).first()
    return DeviceResponse(
        id=device.id,
        host_id=device.host_id,
        hostname=host.hostname if host else None,
        serial_number=device.serial_number,
        model_number=device.model_number,
        firmware_revision=device.firmware_revision,
        world_wide_name=device.world_wide_name,
        device_type=device.device_type.value if device.device_type else None,
        device_path=device.device_path,
        form_factor_inches=device.form_factor_inches,
        rotation_rate_rpm=device.rotation_rate_rpm,
        is_ssd=device.is_ssd,
        logical_sector_size_bytes=device.logical_sector_size_bytes,
        physical_sector_size_bytes=device.physical_sector_size_bytes,
        capacity_bytes=device.capacity_bytes,
        max_lba=device.max_lba,
        protocol=device.protocol,
        max_speed_gbps=device.max_speed_gbps,
        negotiated_speed_gbps=device.negotiated_speed_gbps,
        encryption_support=device.encryption_support,
        ata_security=device.ata_security,
        first_seen=device.first_seen,
        last_seen=device.last_seen,
    )


@router.get("/devices", response_model=DeviceListResponse)
def list_devices(
    device_type: Optional[str] = Query(None),
    host: Optional[str] = Query(None),
    smart_status: Optional[str] = Query(None),
    limit: int = Query(50, ge=1, le=1000),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
):
    """List all devices with optional filters and pagination."""
    query = db.query(Device)

    if device_type:
        query = query.filter(Device.device_type == device_type)

    if host:
        query = query.join(Host).filter(Host.hostname == host)

    if smart_status:
        # Filter by latest snapshot's smart_status
        from sqlalchemy import func

        latest_snapshot_subq = (
            db.query(
                HealthSnapshot.device_id,
                func.max(HealthSnapshot.collected_at).label("max_collected"),
            )
            .group_by(HealthSnapshot.device_id)
            .subquery()
        )
        query = query.join(
            latest_snapshot_subq,
            Device.id == latest_snapshot_subq.c.device_id,
        ).join(
            HealthSnapshot,
            (HealthSnapshot.device_id == Device.id)
            & (HealthSnapshot.collected_at == latest_snapshot_subq.c.max_collected),
        ).filter(HealthSnapshot.smart_status == smart_status)

    total = query.count()
    devices = query.offset(offset).limit(limit).all()

    return DeviceListResponse(
        total=total,
        offset=offset,
        limit=limit,
        devices=[_device_to_response(d, db) for d in devices],
    )


@router.get("/devices/{serial}", response_model=DeviceResponse)
def get_device(serial: str, db: Session = Depends(get_db)):
    """Get device details by serial number."""
    device = db.query(Device).filter(Device.serial_number == serial).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")
    return _device_to_response(device, db)


@router.get("/devices/{serial}/health/history", response_model=HealthHistoryResponse)
def get_health_history(
    serial: str,
    since: Optional[datetime] = Query(None),
    limit: int = Query(100, ge=1, le=10000),
    db: Session = Depends(get_db),
):
    """Get time-series health snapshots for a device."""
    device = db.query(Device).filter(Device.serial_number == serial).first()
    if not device:
        raise HTTPException(status_code=404, detail="Device not found")

    query = (
        db.query(HealthSnapshot)
        .filter(HealthSnapshot.device_id == device.id)
        .order_by(HealthSnapshot.collected_at.desc())
    )

    if since:
        query = query.filter(HealthSnapshot.collected_at >= since)

    snapshots = query.limit(limit).all()

    return HealthHistoryResponse(
        serial_number=serial,
        total=len(snapshots),
        snapshots=[
            HealthSnapshotResponse(
                id=s.id,
                device_id=s.device_id,
                collected_at=s.collected_at,
                temperature_celsius=s.temperature_celsius,
                highest_temp_celsius=s.highest_temp_celsius,
                lowest_temp_celsius=s.lowest_temp_celsius,
                power_on_hours=s.power_on_hours,
                smart_status=s.smart_status.value if s.smart_status else None,
                smart_tripped=s.smart_tripped,
                smart_attributes_json=s.smart_attributes_json,
                annualized_workload_rate=s.annualized_workload_rate,
                total_bytes_read=s.total_bytes_read,
                total_bytes_written=s.total_bytes_written,
                percentage_used_endurance=s.percentage_used_endurance,
            )
            for s in snapshots
        ],
    )
