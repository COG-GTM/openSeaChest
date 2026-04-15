"""FastAPI application entrypoint."""

from fastapi import FastAPI

from .config import settings
from .database import Base, engine
from .routers import collect, devices, firmware, alerts

# Create all tables (for development; production should use Alembic)
Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="OpenSeaChest Fleet Dashboard API",
    description="Backend API for the OpenSeaChest Fleet Dashboard — monitors storage device health, firmware compliance, and alerts.",
    version="0.1.0",
)

app.include_router(collect.router, prefix=settings.api_v1_prefix, tags=["ingestion"])
app.include_router(devices.router, prefix=settings.api_v1_prefix, tags=["devices"])
app.include_router(firmware.router, prefix=settings.api_v1_prefix, tags=["firmware"])
app.include_router(alerts.router, prefix=settings.api_v1_prefix, tags=["alerts"])


@app.get("/health", tags=["health"])
def health_check():
    return {"status": "ok"}
