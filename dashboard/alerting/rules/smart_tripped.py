"""SMART Tripped rule evaluator."""

from __future__ import annotations

from typing import Any, Dict

from dashboard.alerting.models import AlertRule
from dashboard.alerting.rules.base import BaseRuleEvaluator, EvaluationResult


class SMARTTrippedRuleEvaluator(BaseRuleEvaluator):
    """Fires when ``smart.tripped == true`` (i.e. ``smart.status == 'Bad'``)."""

    def evaluate(
        self,
        rule: AlertRule,
        device: Dict[str, Any],
        host: str,
    ) -> EvaluationResult:
        serial = device.get("identity", {}).get("serial_number", "unknown")
        model = device.get("identity", {}).get("model_number", "unknown")
        smart = device.get("smart", {})
        tripped = smart.get("tripped", False)

        if tripped:
            return EvaluationResult(
                triggered=True,
                message=f"SMART health status is Bad for device {serial}",
                device_serial=serial,
                device_model=model,
                host=host,
            )

        return EvaluationResult(triggered=False)
