from __future__ import annotations


class AIServiceError(Exception):
    code: str = "UNKNOWN_ERROR"
    transient: bool = False

    def __init__(
        self,
        message: str = "",
        *,
        cause: Exception | None = None,
        transient: bool | None = None,
    ) -> None:
        super().__init__(message)
        self.message = message
        self.__cause__ = cause
        if transient is not None:
            self.transient = transient

class ConfigError(AIServiceError):
    code = "CONFIG_ERROR"


class MessageMalformed(AIServiceError):
    code = "MESSAGE_MALFORMED"


class MetadataError(AIServiceError):
    code = "METADATA_ERROR"


class EmbeddingFailed(AIServiceError):
    code = "EMBED_FAILED"
    transient = True


class SimilaritySearchFailed(AIServiceError):
    code = "SIMILARITY_FAILED"
    transient = True


class IndexingFailed(AIServiceError):
    code = "INDEXING_GAP"


class LLMUnavailable(AIServiceError):
    code = "LLM_UNAVAILABLE"
    transient = True


class ScreeningUnavailable(AIServiceError):
    code = "SCREENING_UNAVAILABLE"
    transient = True


class ChatNotFound(AIServiceError):
    code = "CHAT_NOT_FOUND"


class PostNotFound(AIServiceError):
    code = "POST_NOT_FOUND"


class PostSourceUnavailable(AIServiceError):
    code = "POST_SOURCE_UNAVAILABLE"
    transient = True


class CommentNotFound(AIServiceError):
    code = "COMMENT_NOT_FOUND"


class CommentSourceUnavailable(AIServiceError):
    code = "COMMENT_SOURCE_UNAVAILABLE"
    transient = True
