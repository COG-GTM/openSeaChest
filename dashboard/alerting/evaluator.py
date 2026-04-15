"""Main evaluation engine – loads rules, evaluates against health data."""

from __future__ import annotations

import logging
from typing import Any, Dict, List, Optional, Type

from sqlalchemy.orm import Session

from dashboard.alerting.deduplication import is_duplicate, record_fired_alert
from dashboard.alerting.models import AlertRule, FiredAlert, RuleType
from dashboard.alerting.notifications.webhook import send_webhook
from dashboard.alerting.rules.base import BaseRuleEvaluator, EvaluationResult
from dashboard.alerting.rules.firmware import FirmwareRuleEvaluator
from dashboard.alerting.rules.poh import POHRuleEvaluator
from dashboard.alerting.rules.smart_tripped import SMARTTrippedRuleEvaluator
from dashboard.alerting.rules.temperature import TemperatureRuleEvaluator

logger = logging.getLogger(__name__)

RULE_EVALUATORS: Dict[RuleType, Type[BaseRuleEvaluator]] = {
    RuleType.smart_tripped: SMARTTrippedRuleEvaluator,
    RuleType.temperature: TemperatureRuleEvaluator,
    RuleType.firmware: FirmwareRuleEvaluator,
    RuleType.poh: POHRuleEvaluator,
}


def _get_evaluator(rule_type: RuleType) -> Optional[BaseRuleEvaluator]:
    cls = RULE_EVALUATORS.get(rule_type)
    if cls is None:
        logger.warning("No evaluator registered for rule type %s", rule_type)
        return None
    return cls()


def evaluate_device(
    session: Session,
    rule: AlertRule,
    device: Dict[str, Any],
    host: str,
    webhook_url: str = "",
    dedup_window_seconds: int = 86400,
) -> Optional[FiredAlert]:
    """Evaluate a single *rule* against a single *device*.

    Returns the ``FiredAlert`` row when the rule triggers (and is not
    deduplicated), otherwise ``None``.
    """
    evaluator = _get_evaluator(rule.rule_type)
    if evaluator is None:
        return None

    result: EvaluationResult = evaluator.evaluate(rule, device, host)
    if not result.triggered:
        return None

    # Deduplication check
    if is_duplicate(
        session,
        rule_id=rule.id,
        device_serial=result.device_serial,
        window_seconds=dedup_window_seconds,
    ):
        logger.info(
            "Duplicate alert suppressed: rule=%s device=%s",
            rule.name,
            result.device_serial,
        )
        return None

    fired = record_fired_alert(
        session,
        rule_id=rule.id,
        device_serial=result.device_serial,
        device_model=result.device_model,
        host=result.host,
        message=result.message,
    )

    # Notification
    notification_sent = False
    if webhook_url:
        notification_sent = send_webhook(
            url=webhook_url,
            alert_type=rule.rule_type.value,
            device_serial=result.device_serial,
            device_model=result.device_model,
            host=result.host,
            message=result.message,
            fired_at=fired.fired_at,
            rule_name=rule.name,
        )

    fired.notification_sent = notification_sent
    session.commit()
    return fired


def evaluate_payload(
    session: Session,
    payload: Dict[str, Any],
    webhook_url: str = "",
    dedup_window_seconds: int = 86400,
) -> List[FiredAlert]:
    """Evaluate all enabled rules against every device in an ingestion *payload*.

    This is the primary entry-point called by ``ingestion_hook`` and by the
    periodic sweep worker.
    """
    host = payload.get("host", {}).get("hostname", "unknown")
    devices: List[Dict[str, Any]] = payload.get("devices", [])

    rules: List[AlertRule] = (
        session.query(AlertRule).filter(AlertRule.enabled.is_(True)).all()
    )

    fired: List[FiredAlert] = []
    for device in devices:
        for rule in rules:
            result = evaluate_device(
                session,
                rule,
                device,
                host,
                webhook_url=webhook_url,
                dedup_window_seconds=dedup_window_seconds,
            )
            if result is not None:
                fired.append(result)

    return fired
