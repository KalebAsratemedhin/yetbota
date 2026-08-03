from functools import lru_cache
from typing import Literal
from urllib.parse import quote_plus

from pydantic import BaseModel, ConfigDict, Field, computed_field

from infrastructure.config.loader import load_config


class AppSettings(BaseModel):
    name: str = "ai-service"
    version: str = "0.0.1"
    debug: bool = False


class HttpSettings(BaseModel):
    host: str = "0.0.0.0"
    port: int = 8989


class CorsSettings(BaseModel):
    hosts: list[str] = Field(default_factory=list)
    allow_credentials: bool = False

    @computed_field  # type: ignore[prop-decorator]
    @property
    def allow_origins(self) -> list[str]:
        origins: list[str] = []
        for entry in self.hosts:
            value = entry.strip()
            if not value:
                continue
            if "://" in value:
                origins.append(value.rstrip("/"))
                continue
            if value.startswith("localhost") or value.startswith("127.0.0.1"):
                origins.append(f"http://{value}")
                continue
            origins.append(f"https://{value}")
        return origins


class WeaviateSettings(BaseModel):
    url: str
    api_key: str = ""
    class_name: str = "YetbotaRag"


class SurrealDBSettings(BaseModel):
    url: str
    username: str
    password: str
    namespace: str
    database: str


class SimilaritySettings(BaseModel):
    enabled: bool = True
    top_n: int = 10
    candidate_oversample: int = 5


class GeminiSettings(BaseModel):
    api_key: str
    llm_model: str = "gemini-2.5-flash"
    embedding_model: str = "gemini-embedding-001"
    embedding_dimensions: int = 1536
    timeout_s: int = 30
    max_tokens: int = 1024
    temperature: float = 0.2


class DedupSettings(BaseModel):
    distance_threshold: float = 0.15
    spatial_radius_m: float = 250.0


class RagSettings(BaseModel):
    top_k: int = 8
    min_similarity: float = 0.35


class ScreeningSettings(BaseModel):
    block_threshold: float = 0.7
    timeout_s: int = 5
    cache_size: int = 1024
    cache_ttl_s: int = 300


class ChunkerSettings(BaseModel):
    size: int = 512
    overlap: int = 64


class RabbitMQSettings(BaseModel):
    url: str
    ingest_queue: str = "ai.ingest"
    prefetch_count: int = 16
    max_delivery_attempts: int = 3


class SummarySettings(BaseModel):
    debounce_minutes: int = 15
    min_contributions: int = 2
    max_source_contributions: int = 20


class ContentServiceSettings(BaseModel):
    base_url: str = "http://localhost:9966"
    api_key: str = ""
    timeout_s: float = 15.0


class PostgresSettings(BaseModel):
    host: str
    port: str
    user: str
    password: str
    db: str
    sslmode: str = "require"
    min_pool_size: int = 1
    max_pool_size: int = 10

    @computed_field  # type: ignore[prop-decorator]
    @property
    def dsn(self) -> str:
        password = quote_plus(self.password)
        return (
            f"postgres://{self.user}:{password}@{self.host}:{self.port}/"
            f"{self.db}?sslmode={self.sslmode}"
        )


class LoggingSettings(BaseModel):
    level: Literal["DEBUG", "INFO", "WARNING", "ERROR"] = "INFO"
    format: Literal["json", "console"] = "json"


class Settings(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    app: AppSettings
    http: HttpSettings = Field(default_factory=HttpSettings)
    cors: CorsSettings
    weaviate: WeaviateSettings
    surrealdb: SurrealDBSettings
    gemini: GeminiSettings
    dedup: DedupSettings = Field(default_factory=DedupSettings)
    rag: RagSettings = Field(default_factory=RagSettings)
    similarity: SimilaritySettings = Field(default_factory=SimilaritySettings)
    screening: ScreeningSettings = Field(default_factory=ScreeningSettings)
    chunker: ChunkerSettings = Field(default_factory=ChunkerSettings)
    rabbitmq: RabbitMQSettings
    postgres: PostgresSettings = Field(validation_alias="database")
    logging: LoggingSettings = Field(default_factory=LoggingSettings)
    summary: SummarySettings = Field(default_factory=SummarySettings)
    content_service: ContentServiceSettings = Field(default_factory=ContentServiceSettings)


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    data = load_config()
    return Settings.model_validate(data)
