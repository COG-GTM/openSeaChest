"""POST /api/v1/hosts/{host}/collect endpoint."""

from datetime import datetime, timezone

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from ..database import get_db
from ..models.host import Host
from ..models.device import Device, DeviceType
from ..models.health_snapshot import HealthSnapshot, SmartStatus
from ..schemas.ingestion import IngestionPayload, IngestionResponse

router = APIRouter()


@router.post("/hosts/{host}/collect", response_model=IngestionResponse)
def collect(host: str, payload: IngestionPayload, db: Session = Depends(get_db)):
    """Ingest device data from a collection agent."""
    now = datetime.now(timezone.utc)

    # Upsert host
    host_record = db.query(Host).filter(Host.hostname == host).first()
    if host_record is None:
        host_record = Host(
            hostname=host,
            os=payload.host.os,
            first_seen=now,
            last_seen=now,
        )
        db.add(host_record)
        db.flush()
    else:
        host_record.os = payload.host.os or host_record.os
        host_record.last_seen = now

    devices_upserted = 0
    snapshots_created = 0

    for dev in payload.devices:
        identity = dev.identity

        # Upsert device by serial_number
        device_record = (
            db.query(Device)
            .filter(Device.serial_number == identity.serial_number)
            .first()
        )

        # Parse device_type enum safely
        device_type_val = None
        try:
            device_type_val = DeviceType(dev.device_type)
        except ValueError:
            pass

        if device_record is None:
            device_record = Device(
                host_id=host_record.id,
                serial_number=identity.serial_number,
                model_number=identity.model_number,
                firmware_revision=identity.firmware_revision,
                world_wide_name=identity.world_wide_name,
                device_type=device_type_val,
                device_path=dev.device_path,
                form_factor_inches=identity.form_factor_inches,
                rotation_rate_rpm=identity.rotation_rate_rpm,
                is_ssd=identity.is_ssd,
                logical_sector_size_bytes=identity.logical_sector_size_bytes,
                physical_sector_size_bytes=identity.physical_sector_size_bytes,
                capacity_bytes=dev.capacity.capacity_bytes if dev.capacity else None,
                max_lba=dev.capacity.max_lba if dev.capacity else None,
                protocol=dev.interface.protocol if dev.interface else None,
                max_speed_gbps=dev.interface.max_speed_gbps if dev.interface else None,
                negotiated_speed_gbps=dev.interface.negotiated_speed_gbps if dev.interface else None,
                encryption_support=dev.security.encryption_support if dev.security else None,
                ata_security=dev.security.ata_security if dev.security else None,
                first_seen=now,
                last_seen=now,
            )
            db.add(device_record)
            db.flush()
        else:
            device_record.host_id = host_record.id
            device_record.model_number = identity.model_number
            device_record.firmware_revision = identity.firmware_revision
            device_record.world_wide_name = identity.world_wide_name
            device_record.device_type = device_type_val
            device_record.device_path = dev.device_path
            device_record.form_factor_inches = identity.form_factor_inches
            device_record.rotation_rate_rpm = identity.rotation_rate_rpm
            device_record.is_ssd = identity.is_ssd
            device_record.logical_sector_size_bytes = identity.logical_sector_size_bytes
            device_record.physical_sector_size_bytes = identity.physical_sector_size_bytes
            if dev.capacity:
                device_record.capacity_bytes = dev.capacity.capacity_bytes
                device_record.max_lba = dev.capacity.max_lba
            if dev.interface:
                device_record.protocol = dev.interface.protocol
                device_record.max_speed_gbps = dev.interface.max_speed_gbps
                device_record.negotiated_speed_gbps = dev.interface.negotiated_speed_gbps
            if dev.security:
                device_record.encryption_support = dev.security.encryption_support
                device_record.ata_security = dev.security.ata_security
            device_record.last_seen = now

        devices_upserted += 1

        # Build health snapshot
        smart_status_val = None
        if dev.smart and dev.smart.status:
            try:
                smart_status_val = SmartStatus(dev.smart.status)
            except ValueError:
                pass

        snapshot = HealthSnapshot(
            device_id=device_record.id,
            collected_at=payload.collected_at,
            temperature_celsius=dev.temperature.current_celsius if dev.temperature else None,
            highest_temp_celsius=dev.temperature.highest_celsius if dev.temperature else None,
            lowest_temp_celsius=dev.temperature.lowest_celsius if dev.temperature else None,
            power_on_hours=dev.power_on.power_on_hours if dev.power_on else None,
            smart_status=smart_status_val,
            smart_tripped=dev.smart.tripped if dev.smart else None,
            smart_attributes_json=(
                [attr.model_dump() for attr in dev.smart.attributes]
                if dev.smart and dev.smart.attributes
                else None
            ),
            annualized_workload_rate=(
                dev.workload.annualized_workload_rate_tb_yr if dev.workload else None
            ),
            total_bytes_read=dev.workload.total_bytes_read if dev.workload else None,
            total_bytes_written=dev.workload.total_bytes_written if dev.workload else None,
            percentage_used_endurance=(
                dev.workload.percentage_used_endurance if dev.workload else None
            ),
        )
        db.add(snapshot)
        snapshots_created += 1

    db.commit()

    return IngestionResponse(
        status="ok",
        host=host,
        devices_upserted=devices_upserted,
        snapshots_created=snapshots_created,
    )
