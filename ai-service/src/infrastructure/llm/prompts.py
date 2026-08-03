from pathlib import Path
from typing import Any

from jinja2 import Environment, FileSystemLoader, StrictUndefined, select_autoescape

from domain.entities import ContentKind, Message, ScoredChunk

_TEMPLATE_DIR = Path(__file__).resolve().parent / "templates"


SCREENING_RESPONSE_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {
        "profanity": {"type": "number"},
        "hate": {"type": "number"},
        "sexual": {"type": "number"},
        "violence": {"type": "number"},
        "harassment": {"type": "number"},
    },
    "required": ["profanity", "hate", "sexual", "violence", "harassment"],
}


def _environment() -> Environment:
    return Environment(
        loader=FileSystemLoader(_TEMPLATE_DIR),
        autoescape=select_autoescape(disabled_extensions=("jinja",)),
        undefined=StrictUndefined,
        trim_blocks=False,
        lstrip_blocks=False,
    )


class AssistantPromptBuilder:
    def __init__(self) -> None:
        self._template = _environment().get_template("assistant.jinja")

    def __call__(
        self, hits: list[ScoredChunk], query: str, history: list[Message] | None = None
    ) -> str:
        return self._template.render(
            chunks=[h.chunk for h in hits],
            query=query,
            history=history or [],
        )


class PlaceSummaryPromptBuilder:
    def __init__(self) -> None:
        self._template = _environment().get_template("place_summary.jinja")

    def __call__(self, chunks: list) -> str:
        return self._template.render(chunks=chunks)


class ScreeningPromptBuilder:
    def __init__(self) -> None:
        self._template = _environment().get_template("screening.jinja")

    def __call__(self, text: str, kind: ContentKind) -> str:
        return self._template.render(text=text, kind=kind)
