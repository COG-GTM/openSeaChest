"""Background worker for periodic sweep evaluation.

Uses APScheduler to run a configurable-interval job that queries the latest
health snapshot for each device and evaluates all enabled rules.
"""

from __future__ import annotations

import logging
from typing import Any, Dict, List

from sqlalchemy import func
from sqlalchemy.orm import Session, sessionmaker

from dashboard.alerting import config
from dashboard.alerting.evaluator import evaluate_payload
from dashboard.alerting.models import Device, HealthSnapshot, init_db

logger = logging.getLogger(__name__)


def _build_payload_from_db(session: Session) -> Dict[str, Any]:
    """Build a synthetic ingestion payload from the latest health snapshots."""

    # Sub-query: latest snapshot per device
    latest_sq = (
        session.query(
            HealthSnapshot.device_serial,
            func.max(HealthSnapshot.id).label("max_id"),
        )
        .group_by(HealthSnapshot.device_serial)
        .subquery()
    )

    snapshots = (
        session.query(HealthSnapshot)
        .join(latest_sq, HealthSnapshot.id == latest_sq.c.max_id)
        .all()
    )

    devices_payload: List[Dict[str, Any]] = []
    for snap in snapshots:
        device = session.get(Device, snap.device_serial)
        fw_rev = device.firmware_revision if device else None
        model = device.model_number if device else None
        host = device.host if device else "unknown"

        devices_payload.append(
            {
                "identity": {
                    "serial_number": snap.device_serial,
                    "model_number": model or "unknown",
                    "firmware_revision": fw_rev or "unknown",
                },
                "temperature": {"current_celsius": snap.temperature_celsius},
                "power_on": {"power_on_hours": snap.power_on_hours},
                "smart": {
                    "status": snap.smart_status,
                    "tripped": snap.smart_tripped or False,
                },
                "firmware": {"revision": fw_rev or "unknown"},
            }
        )

    # Use the host from the first device (all in one sweep share DB context).
    host = "sweep"
    if devices_payload:
        first_device = session.get(
            Device, devices_payload[0]["identity"]["serial_number"]
        )
        if first_device and first_device.host:
            host = first_device.host

    return {
        "schema_version": "1.0.0",
        "host": {"hostname": host},
        "collected_at": "",
        "devices": devices_payload,
    }


def run_sweep(session: Session, webhook_url: str = "", dedup_window_seconds: int = 86400) -> int:
    """Execute a single sweep: build payload from DB and evaluate rules.

    Returns the number of newly-fired alerts.
    """
    payload = _build_payload_from_db(session)
    if not payload["devices"]:
        logger.info("No devices in database; sweep is a no-op.")
        return 0

    fired = evaluate_payload(
        session,
        payload,
        webhook_url=webhook_url,
        dedup_window_seconds=dedup_window_seconds,
    )
    session.commit()
    logger.info("Sweep complete – %d alert(s) fired.", len(fired))
    return len(fired)


def start_worker(
    database_url: str | None = None,
    webhook_url: str | None = None,
    interval_seconds: int | None = None,
) -> None:  # pragma: no cover
    """Start the APScheduler background worker (blocking).

    Intended to be called from a CLI entry-point or process manager.
    """
    from apscheduler.schedulers.blocking import BlockingScheduler

    db_url = database_url or config.DATABASE_URL
    wh_url = webhook_url or config.WEBHOOK_URL
    interval = interval_seconds or config.SWEEP_INTERVAL_SECONDS

    SessionFactory = init_db(db_url)

    scheduler = BlockingScheduler()

    def _job() -> None:
        session = SessionFactory()
        try:
            run_sweep(session, webhook_url=wh_url)
        finally:
            session.close()

    scheduler.add_job(_job, "interval", seconds=interval)
    logger.info("Starting sweep worker (interval=%ds)…", interval)
    scheduler.start()
