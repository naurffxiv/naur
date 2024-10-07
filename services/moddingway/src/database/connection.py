import psycopg2
from psycopg2.extensions import cursor
import logging
from settings import get_settings

settings = get_settings()
logger = logging.getLogger(__name__)


class DatabaseConnection:
    """
    Singleton class for managing database connection
    """

    _instance = None

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

    # TODO: properly spin down DB connection on exit

    def get_cursor(self) -> cursor:
        return self._connection.cursor()
