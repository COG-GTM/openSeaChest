"""Firmware-not-approved rule evaluator."""

from __future__ import annotations

import json
from typing import Any, Dict, List

from dashboard.alerting.models import AlertRule
from dashboard.alerting.rules.base import BaseRuleEvaluator, EvaluationResult


class FirmwareRuleEvaluator(BaseRuleEvaluator):
    """Fires when the device firmware revision is NOT in the approved list."""

    @staticmethod
    def _parse_approved_list(rule: AlertRule) -> List[str]:
        raw = rule.approved_firmware_list
        if not raw:
            return []
        try:
            parsed = json.loads(raw)
            if isinstance(parsed, list):
                return [str(v) for v in parsed]
        except (json.JSONDecodeError, TypeError):
            pass
        return []

    def evaluate(
        self,
        rule: AlertRule,
        device: Dict[str, Any],
        host: str,
    ) -> EvaluationResult:
        serial = device.get("identity", {}).get("serial_number", "unknown")
        model = device.get("identity", {}).get("model_number", "unknown")

        firmware_revision = (
            device.get("firmware", {}).get("revision")
            or device.get("identity", {}).get("firmware_revision")
        )

        approved = self._parse_approved_list(rule)
        if not approved:
            # No approved list configured – nothing to check.
            return EvaluationResult(triggered=False)

        if firmware_revision and firmware_revision not in approved:
            return EvaluationResult(
                triggered=True,
                message=(
                    f"Firmware revision '{firmware_revision}' is not in the approved "
                    f"list for device {serial}"
                ),
                device_serial=serial,
                device_model=model,
                host=host,
            )

        return EvaluationResult(triggered=False)
