# Alerting & Compliance Engine (WI-4)

Background alerting and compliance evaluation engine for the **OpenSeaChest Fleet Dashboard**.

## Overview

This package provides:

- **Seeded alert rules** – SMART Tripped, Temperature Threshold, Firmware Not Approved, Power-On Hours Threshold.
- **Two evaluation modes** – event-driven (on ingestion) and periodic sweep.
- **Deduplication** – configurable time-window suppression of duplicate alerts.
- **Notification channels** – webhook (required) and email (optional).
- **Background worker** – APScheduler-based periodic sweep of the latest health snapshots.

## Package Structure

```
dashboard/alerting/
├── __init__.py
├── config.py               # Environment-based configuration
├── models.py               # SQLAlchemy models (alert_rules, fired_alerts, devices, health_snapshots)
├── rules/
│   ├── __init__.py
│   ├── base.py             # Base rule evaluator interface
│   ├── smart_tripped.py    # SMART tripped rule
│   ├── temperature.py      # Temperature threshold rule
│   ├── firmware.py         # Firmware not in approved list rule
│   └── poh.py              # Power-on hours threshold rule
├── evaluator.py            # Main evaluation engine
├── deduplication.py        # Deduplication logic
├── notifications/
│   ├── __init__.py
│   ├── webhook.py          # Webhook notification channel
│   └── email.py            # Email notification channel (optional)
├── worker.py               # Background worker (APScheduler)
├── ingestion_hook.py       # Event-driven evaluation on data ingestion
├── requirements.txt
├── README.md
└── tests/
    ├── __init__.py
    ├── conftest.py
    ├── test_rules.py
    ├── test_evaluator.py
    ├── test_deduplication.py
    ├── test_notifications.py
    └── test_integration.py
```

## Quick Start

### Install dependencies

```bash
pip install -r dashboard/alerting/requirements.txt
pip install pytest  # for running tests
```

### Configuration (environment variables)

| Variable | Default | Description |
|---|---|---|
| `ALERTING_DATABASE_URL` | `sqlite:///alerting.db` | SQLAlchemy database URL |
| `ALERTING_WEBHOOK_URL` | *(empty)* | Webhook endpoint for alert notifications |
| `ALERTING_DEDUP_WINDOW_SECONDS` | `86400` | Deduplication window (seconds) |
| `ALERTING_SWEEP_INTERVAL_SECONDS` | `300` | Periodic sweep interval (seconds) |
| `ALERTING_EMAIL_ENABLED` | `false` | Enable email notifications |
| `ALERTING_SMTP_HOST` | *(empty)* | SMTP server hostname |
| `ALERTING_SMTP_PORT` | `587` | SMTP server port |
| `ALERTING_SMTP_USERNAME` | *(empty)* | SMTP login username |
| `ALERTING_SMTP_PASSWORD` | *(empty)* | SMTP login password |
| `ALERTING_SMTP_FROM` | *(empty)* | Sender email address |
| `ALERTING_SMTP_TO` | *(empty)* | Recipient email address |

### Event-driven evaluation

```python
from dashboard.alerting.models import init_db
from dashboard.alerting.ingestion_hook import on_ingest

SessionFactory = init_db("sqlite:///alerting.db")
session = SessionFactory()

payload = { ... }  # ingestion payload matching v1.0.0 schema
fired = on_ingest(session, payload, webhook_url="https://hooks.example.com/alert")
session.close()
```

### Periodic sweep worker

```bash
python -m dashboard.alerting.worker
```

Or programmatically:

```python
from dashboard.alerting.worker import start_worker
start_worker()
```

### Running tests

```bash
cd <repo-root>
python -m pytest dashboard/alerting/tests/ -v
```

## Alert Rules

| Rule | Fires When | Default Threshold |
|---|---|---|
| SMART Tripped | `smart.tripped == true` | — |
| Temperature | `temperature.current_celsius > threshold` | 55 °C |
| Firmware Not Approved | firmware revision not in approved list | *(configured per rule)* |
| POH Threshold | `power_on.power_on_hours > threshold` | 40 000 hours |

## Shared Ingestion Schema

The engine consumes payloads conforming to the **v1.0.0 ingestion schema** defined in `docs/dashboard/ingestion-schema.json` on the `dashboard/shared-schema` branch.
