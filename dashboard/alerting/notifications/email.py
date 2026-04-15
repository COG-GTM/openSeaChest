"""Email notification channel (optional)."""

from __future__ import annotations

import logging
import smtplib
from email.mime.text import MIMEText

from dashboard.alerting import config

logger = logging.getLogger(__name__)


def send_email(
    subject: str,
    body: str,
    smtp_host: str = "",
    smtp_port: int = 0,
    smtp_username: str = "",
    smtp_password: str = "",
    from_addr: str = "",
    to_addr: str = "",
) -> bool:
    """Send a plain-text alert e-mail via SMTP.

    Falls back to values from ``config`` when parameters are not provided.
    Returns ``True`` on success, ``False`` on failure.
    """
    host = smtp_host or config.SMTP_HOST
    port = smtp_port or config.SMTP_PORT
    username = smtp_username or config.SMTP_USERNAME
    password = smtp_password or config.SMTP_PASSWORD
    sender = from_addr or config.SMTP_FROM
    recipient = to_addr or config.SMTP_TO

    if not host or not recipient:
        logger.warning("Email not configured; skipping notification.")
        return False

    msg = MIMEText(body, "plain")
    msg["Subject"] = subject
    msg["From"] = sender
    msg["To"] = recipient

    try:
        with smtplib.SMTP(host, port, timeout=10) as server:
            server.ehlo()
            if port == 587:
                server.starttls()
                server.ehlo()
            if username and password:
                server.login(username, password)
            server.sendmail(sender, [recipient], msg.as_string())
        logger.info("Email sent to %s", recipient)
        return True
    except Exception:
        logger.exception("Failed to send email to %s", recipient)
        return False
