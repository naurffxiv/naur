"""Run the database layer against a real, throwaway Postgres container.

Every test here is skipped when Docker is not available.
"""

import shutil
import subprocess
import time
import uuid

import psycopg2
import pytest

from moddingway.database import connection

POSTGRES_IMAGE = "postgres:16"
TABLES = [
    "roles",
    "notes",
    "strikes",
    "exiles",
    "forms",
    "users",
    "announcements",
    "commands",
]


def _docker_available() -> bool:
    if shutil.which("docker") is None:
        return False
    return (
        subprocess.run(["docker", "info"], capture_output=True, check=False).returncode
        == 0
    )


@pytest.fixture(scope="session")
def postgres():
    if not _docker_available():
        pytest.skip("Docker is not available")

    name = f"moddingway-test-{uuid.uuid4().hex[:8]}"
    subprocess.run(
        [
            "docker",
            "run",
            "-d",
            "--rm",
            "--name",
            name,
            "-e",
            "POSTGRES_USER=test",
            "-e",
            "POSTGRES_PASSWORD=test",
            "-e",
            "POSTGRES_DB=moddingway",
            "-e",
            "TZ=UTC",
            "-e",
            "PGTZ=UTC",
            "-p",
            "127.0.0.1::5432",
            "--tmpfs",
            "/var/lib/postgresql/data",
            POSTGRES_IMAGE,
            # Durability is not needed for throwaway test data.
            "-c",
            "fsync=off",
            "-c",
            "synchronous_commit=off",
            "-c",
            "full_page_writes=off",
        ],
        check=True,
        capture_output=True,
    )
    try:
        port = (
            subprocess.run(
                ["docker", "port", name, "5432/tcp"],
                check=True,
                capture_output=True,
                text=True,
            )
            .stdout.strip()
            .rsplit(":", 1)[1]
        )
        params = {
            "host": "127.0.0.1",
            "port": port,
            "dbname": "moddingway",
            "user": "test",
            "password": "test",
        }

        deadline = time.monotonic() + 60
        while True:
            try:
                psycopg2.connect(**params).close()
                break
            except psycopg2.OperationalError:
                if time.monotonic() > deadline:
                    raise
                time.sleep(0.5)

        yield params
    finally:
        subprocess.run(["docker", "stop", name], capture_output=True, check=False)


@pytest.fixture(scope="session")
def database(postgres):
    """Point the app's connection at the container and create the schema."""
    settings = connection.settings
    settings.postgres_host = postgres["host"]
    settings.postgres_port = postgres["port"]
    settings.database_name = postgres["dbname"]
    settings.postgres_username = postgres["user"]
    settings.postgres_password = postgres["password"]

    db = connection.DatabaseConnection()
    db.disconnect()
    db._connection = None
    db.create_tables()
    yield db
    db.disconnect()


@pytest.fixture(autouse=True)
def clean_tables(database):
    with database.get_cursor() as cursor:
        cursor.execute(f"TRUNCATE {', '.join(TABLES)} RESTART IDENTITY CASCADE")
    yield
