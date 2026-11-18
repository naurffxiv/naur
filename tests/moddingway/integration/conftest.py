"""
Integration tests run against a real Postgres (the `postgres_test` compose
service). The database is never mocked.
"""

from collections.abc import Callable

import pytest

from moddingway.database import DatabaseConnection

TABLES = ["users", "exiles", "strikes", "notes", "roles", "forms"]


@pytest.fixture(scope="session")
def database() -> DatabaseConnection:
    connection = DatabaseConnection()
    connection.create_tables()
    return connection


def _truncate(database: DatabaseConnection):
    with database.get_cursor() as cursor:
        cursor.execute(f"TRUNCATE {', '.join(TABLES)} RESTART IDENTITY CASCADE")


@pytest.fixture(autouse=True)
def clean_database(database: DatabaseConnection):
    _truncate(database)
    yield
    _truncate(database)


@pytest.fixture
def query(database: DatabaseConnection) -> Callable[..., list[tuple]]:
    def __query(sql: str, *params) -> list[tuple]:
        with database.get_cursor() as cursor:
            cursor.execute(sql, params)
            return cursor.fetchall()

    return __query
