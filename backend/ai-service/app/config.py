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

    embedding_provider: str = "gemini"
    gemini_embedding_model: str = "gemini-embedding-001"
    embedding_dimensions: int = 768
    # Local fastembed (EMBEDDING_PROVIDER=local)
    embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    embedding_cache_dir: str = ""
    similarity_top_k: int = 10
    similarity_min_score: float = 0.55

    # Chat / RAG — switch with CHAT_LLM_PROVIDER=gemini|openai
    chat_llm_provider: str = "gemini"
    gemini_api_key: str = ""
    gemini_model: str = "gemini-2.5-flash"
    gemini_base_url: str = "https://generativelanguage.googleapis.com/v1beta"
    openai_api_key: str = ""
    openai_model: str = "gpt-4o-mini"
    openai_base_url: str = "https://api.openai.com/v1"
    chat_top_k: int = 8
    chat_min_score: float = 0.22
    chat_soft_min_score: float = 0.12

    # Place directory (content-service internal APIs)
    content_base_url: str = "http://localhost:8082"
    internal_service_token: str = "dev-internal-token"
    place_match_radius_m: float = 150.0
    place_embed_min: float = 0.78
    place_name_min: float = 0.55
    place_ambiguity_margin: float = 0.05
    place_summary_debounce_sec: float = 10.0


@lru_cache
def get_settings() -> Settings:
    return Settings()
