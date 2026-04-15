"""HealthSnapshot model."""

import enum
from datetime import datetime, timezone

from sqlalchemy import Float, Integer, BigInteger, Boolean, DateTime, Enum, ForeignKey, JSON, String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from ..database import Base


class SmartStatus(str, enum.Enum):
    Good = "Good"
    Bad = "Bad"
    Unknown = "Unknown"


class HealthSnapshot(Base):
    __tablename__ = "health_snapshots"

    id: Mapped[int] = mapped_column(primary_key=True)
    device_id: Mapped[int] = mapped_column(ForeignKey("devices.id"), nullable=False)
    collected_at: Mapped[datetime] = mapped_column(
        DateTime, default=lambda: datetime.now(timezone.utc)
    )
    temperature_celsius: Mapped[float | None] = mapped_column(Float, nullable=True)
    highest_temp_celsius: Mapped[float | None] = mapped_column(Float, nullable=True)
    lowest_temp_celsius: Mapped[float | None] = mapped_column(Float, nullable=True)
    power_on_hours: Mapped[float | None] = mapped_column(Float, nullable=True)
    smart_status: Mapped[SmartStatus | None] = mapped_column(
        Enum(SmartStatus), nullable=True
    )
    smart_tripped: Mapped[bool | None] = mapped_column(Boolean, nullable=True)
    smart_attributes_json = mapped_column(JSON, nullable=True)
    annualized_workload_rate: Mapped[float | None] = mapped_column(Float, nullable=True)
    total_bytes_read: Mapped[int | None] = mapped_column(BigInteger, nullable=True)
    total_bytes_written: Mapped[int | None] = mapped_column(BigInteger, nullable=True)
    percentage_used_endurance: Mapped[float | None] = mapped_column(Float, nullable=True)

    device = relationship("Device", back_populates="health_snapshots")
