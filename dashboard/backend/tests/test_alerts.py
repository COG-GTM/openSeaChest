"""Tests for alert rules CRUD and fired alerts listing."""


def test_create_alert_rule(client):
    resp = client.post(
        "/api/v1/alerts/rules",
        json={
            "name": "SMART Tripped",
            "description": "Alert when SMART trips",
            "rule_type": "smart_tripped",
            "enabled": True,
        },
    )
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "SMART Tripped"
    assert data["rule_type"] == "smart_tripped"
    assert data["enabled"] is True
    assert "id" in data


def test_create_alert_rule_with_threshold(client):
    resp = client.post(
        "/api/v1/alerts/rules",
        json={
            "name": "High Temp",
            "rule_type": "temperature_threshold",
            "threshold_value": 60.0,
        },
    )
    assert resp.status_code == 201
    assert resp.json()["threshold_value"] == 60.0


def test_create_alert_rule_invalid_type(client):
    resp = client.post(
        "/api/v1/alerts/rules",
        json={"name": "Bad Rule", "rule_type": "invalid_type"},
    )
    assert resp.status_code == 422


def test_list_alert_rules(client):
    client.post(
        "/api/v1/alerts/rules",
        json={"name": "Rule 1", "rule_type": "smart_tripped"},
    )
    client.post(
        "/api/v1/alerts/rules",
        json={"name": "Rule 2", "rule_type": "poh_threshold", "threshold_value": 50000},
    )
    resp = client.get("/api/v1/alerts/rules")
    assert resp.status_code == 200
    assert len(resp.json()) == 2


def test_get_alert_rule_by_id(client):
    create_resp = client.post(
        "/api/v1/alerts/rules",
        json={"name": "Test Rule", "rule_type": "smart_tripped"},
    )
    rule_id = create_resp.json()["id"]

    resp = client.get(f"/api/v1/alerts/rules/{rule_id}")
    assert resp.status_code == 200
    assert resp.json()["name"] == "Test Rule"


def test_get_alert_rule_not_found(client):
    resp = client.get("/api/v1/alerts/rules/9999")
    assert resp.status_code == 404


def test_update_alert_rule(client):
    create_resp = client.post(
        "/api/v1/alerts/rules",
        json={"name": "Original", "rule_type": "smart_tripped"},
    )
    rule_id = create_resp.json()["id"]

    resp = client.put(
        f"/api/v1/alerts/rules/{rule_id}",
        json={"name": "Updated", "enabled": False},
    )
    assert resp.status_code == 200
    assert resp.json()["name"] == "Updated"
    assert resp.json()["enabled"] is False


def test_update_alert_rule_not_found(client):
    resp = client.put(
        "/api/v1/alerts/rules/9999",
        json={"name": "Nope"},
    )
    assert resp.status_code == 404


def test_delete_alert_rule(client):
    create_resp = client.post(
        "/api/v1/alerts/rules",
        json={"name": "To Delete", "rule_type": "smart_tripped"},
    )
    rule_id = create_resp.json()["id"]

    resp = client.delete(f"/api/v1/alerts/rules/{rule_id}")
    assert resp.status_code == 204

    resp = client.get(f"/api/v1/alerts/rules/{rule_id}")
    assert resp.status_code == 404


def test_delete_alert_rule_not_found(client):
    resp = client.delete("/api/v1/alerts/rules/9999")
    assert resp.status_code == 404


def test_list_fired_alerts_empty(client):
    resp = client.get("/api/v1/alerts")
    assert resp.status_code == 200
    data = resp.json()
    assert data["total"] == 0
    assert data["alerts"] == []


def test_list_fired_alerts_with_filters(client):
    """Test that filter params are accepted (even if no alerts exist)."""
    resp = client.get("/api/v1/alerts?device_serial=ZC20B1TN&resolved=false")
    assert resp.status_code == 200
    assert resp.json()["total"] == 0


def test_create_rule_firmware_approved_list(client):
    resp = client.post(
        "/api/v1/alerts/rules",
        json={
            "name": "FW Check",
            "rule_type": "firmware_not_approved",
            "approved_firmware_list": ["SN03", "SN04"],
        },
    )
    assert resp.status_code == 201
    data = resp.json()
    assert data["approved_firmware_list"] == ["SN03", "SN04"]
