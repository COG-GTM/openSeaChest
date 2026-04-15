"""Unit tests for the evaluation engine."""

from __future__ import annotations

from unittest.mock import patch

from dashboard.alerting.evaluator import evaluate_device, evaluate_payload
from dashboard.alerting.models import AlertRule, FiredAlert, RuleType


class TestEvaluateDevice:
    def test_returns_fired_alert_on_trigger(
        self, db_session, smart_rule, sample_device_bad_smart
    ):
        fired = evaluate_device(
            db_session, smart_rule, sample_device_bad_smart, "server-01"
        )
        assert fired is not None
        assert fired.device_serial == "ZQ3034X7R"
        assert fired.rule_id == smart_rule.id

    def test_returns_none_when_not_triggered(
        self, db_session, smart_rule, sample_device_ok
    ):
        fired = evaluate_device(
            db_session, smart_rule, sample_device_ok, "server-01"
        )
        assert fired is None

    def test_sends_webhook_when_url_provided(
        self, db_session, smart_rule, sample_device_bad_smart
    ):
        with patch(
            "dashboard.alerting.evaluator.send_webhook", return_value=True
        ) as mock_wh:
            fired = evaluate_device(
                db_session,
                smart_rule,
                sample_device_bad_smart,
                "server-01",
                webhook_url="https://example.com/hook",
            )
            assert fired is not None
            assert fired.notification_sent is True
            mock_wh.assert_called_once()

    def test_dedup_suppresses_second_alert(
        self, db_session, smart_rule, sample_device_bad_smart
    ):
        first = evaluate_device(
            db_session, smart_rule, sample_device_bad_smart, "server-01"
        )
        assert first is not None

        second = evaluate_device(
            db_session, smart_rule, sample_device_bad_smart, "server-01"
        )
        assert second is None


class TestEvaluatePayload:
    def test_evaluates_all_rules(self, db_session, smart_rule, sample_payload_bad_smart):
        fired = evaluate_payload(db_session, sample_payload_bad_smart)
        assert len(fired) == 1
        assert fired[0].device_serial == "ZQ3034X7R"

    def test_no_alerts_for_healthy_payload(self, db_session, smart_rule, sample_payload):
        fired = evaluate_payload(db_session, sample_payload)
        assert len(fired) == 0

    def test_multiple_rules_multiple_devices(self, db_session, smart_rule, temperature_rule):
        payload = {
            "schema_version": "1.0.0",
            "host": {"hostname": "host-1"},
            "collected_at": "2024-01-15T10:30:00Z",
            "devices": [
                {
                    "identity": {"serial_number": "A1", "model_number": "M1", "firmware_revision": "F1"},
                    "smart": {"status": "Bad", "tripped": True},
                    "temperature": {"current_celsius": 60},
                },
                {
                    "identity": {"serial_number": "B2", "model_number": "M2", "firmware_revision": "F2"},
                    "smart": {"status": "Good", "tripped": False},
                    "temperature": {"current_celsius": 30},
                },
            ],
        }
        fired = evaluate_payload(db_session, payload)
        # Device A1 triggers both rules, device B2 triggers none
        assert len(fired) == 2
        serials = {f.device_serial for f in fired}
        assert serials == {"A1"}
