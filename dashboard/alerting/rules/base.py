"""Base interface for alert-rule evaluators."""

from __future__ import annotations

import abc
from dataclasses import dataclass
from typing import Any, Dict, Optional

from dashboard.alerting.models import AlertRule


@dataclass
class EvaluationResult:
    """Outcome of evaluating a single rule against a single device."""

    triggered: bool
    message: str = ""
    device_serial: str = ""
    device_model: str = ""
    host: str = ""


class BaseRuleEvaluator(abc.ABC):
    """Abstract evaluator that each concrete rule type must implement."""

    @abc.abstractmethod
    def evaluate(
        self,
        rule: AlertRule,
        device: Dict[str, Any],
        host: str,
    ) -> EvaluationResult:
        """Evaluate *rule* against a single *device* dict from the ingestion payload.

        Parameters
        ----------
        rule:
            The persisted ``AlertRule`` definition (contains thresholds, etc.).
        device:
            A single device object from the ingestion payload.
        host:
            The hostname that reported this device.

        Returns
        -------
        EvaluationResult
            Whether the rule was triggered, along with contextual information.
        """
