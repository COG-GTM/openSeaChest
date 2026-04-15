"""Unit tests for notification channels."""

from __future__ import annotations

import json
from datetime import datetime, timezone
from unittest.mock import MagicMock, patch

from dashboard.alerting.notifications.webhook import send_webhook


class TestWebhook:
    def test_sends_correct_payload(self):
        fired_at = datetime(2024, 1, 15, 10, 30, 0, tzinfo=timezone.utc)

        with patch("dashboard.alerting.notifications.webhook.requests.post") as mock_post:
            mock_response = MagicMock()
            mock_response.raise_for_status = MagicMock()
            mock_post.return_value = mock_response

            result = send_webhook(
                url="https://hooks.example.com/alert",
                alert_type="smart_tripped",
                device_serial="ZQ3034X7R",
                device_model="ST4000DX001-1CE168",
                host="server-01",
                message="SMART health status is Bad for device ZQ3034X7R",
                fired_at=fired_at,
                rule_name="SMART Trip Alert",
            )

            assert result is True
            mock_post.assert_called_once()
            call_kwargs = mock_post.call_args
            payload = json.loads(call_kwargs.kwargs.get("data") or call_kwargs[1]["data"])
            assert payload["alert_type"] == "smart_tripped"
            assert payload["device_serial"] == "ZQ3034X7R"
            assert payload["device_model"] == "ST4000DX001-1CE168"
            assert payload["host"] == "server-01"
            assert payload["rule_name"] == "SMART Trip Alert"
            assert "2024-01-15" in payload["fired_at"]

    def test_returns_false_on_http_error(self):
        import requests as req

        with patch("dashboard.alerting.notifications.webhook.requests.post") as mock_post:
            mock_post.side_effect = req.RequestException("connection failed")
            result = send_webhook(
                url="https://hooks.example.com/alert",
                alert_type="smart_tripped",
                device_serial="SN1",
                device_model="M1",
                host="host-1",
                message="test",
                fired_at=datetime.now(timezone.utc),
                rule_name="Rule 1",
            )
            assert result is False

    def test_returns_false_when_url_empty(self):
        result = send_webhook(
            url="",
            alert_type="smart_tripped",
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="test",
            fired_at=datetime.now(timezone.utc),
            rule_name="Rule 1",
        )
        assert result is False
