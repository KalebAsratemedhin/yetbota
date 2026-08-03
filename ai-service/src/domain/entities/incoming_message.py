from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class IncomingMessage:
    body: bytes
    delivery_count: int
    reply_to: str | None = None
    correlation_id: str | None = None
