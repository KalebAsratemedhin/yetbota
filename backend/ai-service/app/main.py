from fastapi import FastAPI

from app.api.routes import health


def create_app() -> FastAPI:
    app = FastAPI(
        title="YetBota AI Service",
        description=(
            "Post ingestion (RAG), similarity, and Surreal similarity graph. "
            "Scaffold — business APIs not implemented yet."
        ),
        version="0.1.0",
    )
    app.include_router(health.router)
    return app


app = create_app()
