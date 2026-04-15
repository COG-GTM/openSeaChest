"""AlertRule model."""

import enum
from datetime import datetime, timezone

from sqlalchemy import String, Boolean, Float, DateTime, Enum, JSON
from sqlalchemy.orm import Mapped, mapped_column, relationship

from ..database import Base


class RuleType(str, enum.Enum):
    smart_tripped = "smart_tripped"
    temperature_threshold = "temperature_threshold"
    firmware_not_approved = "firmware_not_approved"
    poh_threshold = "poh_threshold"


class AlertRule(Base):
    __tablename__ = "alert_rules"

    id: Mapped[int] = mapped_column(primary_key=True)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    description: Mapped[str | None] = mapped_column(String(1024), nullable=True)
    rule_type: Mapped[RuleType] = mapped_column(Enum(RuleType), nullable=False)
    threshold_value: Mapped[float | None] = mapped_column(Float, nullable=True)
    approved_firmware_list = mapped_column(JSON, nullable=True)
    enabled: Mapped[bool] = mapped_column(Boolean, default=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime, default=lambda: datetime.now(timezone.utc)
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime,
        default=lambda: datetime.now(timezone.utc),
        onupdate=lambda: datetime.now(timezone.utc),
    )

    fired_alerts = relationship("FiredAlert", back_populates="rule")
