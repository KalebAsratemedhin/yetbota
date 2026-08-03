from typing import Any, Protocol, runtime_checkable


@runtime_checkable
class LLM(Protocol):
    async def generate(
        self,
        prompt: str,
        *,
        max_tokens: int,
        temperature: float,
    ) -> str: ...

    async def classify(
        self,
        prompt: str,
        *,
        schema: dict[str, Any],
        temperature: float = 0.0,
    ) -> dict[str, Any]: ...
