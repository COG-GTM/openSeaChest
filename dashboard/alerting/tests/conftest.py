"""Shared fixtures for alerting engine tests."""

from __future__ import annotations

import json
from typing import Any, Dict

import pytest
from sqlalchemy import create_engine
from sqlalchemy.orm import Session, sessionmaker

from dashboard.alerting.models import AlertRule, Base, RuleType


@pytest.fixture()
def db_session() -> Session:
    """In-memory SQLite session with all tables created."""
    engine = create_engine("sqlite:///:memory:")
    Base.metadata.create_all(engine)
    SessionLocal = sessionmaker(bind=engine)
    session = SessionLocal()
    yield session
    session.close()


@pytest.fixture()
def smart_rule(db_session: Session) -> AlertRule:
    rule = AlertRule(
        name="SMART Trip Alert",
        description="Fires when SMART is tripped",
        rule_type=RuleType.smart_tripped,
        enabled=True,
    )
    db_session.add(rule)
    db_session.commit()
    return rule


@pytest.fixture()
def temperature_rule(db_session: Session) -> AlertRule:
    rule = AlertRule(
        name="Temperature Alert",
        description="Fires when temperature exceeds threshold",
        rule_type=RuleType.temperature,
        threshold_value=55.0,
        enabled=True,
    )
    db_session.add(rule)
    db_session.commit()
    return rule


@pytest.fixture()
def firmware_rule(db_session: Session) -> AlertRule:
    rule = AlertRule(
        name="Firmware Compliance Alert",
        description="Fires when firmware is not approved",
        rule_type=RuleType.firmware,
        approved_firmware_list=json.dumps(["CC49", "CC50", "DN02"]),
        enabled=True,
    )
    db_session.add(rule)
    db_session.commit()
    return rule


@pytest.fixture()
def poh_rule(db_session: Session) -> AlertRule:
    rule = AlertRule(
        name="POH Alert",
        description="Fires when power-on hours exceed threshold",
        rule_type=RuleType.poh,
        threshold_value=40000.0,
        enabled=True,
    )
    db_session.add(rule)
    db_session.commit()
    return rule


@pytest.fixture()
def sample_device_ok() -> Dict[str, Any]:
    """A healthy device that should not trigger any default rules."""
    return {
        "device_path": "/dev/sg1",
        "device_type": "SATA_HDD",
        "identity": {
            "model_number": "ST4000DX001-1CE168",
            "serial_number": "ZQ3034X7R",
            "firmware_revision": "CC49",
        },
        "temperature": {"current_celsius": 35},
        "power_on": {"power_on_hours": 10000},
        "smart": {"status": "Good", "tripped": False},
        "firmware": {"revision": "CC49"},
    }


@pytest.fixture()
def sample_device_bad_smart() -> Dict[str, Any]:
    """Device with SMART tripped."""
    return {
        "device_path": "/dev/sg2",
        "device_type": "SATA_HDD",
        "identity": {
            "model_number": "ST4000DX001-1CE168",
            "serial_number": "ZQ3034X7R",
            "firmware_revision": "CC49",
        },
        "temperature": {"current_celsius": 35},
        "power_on": {"power_on_hours": 10000},
        "smart": {"status": "Bad", "tripped": True},
        "firmware": {"revision": "CC49"},
    }


@pytest.fixture()
def sample_payload(sample_device_ok: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "schema_version": "1.0.0",
        "host": {"hostname": "server-01", "os": "Linux 6.1", "agent_version": "0.1.0"},
        "collected_at": "2024-01-15T10:30:00Z",
        "devices": [sample_device_ok],
    }


@pytest.fixture()
def sample_payload_bad_smart(sample_device_bad_smart: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "schema_version": "1.0.0",
        "host": {"hostname": "server-01", "os": "Linux 6.1", "agent_version": "0.1.0"},
        "collected_at": "2024-01-15T10:30:00Z",
        "devices": [sample_device_bad_smart],
    }
