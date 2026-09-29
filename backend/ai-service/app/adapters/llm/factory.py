"""Build a ChatLLM adapter from settings (provider switch)."""

from __future__ import annotations

from app.adapters.llm.gemini_chat import GeminiChatClient
from app.adapters.llm.openai_chat import OpenAIChatClient
from app.config import Settings
from app.services.chat import ChatLLM


def build_chat_llm(settings: Settings) -> ChatLLM:
    provider = (settings.chat_llm_provider or "").strip().lower()
    if provider == "gemini":
        return GeminiChatClient(
            settings.gemini_api_key,
            model=settings.gemini_model,
            base_url=settings.gemini_base_url,
        )
    if provider == "openai":
        return OpenAIChatClient(
            settings.openai_api_key,
            model=settings.openai_model,
            base_url=settings.openai_base_url,
        )
    raise ValueError(
        f"unsupported CHAT_LLM_PROVIDER={settings.chat_llm_provider!r} "
        "(expected gemini or openai)"
    )
