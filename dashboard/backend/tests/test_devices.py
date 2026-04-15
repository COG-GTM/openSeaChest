"""Tests for device listing and health history endpoints."""

from .conftest import SAMPLE_PAYLOAD


def _seed(client):
    client.post("/api/v1/hosts/server-01/collect", json=SAMPLE_PAYLOAD)


def test_list_devices(client):
    _seed(client)
    resp = client.get("/api/v1/devices")
    assert resp.status_code == 200
    data = resp.json()
    assert data["total"] == 2
    assert len(data["devices"]) == 2


def test_list_devices_pagination(client):
    _seed(client)
    resp = client.get("/api/v1/devices?limit=1&offset=0")
    data = resp.json()
    assert data["limit"] == 1
    assert len(data["devices"]) == 1

    resp2 = client.get("/api/v1/devices?limit=1&offset=1")
    data2 = resp2.json()
    assert len(data2["devices"]) == 1
    assert data2["devices"][0]["serial_number"] != data["devices"][0]["serial_number"]


def test_list_devices_filter_by_type(client):
    _seed(client)
    resp = client.get("/api/v1/devices?device_type=SATA_SSD")
    data = resp.json()
    assert data["total"] == 1
    assert data["devices"][0]["device_type"] == "SATA_SSD"


def test_list_devices_filter_by_host(client):
    _seed(client)
    resp = client.get("/api/v1/devices?host=server-01")
    assert resp.json()["total"] == 2

    resp = client.get("/api/v1/devices?host=nonexistent")
    assert resp.json()["total"] == 0


def test_list_devices_filter_by_smart_status(client):
    _seed(client)
    resp = client.get("/api/v1/devices?smart_status=Good")
    data = resp.json()
    assert data["total"] == 2


def test_get_device_by_serial(client):
    _seed(client)
    resp = client.get("/api/v1/devices/ZC20B1TN")
    assert resp.status_code == 200
    assert resp.json()["serial_number"] == "ZC20B1TN"


def test_get_device_not_found(client):
    resp = client.get("/api/v1/devices/NONEXISTENT")
    assert resp.status_code == 404


def test_health_history(client):
    _seed(client)
    resp = client.get("/api/v1/devices/ZC20B1TN/health/history")
    assert resp.status_code == 200
    data = resp.json()
    assert data["serial_number"] == "ZC20B1TN"
    assert data["total"] == 1
    assert data["snapshots"][0]["temperature_celsius"] == 35


def test_health_history_since_filter(client):
    _seed(client)
    resp = client.get(
        "/api/v1/devices/ZC20B1TN/health/history?since=2026-01-01T00:00:00"
    )
    # Our sample is from 2025, so nothing should match
    assert resp.json()["total"] == 0


def test_health_history_device_not_found(client):
    resp = client.get("/api/v1/devices/NONEXISTENT/health/history")
    assert resp.status_code == 404
