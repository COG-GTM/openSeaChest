# OpenSeaChest Fleet Dashboard — Backend API

FastAPI backend for the OpenSeaChest Fleet Dashboard. Provides REST APIs for device ingestion, health monitoring, firmware compliance checking, and alert management.

## Quick Start

```bash
cd dashboard/backend
pip install -r requirements.txt
uvicorn dashboard.backend.main:app --reload
```

The API docs are available at:
- Swagger UI: http://localhost:8000/docs
- ReDoc: http://localhost:8000/redoc
- OpenAPI JSON: http://localhost:8000/openapi.json

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/hosts/{host}/collect` | Ingest device data from collection agent |
| GET | `/api/v1/devices` | List devices (filterable by type, host, smart_status) |
| GET | `/api/v1/devices/{serial}` | Get device by serial number |
| GET | `/api/v1/devices/{serial}/health/history` | Health snapshot time-series |
| GET | `/api/v1/firmware/compliance` | Firmware compliance report |
| POST | `/api/v1/alerts/rules` | Create alert rule |
| GET | `/api/v1/alerts/rules` | List alert rules |
| GET | `/api/v1/alerts/rules/{id}` | Get alert rule |
| PUT | `/api/v1/alerts/rules/{id}` | Update alert rule |
| DELETE | `/api/v1/alerts/rules/{id}` | Delete alert rule |
| GET | `/api/v1/alerts` | List fired alerts |

## Database

Uses SQLAlchemy with SQLite by default. Set `DASHBOARD_DATABASE_URL` environment variable to use a different database.

### Migrations (Alembic)

```bash
cd dashboard/backend
alembic revision --autogenerate -m "description"
alembic upgrade head
```

## Testing

```bash
cd /path/to/repo/root
python -m pytest dashboard/backend/tests/ -v
```

Tests use an in-memory SQLite database and FastAPI's TestClient.

## Ingestion Schema

The `POST /collect` endpoint accepts JSON conforming to the shared ingestion schema v1.0.0 defined in `docs/dashboard/ingestion-schema.json` on the `dashboard/shared-schema` branch.
