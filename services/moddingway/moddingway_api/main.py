import logging
from contextlib import asynccontextmanager

import anyio
from fastapi import FastAPI
from fastapi.responses import RedirectResponse
from fastapi_pagination import add_pagination

from moddingway.database import DatabaseConnection
from moddingway_api.routes import (
    banform_router,
    banneduser_router,
    mod_router,
    user_router,
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    configure_logging()
    database_connection = DatabaseConnection()
    await anyio.to_thread.run_sync(database_connection.connect)  # ty: ignore[unresolved-attribute]
    await anyio.to_thread.run_sync(database_connection.create_tables)  # ty: ignore[unresolved-attribute]
    app.state.db = database_connection

    try:
        yield
    finally:
        try:
            app.state.db.disconnect()
        except Exception:
            logging.getLogger(__name__).exception("Error during DB disconnect")


def configure_logging():
    logging.basicConfig()


app = FastAPI(title="Moddingway API", lifespan=lifespan)
add_pagination(app)


@app.get("/", include_in_schema=False)
def root():
    return RedirectResponse(url="/docs")


@app.get("/health")
def health_check():
    return {"status": "ok"}


app.include_router(user_router, tags=["user"])
app.include_router(mod_router, tags=["mod"])
app.include_router(banneduser_router, tags=["ban"])
app.include_router(banform_router, tags=["forms"])
