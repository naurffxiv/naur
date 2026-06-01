FROM python:3.14-slim

COPY --from=ghcr.io/astral-sh/uv:0.11.14 /uv /uvx /bin/

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc-dev \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

ENV PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1

WORKDIR /opt/app

RUN useradd -m -u 10001 appuser && chown appuser:appuser /opt/app
USER appuser

COPY --chown=appuser:appuser pyproject.toml uv.lock ./

RUN --mount=type=cache,target=/home/appuser/.cache/uv,uid=10001,gid=10001 \
    uv sync --frozen --no-dev --no-install-project --link-mode=copy

COPY --chown=appuser:appuser . .

RUN --mount=type=cache,target=/home/appuser/.cache/uv,uid=10001,gid=10001 \
    uv sync --frozen --no-dev --link-mode=copy

CMD ["uv", "run", "--no-sync", "uvicorn", "moddingway_api.main:app", "--host", "0.0.0.0", "--port", "8000"]
