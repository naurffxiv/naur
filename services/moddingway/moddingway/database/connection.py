import asyncio
import logging
import os
import threading
from collections.abc import Iterator
from contextlib import contextmanager
from contextvars import ContextVar

import psycopg2
from psycopg2.extensions import connection, cursor

from moddingway.settings import get_settings

settings = get_settings()
logger = logging.getLogger(__name__)

# The open transaction's connection and the task (or thread) that owns it.
# Tasks created inside a transaction inherit this value, so the owner is checked
# before the connection is used.
_active_transaction: ContextVar[tuple[connection, object] | None] = ContextVar(
    "active_transaction", default=None
)


def _current_owner() -> object:
    """The running asyncio task, or the current thread when there is none."""
    try:
        task = asyncio.current_task()
    except RuntimeError:
        task = None
    return task if task is not None else threading.current_thread()


def _owned_transaction() -> connection | None:
    """The connection of the transaction the caller owns, if any."""
    active = _active_transaction.get()
    if active is None:
        return None
    transaction_connection, owner = active
    return transaction_connection if owner is _current_owner() else None


class DatabaseConnection:
    """
    Singleton class for managing database connection
    """

    _instance = None
    _connect_lock = threading.Lock()

    # Prevent multiple instances of the connection to be created
    def __new__(cls, *args, **kwargs):
        if cls._instance is None:
            cls._instance = super().__new__(cls, *args, **kwargs)
            cls._instance._connection = None
        return cls._instance

    @staticmethod
    def _open_connection():
        return psycopg2.connect(
            host=settings.postgres_host,
            port=settings.postgres_port,
            dbname=settings.database_name,
            user=settings.postgres_username,
            password=settings.postgres_password,
        )

    # This is run on startup
    def connect(self):
        try:
            self._connection = self._open_connection()
            self._connection.set_session(autocommit=True)
        except Exception as e:
            logger.error(f"Failed to connect to database: {e}", exc_info=e)
            raise ValueError(f"Failed to connect to database: {e}") from e

    def create_tables(self):
        """
        Created the tables the application relies on. This only needs to be run on startup
        """
        base_dir = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))
        script_file_path = os.path.join(base_dir, "postgres", "create_tables.sql")

        if not os.path.exists(script_file_path):
            logger.error(
                "create_tables.sql not found at %s", os.path.abspath(script_file_path)
            )
            raise FileNotFoundError(script_file_path)

        with self.get_cursor() as cursor:
            with open(script_file_path, encoding="utf-8") as fd:
                script = fd.read()
                cursor.execute(script)

    def disconnect(self):
        """
        Close the underlying connection if open.
        """
        with self._connect_lock:
            try:
                if self._connection is not None and not self._connection.closed:
                    self._connection.close()
            except Exception:
                logger.exception("Error while closing DB connection")

    def get_cursor(self) -> cursor:
        """
        Return a cursor on the current transaction's connection when called inside
        ``transaction()``, otherwise on the shared autocommit connection.
        """
        transaction_connection = _owned_transaction()
        if transaction_connection is not None:
            return transaction_connection.cursor()
        with self._connect_lock:
            if self._connection is None or getattr(self._connection, "closed", True):
                self.connect()
            return self._connection.cursor()

    @contextmanager
    def transaction(self) -> Iterator[None]:
        """
        Run every database call made inside the block in one transaction.

        The transaction is committed when the block exits normally and rolled
        back if it raises. It uses a dedicated, non-autocommit connection that is
        only used by the task (or thread) that opened it, so statements issued
        by other coroutines on the shared autocommit connection (e.g. while this
        block awaits a Discord call) never join it. That includes tasks created
        inside the block, which inherit the ``ContextVar`` but not ownership.
        A nested ``transaction()`` in the same task joins the outer one.
        """
        if _owned_transaction() is not None:
            yield
            return

        transaction_connection = self._open_connection()
        token = _active_transaction.set((transaction_connection, _current_owner()))
        try:
            with transaction_connection:
                yield
        finally:
            _active_transaction.reset(token)
            transaction_connection.close()
