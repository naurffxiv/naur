FROM python:3.14-slim

COPY --from=ghcr.io/astral-sh/uv:0.10.4 /uv /uvx /bin/

RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Required for compatibility with prebuilt Python wheels
ENV PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1

WORKDIR /opt/app

# Copy dependency declarations
COPY pyproject.toml uv.lock ./

# Install dependencies (frozen lockfile)
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev --link-mode=copy

COPY . .

CMD ["uv", "run", "--no-sync", "python", "main.py"]
