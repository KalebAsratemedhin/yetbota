"""Surreal client stub — connect / close only; no graph writes yet."""

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


class SurrealClient:
    """Lazy placeholder. Wire `surrealdb` AsyncSurreal when implementing similarity edges."""

    def __init__(self, config: SurrealConfig | None = None) -> None:
        self.config = config or SurrealConfig.from_settings()
        self._connected = False

    async def connect(self) -> None:
        # Intentionally no-op in scaffold — avoids requiring a live Surreal on import/boot.
        self._connected = True

    async def close(self) -> None:
        self._connected = False

    @property
    def connected(self) -> bool:
        return self._connected
