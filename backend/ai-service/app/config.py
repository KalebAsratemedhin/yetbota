from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Runtime config — same Postgres/Surreal as other local services."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    database_url: str = (
        "postgresql+asyncpg://postgres:postgres@localhost:5433/yetbota"
    )

    surreal_url: str = "ws://localhost:8000"
    surreal_namespace: str = "yetbota"
    surreal_database: str = "main"
    surreal_username: str = "root"
    surreal_password: str = "root"

    http_host: str = "0.0.0.0"
    http_port: int = 8083

    rabbitmq_url: str = "amqp://yetbota:yetbota@localhost:5672/"

    # Reserved for future ingest / embeddings (unused in scaffold).
    embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    openai_api_key: str = ""


@lru_cache
def get_settings() -> Settings:
    return Settings()
