FROM python:3.14-slim

COPY --from=ghcr.io/astral-sh/uv:0.10.4 /uv /uvx /bin/

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Required for compatibility with prebuilt Python wheels
ENV PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1

WORKDIR /opt/app

# Create non-root user
RUN useradd -m -u 10001 appuser && chown appuser:appuser /opt/app
USER appuser

# Copy dependency declarations
COPY --chown=appuser:appuser pyproject.toml uv.lock ./

# Install dependencies (frozen lockfile)
RUN --mount=type=cache,target=/home/appuser/.cache/uv,uid=10001,gid=10001 \
    uv sync --frozen --no-dev --link-mode=copy

COPY --chown=appuser:appuser . .

CMD ["uv", "run", "--no-sync", "python", "main.py"]
