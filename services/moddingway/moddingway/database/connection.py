import logging
import os
import threading

import psycopg2
from psycopg2.extensions import cursor

from moddingway.settings import get_settings

settings = get_settings()
logger = logging.getLogger(__name__)


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

    # This is run on startup
    def connect(self):
        try:
            self._connection = psycopg2.connect(
                host=settings.postgres_host,
                port=settings.postgres_port,
                dbname=settings.database_name,
                user=settings.postgres_username,
                password=settings.postgres_password,
            )
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
        with self._connect_lock:
            if self._connection is None or getattr(self._connection, "closed", True):
                self.connect()
            return self._connection.cursor()
