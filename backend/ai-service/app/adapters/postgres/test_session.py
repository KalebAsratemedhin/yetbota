from app.adapters.postgres.session import asyncpg_database_url


def test_asyncpg_database_url_normalizes_common_forms():
    assert (
        asyncpg_database_url("postgres://u:p@h:5432/db")
        == "postgresql+asyncpg://u:p@h:5432/db"
    )
    assert (
        asyncpg_database_url("postgresql://u:p@h/db")
        == "postgresql+asyncpg://u:p@h/db"
    )
    assert (
        asyncpg_database_url("postgresql+asyncpg://u:p@h/db")
        == "postgresql+asyncpg://u:p@h/db"
    )
    assert (
        asyncpg_database_url("postgresql+psycopg://u:p@h/db")
        == "postgresql+asyncpg://u:p@h/db"
    )
