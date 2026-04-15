"""Event-driven evaluation triggered on data ingestion.

Call ``on_ingest`` when the WI-2 backend receives a new ingestion payload.
The function evaluates all enabled alert rules against every device in the
payload, respecting deduplication and notification settings.
"""

from __future__ import annotations

import logging
from datetime import datetime, timezone
from typing import Any, Dict, List

from sqlalchemy.orm import Session

from dashboard.alerting import config
from dashboard.alerting.evaluator import evaluate_payload
from dashboard.alerting.models import Device, FiredAlert, HealthSnapshot

logger = logging.getLogger(__name__)


def _upsert_device_and_snapshot(
    session: Session,
    device_data: Dict[str, Any],
    host: str,
    collected_at: str,
) -> None:
    """Upsert a ``Device`` row and insert a ``HealthSnapshot``."""
    identity = device_data.get("identity", {})
    serial = identity.get("serial_number", "")
    if not serial:
        return

    device = session.get(Device, serial)
    if device is None:
        device = Device(serial_number=serial)
        session.add(device)

    device.model_number = identity.get("model_number")
    device.firmware_revision = identity.get("firmware_revision")
    device.host = host

    temp = device_data.get("temperature", {})
    poh = device_data.get("power_on", {})
    smart = device_data.get("smart", {})

    try:
        ts = datetime.fromisoformat(collected_at.replace("Z", "+00:00"))
    except (ValueError, AttributeError):
        ts = datetime.now(timezone.utc)

    snapshot = HealthSnapshot(
        device_serial=serial,
        collected_at=ts,
        temperature_celsius=temp.get("current_celsius"),
        power_on_hours=poh.get("power_on_hours"),
        smart_status=smart.get("status"),
        smart_tripped=smart.get("tripped"),
    )
    session.add(snapshot)
    session.flush()


def on_ingest(
    session: Session,
    payload: Dict[str, Any],
    webhook_url: str | None = None,
    dedup_window_seconds: int | None = None,
) -> List[FiredAlert]:
    """Primary entry-point for event-driven alert evaluation.

    1. Upserts ``Device`` / ``HealthSnapshot`` rows.
    2. Evaluates all enabled rules against each device.
    3. Returns any newly-fired alerts.
    """
    wh_url = webhook_url if webhook_url is not None else config.WEBHOOK_URL
    dedup = (
        dedup_window_seconds
        if dedup_window_seconds is not None
        else config.DEDUP_WINDOW_SECONDS
    )

    host = payload.get("host", {}).get("hostname", "unknown")
    collected_at = payload.get("collected_at", "")

    for device_data in payload.get("devices", []):
        _upsert_device_and_snapshot(session, device_data, host, collected_at)

    session.flush()

    fired = evaluate_payload(
        session,
        payload,
        webhook_url=wh_url,
        dedup_window_seconds=dedup,
    )
    session.commit()
    return fired
