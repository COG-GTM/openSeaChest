"""Integration test: ingest SMART-tripped payload -> verify alert fires."""

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

from dashboard.alerting.ingestion_hook import on_ingest
from dashboard.alerting.models import AlertRule, FiredAlert, RuleType


class TestIngestionIntegration:
    """End-to-end: set up a SMART tripped rule, ingest a payload with a
    tripped device, and verify that a ``FiredAlert`` is created and the
    webhook is called with the correct payload.
    """

    def test_smart_tripped_integration(self, db_session):
        # 1. Set up a SMART tripped alert rule
        rule = AlertRule(
            name="SMART Trip Alert",
            description="Fires when SMART is tripped",
            rule_type=RuleType.smart_tripped,
            enabled=True,
        )
        db_session.add(rule)
        db_session.commit()

        # 2. Build an ingestion payload where one device has smart.tripped = true
        payload = {
            "schema_version": "1.0.0",
            "host": {
                "hostname": "server-01",
                "os": "Linux 6.1",
                "agent_version": "0.1.0",
            },
            "collected_at": "2024-01-15T10:30:00Z",
            "devices": [
                {
                    "device_path": "/dev/sg1",
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
                },
                {
                    "device_path": "/dev/sg2",
                    "device_type": "SATA_SSD",
                    "identity": {
                        "model_number": "Samsung860",
                        "serial_number": "SAM001",
                        "firmware_revision": "RVT04B6Q",
                    },
                    "temperature": {"current_celsius": 30},
                    "power_on": {"power_on_hours": 5000},
                    "smart": {"status": "Good", "tripped": False},
                    "firmware": {"revision": "RVT04B6Q"},
                },
            ],
        }

        # 3. Ingest with a mocked webhook
        with patch(
            "dashboard.alerting.notifications.webhook.requests.post"
        ) as mock_post:
            mock_response = MagicMock()
            mock_response.raise_for_status = MagicMock()
            mock_post.return_value = mock_response

            fired = on_ingest(
                db_session,
                payload,
                webhook_url="https://hooks.example.com/alert",
                dedup_window_seconds=86400,
            )

        # 4. Verify that exactly one fired_alert record was created
        assert len(fired) == 1
        alert = fired[0]
        assert alert.device_serial == "ZQ3034X7R"
        assert alert.device_model == "ST4000DX001-1CE168"
        assert alert.host == "server-01"
        assert "Bad" in alert.message

        # Verify persistence
        stored = db_session.query(FiredAlert).all()
        assert len(stored) == 1
        assert stored[0].rule_id == rule.id

        # 5. Verify the webhook was called with correct payload
        mock_post.assert_called_once()
        call_kwargs = mock_post.call_args
        sent_payload = json.loads(
            call_kwargs.kwargs.get("data") or call_kwargs[1]["data"]
        )
        assert sent_payload["alert_type"] == "smart_tripped"
        assert sent_payload["device_serial"] == "ZQ3034X7R"
        assert sent_payload["device_model"] == "ST4000DX001-1CE168"
        assert sent_payload["host"] == "server-01"
        assert sent_payload["rule_name"] == "SMART Trip Alert"

    def test_dedup_prevents_duplicate_on_reingest(self, db_session):
        """Ingest the same tripped payload twice – only one alert should fire."""
        rule = AlertRule(
            name="SMART Trip Alert",
            rule_type=RuleType.smart_tripped,
            enabled=True,
        )
        db_session.add(rule)
        db_session.commit()

        payload = {
            "schema_version": "1.0.0",
            "host": {"hostname": "server-01"},
            "collected_at": "2024-01-15T10:30:00Z",
            "devices": [
                {
                    "device_path": "/dev/sg1",
                    "device_type": "SATA_HDD",
                    "identity": {
                        "model_number": "M1",
                        "serial_number": "SN1",
                        "firmware_revision": "F1",
                    },
                    "smart": {"status": "Bad", "tripped": True},
                },
            ],
        }

        with patch(
            "dashboard.alerting.notifications.webhook.requests.post"
        ) as mock_post:
            mock_response = MagicMock()
            mock_response.raise_for_status = MagicMock()
            mock_post.return_value = mock_response

            first = on_ingest(db_session, payload, webhook_url="https://example.com/hook")
            second = on_ingest(db_session, payload, webhook_url="https://example.com/hook")

        assert len(first) == 1
        assert len(second) == 0  # deduplicated

        total_alerts = db_session.query(FiredAlert).count()
        assert total_alerts == 1
