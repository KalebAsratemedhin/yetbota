from surrealdb import AsyncSurreal, RecordID

from application.errors import IndexingFailed
from domain.entities import SimilarPost
from infrastructure.config.settings import SurrealDBSettings
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_SCHEMA = """
DEFINE TABLE OVERWRITE post SCHEMALESS;
DEFINE TABLE OVERWRITE similar_to TYPE RELATION IN post OUT post;
DEFINE TABLE OVERWRITE duplicate_of TYPE RELATION IN post OUT post;
"""


class SurrealDBSimilarityGraph:
    def __init__(self, settings: SurrealDBSettings) -> None:
        self._settings = settings
        self._db: AsyncSurreal | None = None

    async def connect(self) -> None:
        db = AsyncSurreal(self._settings.url)
        await db.connect()
        await db.signin(
            {
                "username": self._settings.username,
                "password": self._settings.password,
            }
        )
        await db.use(self._settings.namespace, self._settings.database)
        await db.query(_SCHEMA)
        self._db = db
        logger.info(
            "surrealdb.connected",
            url=self._settings.url,
            namespace=self._settings.namespace,
            database=self._settings.database,
        )

    async def update_similar(self, post_id: str, similar: list[SimilarPost]) -> None:
        if self._db is None:
            raise RuntimeError("surrealdb graph not connected")
        rows = [{"post_id": s.post_id, "score": float(s.score)} for s in similar]
        try:
            await self._db.query(
                """
                LET $p = type::record('post', $post_id);
                UPSERT $p;
                DELETE similar_to WHERE in = $p;
                FOR $row IN $similar {
                    LET $q = type::record('post', $row.post_id);
                    UPSERT $q;
                    RELATE $p->similar_to->$q SET score = $row.score;
                };
                """,
                {"post_id": post_id, "similar": rows},
            )
        except Exception as exc:
            raise IndexingFailed(f"surrealdb update_similar failed: {exc}", cause=exc) from exc

    async def mark_duplicate(
        self, duplicate_post_id: str, original_post_id: str, score: float
    ) -> None:
        if self._db is None:
            raise RuntimeError("surrealdb graph not connected")
        edge_id = f"{duplicate_post_id}:{original_post_id}"
        try:
            await self._db.query(
                """
                LET $dup = type::record('post', $duplicate_id);
                LET $orig = type::record('post', $original_id);
                LET $edge = type::record('duplicate_of', $edge_id);
                UPSERT $dup;
                UPSERT $orig;
                RELATE OR UPDATE $dup->$edge->$orig SET score = $score;
                """,
                {
                    "duplicate_id": duplicate_post_id,
                    "original_id": original_post_id,
                    "edge_id": edge_id,
                    "score": float(score),
                },
            )
        except Exception as exc:
            raise IndexingFailed(f"surrealdb mark_duplicate failed: {exc}", cause=exc) from exc

    async def delete_post(self, post_id: str) -> None:
        if self._db is None:
            raise RuntimeError("surrealdb graph not connected")
        try:
            await self._db.query(
                """
                LET $p = type::record('post', $post_id);
                DELETE similar_to WHERE in = $p OR out = $p;
                DELETE duplicate_of WHERE in = $p OR out = $p;
                DELETE $p;
                """,
                {"post_id": post_id},
            )
        except Exception as exc:
            raise IndexingFailed(f"surrealdb delete_post failed: {exc}", cause=exc) from exc

    async def similar_posts_tree(
        self, post_id: str, max_depth: int
    ) -> list[dict[str, int | str]]:
        if self._db is None:
            raise RuntimeError("surrealdb graph not connected")
        if max_depth < 1:
            return []

        seen: dict[str, int] = {}
        post = RecordID("post", post_id)
        try:
            for depth in range(1, max_depth + 1):
                result = await self._db.query(
                    f"SELECT VALUE meta::id(id) FROM $post.{{{depth}}}(<->similar_to<->post);",
                    {"post": post},
                )
                ids = _as_id_list(result)
                for pid in ids:
                    if not pid or pid == post_id:
                        continue
                    if pid not in seen:
                        seen[pid] = depth
        except Exception as exc:
            raise IndexingFailed(
                f"surrealdb similar_posts_tree failed: {exc}", cause=exc
            ) from exc

        return [{"post_id": pid, "depth": depth} for pid, depth in seen.items()]

    async def close(self) -> None:
        if self._db is not None:
            await self._db.close()
            self._db = None
            logger.info("surrealdb.closed")


def _as_id_list(result: object) -> list[str]:
    if result is None:
        return []
    if isinstance(result, list):
        out: list[str] = []
        for item in result:
            if isinstance(item, str):
                out.append(item)
            elif isinstance(item, RecordID):
                out.append(str(item.id))
        return out
    return []
