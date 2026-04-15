"""Central configuration for the Alerting & Compliance Engine."""

from __future__ import annotations

import os


DATABASE_URL: str = os.getenv(
    "ALERTING_DATABASE_URL",
    "sqlite:///alerting.db",
)

WEBHOOK_URL: str = os.getenv("ALERTING_WEBHOOK_URL", "")

DEDUP_WINDOW_SECONDS: int = int(
    os.getenv("ALERTING_DEDUP_WINDOW_SECONDS", str(24 * 3600))
)

SWEEP_INTERVAL_SECONDS: int = int(
    os.getenv("ALERTING_SWEEP_INTERVAL_SECONDS", str(5 * 60))
)

# SMTP / e-mail (optional)
SMTP_HOST: str = os.getenv("ALERTING_SMTP_HOST", "")
SMTP_PORT: int = int(os.getenv("ALERTING_SMTP_PORT", "587"))
SMTP_USERNAME: str = os.getenv("ALERTING_SMTP_USERNAME", "")
SMTP_PASSWORD: str = os.getenv("ALERTING_SMTP_PASSWORD", "")
SMTP_FROM: str = os.getenv("ALERTING_SMTP_FROM", "")
SMTP_TO: str = os.getenv("ALERTING_SMTP_TO", "")
EMAIL_ENABLED: bool = os.getenv("ALERTING_EMAIL_ENABLED", "false").lower() == "true"
