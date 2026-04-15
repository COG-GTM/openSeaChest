"""Power-On Hours (POH) threshold rule evaluator."""

from __future__ import annotations

from typing import Any, Dict

from dashboard.alerting.models import AlertRule
from dashboard.alerting.rules.base import BaseRuleEvaluator, EvaluationResult

DEFAULT_POH_THRESHOLD = 40000.0


class POHRuleEvaluator(BaseRuleEvaluator):
    """Fires when ``power_on.power_on_hours`` exceeds the configured threshold."""

    def evaluate(
        self,
        rule: AlertRule,
        device: Dict[str, Any],
        host: str,
    ) -> EvaluationResult:
        serial = device.get("identity", {}).get("serial_number", "unknown")
        model = device.get("identity", {}).get("model_number", "unknown")
        poh = device.get("power_on", {}).get("power_on_hours")

        threshold = (
            rule.threshold_value
            if rule.threshold_value is not None
            else DEFAULT_POH_THRESHOLD
        )

        if poh is not None and poh > threshold:
            return EvaluationResult(
                triggered=True,
                message=(
                    f"Power-on hours {poh} exceeds threshold {threshold} "
                    f"for device {serial}"
                ),
                device_serial=serial,
                device_model=model,
                host=host,
            )

        return EvaluationResult(triggered=False)
