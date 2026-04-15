"""Webhook notification channel."""

from __future__ import annotations

import json
import logging
from datetime import datetime
from typing import Optional

import requests

logger = logging.getLogger(__name__)


def send_webhook(
    url: str,
    alert_type: str,
    device_serial: str,
    device_model: str,
    host: str,
    message: str,
    fired_at: datetime,
    rule_name: str,
) -> bool:
    """POST an alert payload to the configured webhook URL.

    Returns ``True`` when the request succeeds (2xx), ``False`` otherwise.
    """
    if not url:
        logger.warning("Webhook URL is not configured; skipping notification.")
        return False

    payload = {
        "alert_type": alert_type,
        "device_serial": device_serial,
        "device_model": device_model,
        "host": host,
        "message": message,
        "fired_at": fired_at.isoformat() + ("Z" if fired_at.tzinfo else ""),
        "rule_name": rule_name,
    }

    try:
        response = requests.post(
            url,
            data=json.dumps(payload),
            headers={"Content-Type": "application/json"},
            timeout=10,
        )
        response.raise_for_status()
        logger.info("Webhook delivered successfully to %s", url)
        return True
    except requests.RequestException:
        logger.exception("Failed to deliver webhook to %s", url)
        return False
