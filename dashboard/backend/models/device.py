"""Device model."""

import enum
from datetime import datetime, timezone

from sqlalchemy import String, Integer, BigInteger, Float, Boolean, DateTime, Enum, ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, relationship

from ..database import Base


class DeviceType(str, enum.Enum):
    SATA_HDD = "SATA_HDD"
    SATA_SSD = "SATA_SSD"
    SAS_HDD = "SAS_HDD"
    SAS_SSD = "SAS_SSD"
    NVMe = "NVMe"


class Device(Base):
    __tablename__ = "devices"

    id: Mapped[int] = mapped_column(primary_key=True)
    host_id: Mapped[int] = mapped_column(ForeignKey("hosts.id"), nullable=False)
    serial_number: Mapped[str] = mapped_column(String(255), unique=True, nullable=False)
    model_number: Mapped[str | None] = mapped_column(String(255), nullable=True)
    firmware_revision: Mapped[str | None] = mapped_column(String(255), nullable=True)
    world_wide_name: Mapped[str | None] = mapped_column(String(255), nullable=True)
    device_type: Mapped[DeviceType | None] = mapped_column(
        Enum(DeviceType), nullable=True
    )
    device_path: Mapped[str | None] = mapped_column(String(255), nullable=True)
    form_factor_inches: Mapped[float | None] = mapped_column(Float, nullable=True)
    rotation_rate_rpm: Mapped[int | None] = mapped_column(Integer, nullable=True)
    is_ssd: Mapped[bool | None] = mapped_column(Boolean, nullable=True)
    logical_sector_size_bytes: Mapped[int | None] = mapped_column(Integer, nullable=True)
    physical_sector_size_bytes: Mapped[int | None] = mapped_column(Integer, nullable=True)
    capacity_bytes: Mapped[int | None] = mapped_column(BigInteger, nullable=True)
    max_lba: Mapped[int | None] = mapped_column(BigInteger, nullable=True)
    protocol: Mapped[str | None] = mapped_column(String(50), nullable=True)
    max_speed_gbps: Mapped[float | None] = mapped_column(Float, nullable=True)
    negotiated_speed_gbps: Mapped[float | None] = mapped_column(Float, nullable=True)
    encryption_support: Mapped[str | None] = mapped_column(String(255), nullable=True)
    ata_security: Mapped[str | None] = mapped_column(String(255), nullable=True)
    first_seen: Mapped[datetime] = mapped_column(
        DateTime, default=lambda: datetime.now(timezone.utc)
    )
    last_seen: Mapped[datetime] = mapped_column(
        DateTime, default=lambda: datetime.now(timezone.utc)
    )

    host = relationship("Host", back_populates="devices")
    health_snapshots = relationship("HealthSnapshot", back_populates="device")
    fired_alerts = relationship("FiredAlert", back_populates="device")
