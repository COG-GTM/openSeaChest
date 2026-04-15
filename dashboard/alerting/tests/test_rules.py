"""Unit tests for each alert rule evaluator."""

from __future__ import annotations

import json

from dashboard.alerting.models import AlertRule, RuleType
from dashboard.alerting.rules.firmware import FirmwareRuleEvaluator
from dashboard.alerting.rules.poh import POHRuleEvaluator
from dashboard.alerting.rules.smart_tripped import SMARTTrippedRuleEvaluator
from dashboard.alerting.rules.temperature import TemperatureRuleEvaluator


# ---------------------------------------------------------------------------
# SMART Tripped
# ---------------------------------------------------------------------------

class TestSMARTTrippedRule:
    def test_triggers_when_tripped(self, smart_rule, sample_device_bad_smart):
        evaluator = SMARTTrippedRuleEvaluator()
        result = evaluator.evaluate(smart_rule, sample_device_bad_smart, "server-01")
        assert result.triggered is True
        assert "Bad" in result.message
        assert result.device_serial == "ZQ3034X7R"

    def test_does_not_trigger_when_healthy(self, smart_rule, sample_device_ok):
        evaluator = SMARTTrippedRuleEvaluator()
        result = evaluator.evaluate(smart_rule, sample_device_ok, "server-01")
        assert result.triggered is False

    def test_does_not_trigger_when_smart_missing(self, smart_rule):
        device = {
            "identity": {"serial_number": "AAA", "model_number": "M1", "firmware_revision": "F1"},
        }
        evaluator = SMARTTrippedRuleEvaluator()
        result = evaluator.evaluate(smart_rule, device, "server-01")
        assert result.triggered is False


# ---------------------------------------------------------------------------
# Temperature
# ---------------------------------------------------------------------------

class TestTemperatureRule:
    def test_triggers_above_threshold(self, temperature_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "temperature": {"current_celsius": 60},
        }
        evaluator = TemperatureRuleEvaluator()
        result = evaluator.evaluate(temperature_rule, device, "host-1")
        assert result.triggered is True
        assert "60" in result.message

    def test_does_not_trigger_below_threshold(self, temperature_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "temperature": {"current_celsius": 50},
        }
        evaluator = TemperatureRuleEvaluator()
        result = evaluator.evaluate(temperature_rule, device, "host-1")
        assert result.triggered is False

    def test_does_not_trigger_at_exact_threshold(self, temperature_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "temperature": {"current_celsius": 55},
        }
        evaluator = TemperatureRuleEvaluator()
        result = evaluator.evaluate(temperature_rule, device, "host-1")
        assert result.triggered is False

    def test_uses_default_threshold_when_none(self, db_session):
        rule = AlertRule(
            name="Temp no threshold",
            rule_type=RuleType.temperature,
            threshold_value=None,
            enabled=True,
        )
        db_session.add(rule)
        db_session.commit()
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "temperature": {"current_celsius": 56},
        }
        evaluator = TemperatureRuleEvaluator()
        result = evaluator.evaluate(rule, device, "host-1")
        assert result.triggered is True

    def test_handles_missing_temperature(self, temperature_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
        }
        evaluator = TemperatureRuleEvaluator()
        result = evaluator.evaluate(temperature_rule, device, "host-1")
        assert result.triggered is False


# ---------------------------------------------------------------------------
# Firmware Not Approved
# ---------------------------------------------------------------------------

class TestFirmwareRule:
    def test_triggers_when_not_approved(self, firmware_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "XX01"},
            "firmware": {"revision": "XX01"},
        }
        evaluator = FirmwareRuleEvaluator()
        result = evaluator.evaluate(firmware_rule, device, "host-1")
        assert result.triggered is True
        assert "XX01" in result.message

    def test_does_not_trigger_when_approved(self, firmware_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "CC49"},
            "firmware": {"revision": "CC49"},
        }
        evaluator = FirmwareRuleEvaluator()
        result = evaluator.evaluate(firmware_rule, device, "host-1")
        assert result.triggered is False

    def test_does_not_trigger_when_list_empty(self, db_session):
        rule = AlertRule(
            name="FW empty",
            rule_type=RuleType.firmware,
            approved_firmware_list=None,
            enabled=True,
        )
        db_session.add(rule)
        db_session.commit()
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "XX01"},
            "firmware": {"revision": "XX01"},
        }
        evaluator = FirmwareRuleEvaluator()
        result = evaluator.evaluate(rule, device, "host-1")
        assert result.triggered is False

    def test_falls_back_to_identity_firmware_revision(self, firmware_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "XX01"},
        }
        evaluator = FirmwareRuleEvaluator()
        result = evaluator.evaluate(firmware_rule, device, "host-1")
        assert result.triggered is True


# ---------------------------------------------------------------------------
# Power-On Hours
# ---------------------------------------------------------------------------

class TestPOHRule:
    def test_triggers_above_threshold(self, poh_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "power_on": {"power_on_hours": 50000},
        }
        evaluator = POHRuleEvaluator()
        result = evaluator.evaluate(poh_rule, device, "host-1")
        assert result.triggered is True

    def test_does_not_trigger_below_threshold(self, poh_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "power_on": {"power_on_hours": 30000},
        }
        evaluator = POHRuleEvaluator()
        result = evaluator.evaluate(poh_rule, device, "host-1")
        assert result.triggered is False

    def test_handles_missing_poh(self, poh_rule):
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
        }
        evaluator = POHRuleEvaluator()
        result = evaluator.evaluate(poh_rule, device, "host-1")
        assert result.triggered is False

    def test_uses_default_threshold(self, db_session):
        rule = AlertRule(
            name="POH default",
            rule_type=RuleType.poh,
            threshold_value=None,
            enabled=True,
        )
        db_session.add(rule)
        db_session.commit()
        device = {
            "identity": {"serial_number": "SN1", "model_number": "M1", "firmware_revision": "F1"},
            "power_on": {"power_on_hours": 41000},
        }
        evaluator = POHRuleEvaluator()
        result = evaluator.evaluate(rule, device, "host-1")
        assert result.triggered is True
