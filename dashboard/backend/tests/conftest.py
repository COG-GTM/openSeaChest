"""Pytest fixtures: test DB, test client."""

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from dashboard.backend.database import Base, get_db
from dashboard.backend.main import app


@pytest.fixture(name="db")
def db_session():
    """Create an in-memory SQLite database for each test."""
    engine = create_engine(
        "sqlite://",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    Base.metadata.create_all(bind=engine)
    TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    session = TestingSessionLocal()
    try:
        yield session
    finally:
        session.close()
        Base.metadata.drop_all(bind=engine)


@pytest.fixture(name="client")
def test_client(db):
    """FastAPI TestClient wired to the test database."""

    def _override_get_db():
        try:
            yield db
        finally:
            pass

    app.dependency_overrides[get_db] = _override_get_db
    with TestClient(app) as c:
        yield c
    app.dependency_overrides.clear()


SAMPLE_PAYLOAD = {
    "schema_version": "1.0.0",
    "host": {
        "hostname": "server-01",
        "os": "Linux 6.1",
        "agent_version": "0.1.0",
    },
    "collected_at": "2025-06-01T12:00:00Z",
    "devices": [
        {
            "device_path": "/dev/sg0",
            "device_type": "SATA_HDD",
            "identity": {
                "model_number": "ST4000NM000A",
                "serial_number": "ZC20B1TN",
                "firmware_revision": "SN03",
                "world_wide_name": "5000C500A1B2C3D4",
                "form_factor_inches": 3.5,
                "rotation_rate_rpm": 7200,
                "is_ssd": False,
                "logical_sector_size_bytes": 512,
                "physical_sector_size_bytes": 4096,
            },
            "capacity": {
                "capacity_bytes": 4000787030016,
                "capacity_display": "4.00 TB",
                "max_lba": 7814037167,
            },
            "temperature": {
                "current_celsius": 35,
                "highest_celsius": 42,
                "lowest_celsius": 20,
            },
            "power_on": {
                "power_on_hours": 8760.5,
                "power_on_hours_display": "365 days 0 hours",
            },
            "interface": {
                "protocol": "SATA",
                "max_speed_gbps": 6.0,
                "negotiated_speed_gbps": 6.0,
            },
            "workload": {
                "annualized_workload_rate_tb_yr": 150.5,
                "total_bytes_read": 82500000000000,
                "total_bytes_written": 67500000000000,
            },
            "smart": {
                "status": "Good",
                "tripped": False,
                "attributes": [
                    {
                        "id": 1,
                        "name": "Read Error Rate",
                        "nominal": 117,
                        "worst": 99,
                        "raw_value": 158433204,
                    },
                    {
                        "id": 5,
                        "name": "Reallocated Sectors Count",
                        "nominal": 100,
                        "worst": 100,
                        "raw_value": 0,
                    },
                ],
            },
            "firmware": {"revision": "SN03"},
            "security": {
                "encryption_support": "Not Supported",
                "ata_security": "Supported, Frozen",
            },
        },
        {
            "device_path": "/dev/sg1",
            "device_type": "SATA_SSD",
            "identity": {
                "model_number": "Samsung SSD 870 EVO",
                "serial_number": "S5XXNF0R123456",
                "firmware_revision": "SVT01B6Q",
                "is_ssd": True,
                "logical_sector_size_bytes": 512,
                "physical_sector_size_bytes": 512,
            },
            "capacity": {
                "capacity_bytes": 500107862016,
                "max_lba": 976773167,
            },
            "temperature": {
                "current_celsius": 28,
                "highest_celsius": 55,
                "lowest_celsius": 15,
            },
            "power_on": {"power_on_hours": 4380.0},
            "interface": {
                "protocol": "SATA",
                "max_speed_gbps": 6.0,
                "negotiated_speed_gbps": 6.0,
            },
            "workload": {
                "total_bytes_written": 25000000000000,
                "percentage_used_endurance": 2.0,
            },
            "smart": {"status": "Good", "tripped": False},
            "firmware": {"revision": "SVT01B6Q"},
            "security": {"encryption_support": "Not Supported"},
        },
    ],
}
