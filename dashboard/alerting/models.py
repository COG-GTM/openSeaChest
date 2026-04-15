"""SQLAlchemy models for the Alerting & Compliance Engine."""

from __future__ import annotations

import enum
from datetime import datetime, timezone

from sqlalchemy import (
    Boolean,
    Column,
    DateTime,
    Enum,
    Float,
    ForeignKey,
    Integer,
    String,
    Text,
    create_engine,
)
from sqlalchemy.orm import DeclarativeBase, Session, relationship, sessionmaker


class Base(DeclarativeBase):
    """Declarative base for all alerting models."""


class RuleType(enum.Enum):
    """Supported alert-rule types."""

    smart_tripped = "smart_tripped"
    temperature = "temperature"
    firmware = "firmware"
    poh = "poh"


class AlertRule(Base):
    """Persistent definition of an alert rule."""

    __tablename__ = "alert_rules"

    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String(255), nullable=False)
    description = Column(Text, nullable=True)
    rule_type = Column(Enum(RuleType), nullable=False)
    threshold_value = Column(Float, nullable=True)
    approved_firmware_list = Column(Text, nullable=True)  # JSON array string
    enabled = Column(Boolean, default=True, nullable=False)
    created_at = Column(
        DateTime, default=lambda: datetime.now(timezone.utc), nullable=False
    )
    updated_at = Column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
        nullable=False,
    )

    fired_alerts = relationship("FiredAlert", back_populates="rule")


class FiredAlert(Base):
    """Record of an alert that has been fired."""

    __tablename__ = "fired_alerts"

    id = Column(Integer, primary_key=True, autoincrement=True)
    rule_id = Column(Integer, ForeignKey("alert_rules.id"), nullable=False)
    device_serial = Column(String(255), nullable=False)
    device_model = Column(String(255), nullable=False)
    host = Column(String(255), nullable=False)
    fired_at = Column(
        DateTime, default=lambda: datetime.now(timezone.utc), nullable=False
    )
    message = Column(Text, nullable=False)
    resolved = Column(Boolean, default=False, nullable=False)
    resolved_at = Column(DateTime, nullable=True)
    notification_sent = Column(Boolean, default=False, nullable=False)

    rule = relationship("AlertRule", back_populates="fired_alerts")


class Device(Base):
    """Minimal device record for querying latest health data."""

    __tablename__ = "devices"

    serial_number = Column(String(255), primary_key=True)
    model_number = Column(String(255), nullable=True)
    firmware_revision = Column(String(255), nullable=True)
    host = Column(String(255), nullable=True)


class HealthSnapshot(Base):
    """Point-in-time health snapshot for a device."""

    __tablename__ = "health_snapshots"

    id = Column(Integer, primary_key=True, autoincrement=True)
    device_serial = Column(String(255), nullable=False, index=True)
    collected_at = Column(DateTime, nullable=False)
    temperature_celsius = Column(Float, nullable=True)
    power_on_hours = Column(Float, nullable=True)
    smart_status = Column(String(50), nullable=True)
    smart_tripped = Column(Boolean, nullable=True)


def init_db(url: str) -> sessionmaker:
    """Create engine, ensure tables exist, and return a session factory."""
    engine = create_engine(url)
    Base.metadata.create_all(engine)
    return sessionmaker(bind=engine)
