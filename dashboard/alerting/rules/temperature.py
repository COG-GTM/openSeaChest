"""Temperature threshold rule evaluator."""

from __future__ import annotations

from typing import Any, Dict

from dashboard.alerting.models import AlertRule
from dashboard.alerting.rules.base import BaseRuleEvaluator, EvaluationResult

DEFAULT_TEMPERATURE_THRESHOLD = 55.0


class TemperatureRuleEvaluator(BaseRuleEvaluator):
    """Fires when ``temperature.current_celsius`` exceeds the configured threshold."""

    def evaluate(
        self,
        rule: AlertRule,
        device: Dict[str, Any],
        host: str,
    ) -> EvaluationResult:
        serial = device.get("identity", {}).get("serial_number", "unknown")
        model = device.get("identity", {}).get("model_number", "unknown")
        temp_info = device.get("temperature", {})
        current = temp_info.get("current_celsius")

        threshold = (
            rule.threshold_value
            if rule.threshold_value is not None
            else DEFAULT_TEMPERATURE_THRESHOLD
        )

        if current is not None and current > threshold:
            return EvaluationResult(
                triggered=True,
                message=(
                    f"Temperature {current}°C exceeds threshold {threshold}°C "
                    f"for device {serial}"
                ),
                device_serial=serial,
                device_model=model,
                host=host,
            )

        return EvaluationResult(triggered=False)
