from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from infrastructure.config import Settings
from infrastructure.observability import get_logger
from interfaces.http.routes import assistant, metrics, places, screening

logger = get_logger(__name__)


def create_app(settings: Settings) -> FastAPI:
    app = FastAPI(
        title=settings.app.name,
        version=settings.app.version,
        debug=settings.app.debug,
    )
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors.allow_origins,
        allow_credentials=settings.cors.allow_credentials,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    logger.info(
        "cors.configured",
        configured_hosts=settings.cors.hosts,
        allow_origins=settings.cors.allow_origins,
    )
    app.include_router(metrics.router)
    app.include_router(assistant.router)
    app.include_router(screening.router)
    app.include_router(places.router)
    return app
