"""Surreal connection config (shared by store)."""

from dataclasses import dataclass

from app.config import Settings, get_settings


@dataclass
class SurrealConfig:
    url: str
    namespace: str
    database: str
    username: str
    password: str

    @classmethod
    def from_settings(cls, settings: Settings | None = None) -> "SurrealConfig":
        s = settings or get_settings()
        return cls(
            url=s.surreal_url,
            namespace=s.surreal_namespace,
            database=s.surreal_database,
            username=s.surreal_username,
            password=s.surreal_password,
        )
