"""Tests for the POST /api/v1/hosts/{host}/collect endpoint."""

import copy

from .conftest import SAMPLE_PAYLOAD


def test_collect_creates_host_and_devices(client):
    """POST a valid payload, then verify host + devices are created."""
    resp = client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "ok"
    assert data["host"] == "server-01"
    assert data["devices_upserted"] == 2
    assert data["snapshots_created"] == 2


def test_collect_upserts_on_second_call(client):
    """Second POST should update existing records, not duplicate."""
    client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)
    resp = client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)
    assert resp.status_code == 200

    # Should still be 2 devices, but 4 total snapshots (2 per collection)
    devices_resp = client.get("/api/v1/devices")
    assert devices_resp.json()["total"] == 2

    history = client.get("/api/v1/devices/ZC20B1TN/health/history")
    assert history.json()["total"] == 2


def test_collect_end_to_end_queryable(client):
    """After ingestion, devices and snapshots are queryable via GET."""
    client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)

    # Query device
    resp = client.get("/api/v1/devices/ZC20B1TN")
    assert resp.status_code == 200
    device = resp.json()
    assert device["serial_number"] == "ZC20B1TN"
    assert device["model_number"] == "ST4000NM000A"
    assert device["firmware_revision"] == "SN03"
    assert device["device_type"] == "SATA_HDD"
    assert device["capacity_bytes"] == 4000787030016
    assert device["hostname"] == "server-01"

    # Query health history
    resp = client.get("/api/v1/devices/ZC20B1TN/health/history")
    assert resp.status_code == 200
    history = resp.json()
    assert history["total"] == 1
    snap = history["snapshots"][0]
    assert snap["temperature_celsius"] == 35
    assert snap["power_on_hours"] == 8760.5
    assert snap["smart_status"] == "Good"
    assert snap["smart_tripped"] is False


def test_collect_minimal_device(client):
    """Ingest a device with only required fields."""
    payload = {
        "schema_version": "1.0.0",
        "host": {"hostname": "minimal-host"},
        "collected_at": "2025-06-01T12:00:00Z",
        "devices": [
            {
                "device_path": "/dev/sg5",
                "device_type": "NVMe",
                "identity": {
                    "model_number": "NVMe Test",
                    "serial_number": "NVME001",
                    "firmware_revision": "1.0",
                },
            }
        ],
    }
    resp = client.post("/api/v1/hosts/minimal-host/collect", json=payload)
    assert resp.status_code == 200
    assert resp.json()["devices_upserted"] == 1

    device = client.get("/api/v1/devices/NVME001").json()
    assert device["device_type"] == "NVMe"


def test_collect_updates_firmware_revision(client):
    """Firmware revision should be updated on subsequent collection."""
    client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)

    updated_payload = copy.deepcopy(SAMPLE_PAYLOAD)
    updated_payload["devices"][0]["identity"]["firmware_revision"] = "SN04"
    client.post("/api/v1/hosts/server-01/collect", json=updated_payload)

    device = client.get("/api/v1/devices/ZC20B1TN").json()
    assert device["firmware_revision"] == "SN04"
