"""Deduplication logic – prevents the same alert from firing repeatedly."""

from __future__ import annotations

from datetime import datetime, timedelta, timezone

from sqlalchemy.orm import Session

from dashboard.alerting.models import FiredAlert


def is_duplicate(
    session: Session,
    rule_id: int,
    device_serial: str,
    window_seconds: int = 86400,
) -> bool:
    """Return ``True`` if an unresolved alert for the same rule + device
    already exists within the deduplication window.
    """
    cutoff = datetime.now(timezone.utc) - timedelta(seconds=window_seconds)
    existing = (
        session.query(FiredAlert)
        .filter(
            FiredAlert.rule_id == rule_id,
            FiredAlert.device_serial == device_serial,
            FiredAlert.resolved.is_(False),
            FiredAlert.fired_at >= cutoff,
        )
        .first()
    )
    return existing is not None


def record_fired_alert(
    session: Session,
    rule_id: int,
    device_serial: str,
    device_model: str,
    host: str,
    message: str,
) -> FiredAlert:
    """Insert a new ``FiredAlert`` row and flush (but do not commit)."""
    alert = FiredAlert(
        rule_id=rule_id,
        device_serial=device_serial,
        device_model=device_model,
        host=host,
        message=message,
    )
    session.add(alert)
    session.flush()
    return alert
