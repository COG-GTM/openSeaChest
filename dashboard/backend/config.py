"""Application configuration using pydantic-settings."""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    database_url: str = "sqlite:///./dashboard.db"
    api_v1_prefix: str = "/api/v1"
    debug: bool = False

    model_config = {"env_prefix": "DASHBOARD_"}


settings = Settings()
