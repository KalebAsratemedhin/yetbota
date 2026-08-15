from domain.entities.place_match import PlaceMatchQuery, PlaceMatchResult
from domain.ports.embedder import Embedder
from domain.ports.vector_store import VectorStore
from infrastructure.observability import get_logger

logger = get_logger(__name__)

_DEFAULT_GEO_RADIUS_KM = 0.5


class MatchPlace:
    def __init__(self, *, embedder: Embedder, vector_store: VectorStore) -> None:
        self._embedder = embedder
        self._vector_store = vector_store

    async def execute(self, query: PlaceMatchQuery) -> PlaceMatchResult:
        text = query.text.strip()
        if not text:
            return PlaceMatchResult()

        embedding = await self._embedder.embed(text, task_type="RETRIEVAL_QUERY")
        lat = query.latitude
        lon = query.longitude
        radius_m: float | None = None
        if lat is not None and lon is not None:
            radius_km = query.radius_km if query.radius_km and query.radius_km > 0 else _DEFAULT_GEO_RADIUS_KM
            radius_m = radius_km * 1000.0

        hits = await self._vector_store.search(
            embedding,
            limit=query.limit,
            lat=lat,
            lon=lon,
            radius_m=radius_m,
        )

        best: dict[str, float] = {}
        for hit in hits:
            if query.exclude_source_id and hit.chunk.source_id == query.exclude_source_id:
                continue
            place_id = (hit.chunk.metadata or {}).get("place_id", "").strip()
            if not place_id:
                continue
            current = best.get(place_id)
            if current is None or hit.score > current:
                best[place_id] = hit.score

        if not best:
            logger.info("place_match.no_candidate", text_len=len(text))
            return PlaceMatchResult()

        place_id, score = max(best.items(), key=lambda kv: kv[1])
        logger.info("place_match.hit", place_id=place_id, score=round(score, 4))
        return PlaceMatchResult(place_id=place_id, score=score)
