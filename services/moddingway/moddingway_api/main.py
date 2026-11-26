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


@app.get("/", include_in_schema=False)
def root():
    return RedirectResponse(url="/docs")


@app.get("/health")
def health_check():
    return {"status": "ok"}


# FastAPI keeps included routers as separate entries rather than copying their
# routes into app.routes, so add_pagination(app) never sees the paginated routes.
# Add pagination to each router as it is included.
for router, tag in [
    (user_router, "user"),
    (mod_router, "mod"),
    (banneduser_router, "ban"),
    (banform_router, "forms"),
]:
    app.include_router(add_pagination(router), tags=[tag])
