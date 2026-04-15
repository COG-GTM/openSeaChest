"""Tests for firmware compliance endpoint."""

from .conftest import SAMPLE_PAYLOAD


def _seed(client):
    client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)


def test_firmware_compliance_all_compliant(client):
    """When target matches all devices, all should be compliant."""
    # Seed a single device with known firmware
    payload = {
        "schema_version": "1.0.0",
        "host": {"hostname": "fw-host"},
        "collected_at": "2025-06-01T12:00:00Z",
        "devices": [
            {
                "device_path": "/dev/sg0",
                "device_type": "SATA_HDD",
                "identity": {
                    "model_number": "TestModel",
                    "serial_number": "FW001",
                    "firmware_revision": "MATCH",
                },
            }
        ],
    }
    client.post("/api/v1/hosts/fw-host/collect", json=payload)

    resp = client.get("/api/v1/firmware/compliance?target_firmware=MATCH")
    assert resp.status_code == 200
    data = resp.json()
    assert data["compliant_count"] == 1
    assert data["non_compliant_count"] == 0
    assert data["compliant"][0]["serial_number"] == "FW001"


def test_firmware_compliance_mixed(client):
    """Sample payload has two different firmware revisions."""
    _seed(client)
    resp = client.get("/api/v1/firmware/compliance?target_firmware=SN03")
    data = resp.json()
    assert data["total_devices"] == 2
    assert data["compliant_count"] == 1
    assert data["non_compliant_count"] == 1
    assert data["compliant"][0]["firmware_revision"] == "SN03"
    assert data["non_compliant"][0]["firmware_revision"] == "SVT01B6Q"


def test_firmware_compliance_none_compliant(client):
    _seed(client)
    resp = client.get("/api/v1/firmware/compliance?target_firmware=NOPE")
    data = resp.json()
    assert data["compliant_count"] == 0
    assert data["non_compliant_count"] == 2


def test_firmware_compliance_requires_target(client):
    """target_firmware is a required query param."""
    resp = client.get("/api/v1/firmware/compliance")
    assert resp.status_code == 422
