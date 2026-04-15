"""Unit tests for the deduplication logic."""

from __future__ import annotations

from datetime import datetime, timedelta, timezone

from dashboard.alerting.deduplication import is_duplicate, record_fired_alert
from dashboard.alerting.models import FiredAlert


class TestIsDuplicate:
    def test_no_duplicate_when_empty(self, db_session, smart_rule):
        assert is_duplicate(db_session, smart_rule.id, "SN1") is False

    def test_duplicate_within_window(self, db_session, smart_rule):
        record_fired_alert(
            db_session,
            rule_id=smart_rule.id,
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="test",
        )
        db_session.commit()
        assert is_duplicate(db_session, smart_rule.id, "SN1") is True

    def test_not_duplicate_outside_window(self, db_session, smart_rule):
        alert = FiredAlert(
            rule_id=smart_rule.id,
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="old alert",
            fired_at=datetime.now(timezone.utc) - timedelta(hours=25),
        )
        db_session.add(alert)
        db_session.commit()
        assert is_duplicate(db_session, smart_rule.id, "SN1", window_seconds=86400) is False

    def test_resolved_alert_is_not_duplicate(self, db_session, smart_rule):
        alert = FiredAlert(
            rule_id=smart_rule.id,
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="resolved alert",
            resolved=True,
            resolved_at=datetime.now(timezone.utc),
        )
        db_session.add(alert)
        db_session.commit()
        assert is_duplicate(db_session, smart_rule.id, "SN1") is False

    def test_different_device_is_not_duplicate(self, db_session, smart_rule):
        record_fired_alert(
            db_session,
            rule_id=smart_rule.id,
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="test",
        )
        db_session.commit()
        assert is_duplicate(db_session, smart_rule.id, "SN2") is False

    def test_different_rule_is_not_duplicate(self, db_session, smart_rule, temperature_rule):
        record_fired_alert(
            db_session,
            rule_id=smart_rule.id,
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="test",
        )
        db_session.commit()
        assert is_duplicate(db_session, temperature_rule.id, "SN1") is False


class TestRecordFiredAlert:
    def test_creates_alert_row(self, db_session, smart_rule):
        alert = record_fired_alert(
            db_session,
            rule_id=smart_rule.id,
            device_serial="SN1",
            device_model="M1",
            host="host-1",
            message="test message",
        )
        db_session.commit()
        assert alert.id is not None
        assert alert.resolved is False
        assert alert.notification_sent is False

        stored = db_session.query(FiredAlert).filter_by(id=alert.id).one()
        assert stored.message == "test message"
