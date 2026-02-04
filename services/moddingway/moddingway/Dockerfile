FROM python:3.14-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Required for compatibility with prebuilt Python wheels
ENV PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1

WORKDIR /opt/app

# install required packages
RUN --mount=type=bind,source=requirements.txt,target=requirements.txt \
    uv pip install --system --no-cache -r requirements.txt

COPY . .

CMD ["python", "main.py"]
