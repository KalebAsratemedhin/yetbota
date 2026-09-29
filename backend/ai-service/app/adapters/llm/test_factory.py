"""Unit tests for Gemini message mapping and LLM factory."""

from __future__ import annotations

import pytest

from app.adapters.llm.factory import build_chat_llm
from app.adapters.llm.gemini_chat import GeminiChatClient, _to_gemini_contents
from app.adapters.llm.openai_chat import OpenAIChatClient
from app.config import Settings


def test_to_gemini_contents_maps_roles_and_merges():
    contents = _to_gemini_contents(
        [
            {"role": "user", "content": "hi"},
            {"role": "assistant", "content": "hello"},
            {"role": "user", "content": "more"},
            {"role": "user", "content": "again"},
        ]
    )
    assert contents == [
        {"role": "user", "parts": [{"text": "hi"}]},
        {"role": "model", "parts": [{"text": "hello"}]},
        {"role": "user", "parts": [{"text": "more\n\nagain"}]},
    ]


def test_build_chat_llm_gemini_default():
    settings = Settings(
        chat_llm_provider="gemini",
        gemini_api_key="g-key",
        gemini_model="gemini-2.5-flash",
    )
    llm = build_chat_llm(settings)
    assert isinstance(llm, GeminiChatClient)
    assert llm.configured
    assert llm.model_name == "gemini-2.5-flash"


def test_build_chat_llm_openai():
    settings = Settings(
        chat_llm_provider="openai",
        openai_api_key="o-key",
        openai_model="gpt-4o-mini",
    )
    llm = build_chat_llm(settings)
    assert isinstance(llm, OpenAIChatClient)
    assert llm.configured


def test_build_chat_llm_rejects_unknown():
    with pytest.raises(ValueError, match="unsupported CHAT_LLM_PROVIDER"):
        build_chat_llm(Settings(chat_llm_provider="claude"))
