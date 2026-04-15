"""Firmware compliance endpoint."""

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models.device import Device
from ..models.host import Host
from ..schemas.firmware import FirmwareComplianceResponse, FirmwareDeviceSummary

router = APIRouter()


@router.get("/firmware/compliance", response_model=FirmwareComplianceResponse)
def firmware_compliance(
    target_firmware: str = Query(..., description="Target firmware revision to check compliance against"),
    db: Session = Depends(get_db),
):
    """Check firmware compliance across all devices."""
    devices = db.query(Device).all()

    compliant = []
    non_compliant = []

    for device in devices:
        host = db.query(Host).filter(Host.id == device.host_id).first()
        summary = FirmwareDeviceSummary(
            serial_number=device.serial_number,
            model_number=device.model_number,
            firmware_revision=device.firmware_revision,
            hostname=host.hostname if host else None,
        )
        if device.firmware_revision == target_firmware:
            compliant.append(summary)
        else:
            non_compliant.append(summary)

    return FirmwareComplianceResponse(
        target_firmware=target_firmware,
        total_devices=len(devices),
        compliant_count=len(compliant),
        non_compliant_count=len(non_compliant),
        compliant=compliant,
        non_compliant=non_compliant,
    )
