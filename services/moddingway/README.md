# Moddingway

Discord moderation bot for NAUR.

## Development Setup

### First Time Setup

1. **Create a Bot**: Create a personal version of the moddingway bot via Discord's developer portal. Follow [Discord's Getting Started](https://discord.com/developers/docs/quick-start/getting-started#step-1-creating-an-app).
2. **Environment**: Copy `.env_example` to `.env` and configure the required environment variables (see [Environment Variables](#environment-variables) below).
3. **Docker Desktop**: Install [Docker Desktop](https://www.docker.com/products/docker-desktop/) to run the containerized database.
4. **Install uv**: Follow the [official uv installation guide](https://docs.astral.sh/uv/getting-started/installation/).
5. **Install Dependencies**: Run `make sync`. This will:
   - Automatically ensure the correct Python version (3.14) is available.
   - Create a local `.venv` if it doesn't exist.
   - Synchronize all dependencies from `uv.lock`.

### Server Setup

The server used for development needs the following roles (in priority order):

- Mod
- Verified
- Exiled

Give yourself the `Mod` role to run moderation commands.

## Make Commands

We use a Makefile to simplify common development tasks.

### Setup & Installation

- `make lock` - Regenerates `uv.lock` from `pyproject.toml`. Run after editing dependencies.
- `make sync` - Runs `make lock` then `uv sync` to install/update all Python dependencies.
- `make clean-venv` - Clears and recreates the local `.venv` directory.
- `make clean-docker` - Removes unused Docker images to reclaim disk space.
- `make clean` - Runs both `clean-venv` and `clean-docker`.

### Running the Application

**Docker (Recommended):**

- `make run` - Stops existing containers (if running), rebuilds, and launches the bot with Docker. This is the most common command for development.
- `make build` - Builds the Docker image without running it.
- `make stop` - Stops all running Docker containers for this project.

**Local Python:**

- `make python-run` - Runs the bot locally using your virtual environment. Requires the database to be running first (`make database-run`).

### Database Management

- `make database-run` - Launches the containerized Postgres persistent database. The container uses values from your `.env` file for database configuration. Tables are automatically created when you first run the bot. Data persists between container restarts.
- `make database-clean` - Deletes all data in the persistent database, recreates tables, and loads seed data. Requires the database container to be running.
- `make database-run-ephemeral` - Launches a temporary Postgres database for testing. Uses `.env` for configuration. Tables are automatically created. Data does NOT persist between restarts.
- `make database-clean-ephemeral` - Deletes all data in the temporary database, recreates tables, and loads seed data. Requires the ephemeral database container to be running.

### API Development

- `make api` - Starts the moddingway API server at `localhost:8000`.
- `make api-reload` - Starts the API with hot-reload enabled, so you don't need to restart after code changes. **Recommended for API development.**

### Code Quality & Testing

- `make format` - Runs `ruff format` and `ruff check --fix`. **Required before merging.**
- `make lint` - Runs `format` then `ty check`. Full code quality pass.
- `make ty` - Runs `ty check` for type safety.
- `make test` - Runs automated unit tests with `pytest`.

## Dependency Management

We use only `pyproject.toml` and `uv.lock` for dependencies. `requirements.txt` files do not exist. All dependency management is handled exclusively through `uv`.

### Adding/Removing Dependencies

Use `uv` directly to manage packages:

```bash
# Add a production dependency
uv add package-name

# Add a development dependency
uv add --dev package-name

# Remove a dependency
uv remove package-name
```

Always commit both `pyproject.toml` and `uv.lock` together. After editing `pyproject.toml`, run `make lock` to regenerate the lockfile before committing.

## Typical Development Workflow

### Docker-based development (Recommended)

```bash
make sync            # Install dev tools (optional if only running)
make run             # Start the bot (includes database)
make format          # Before committing
```

### Local Python development

```bash
make sync             # First time setup/sync
make database-run     # Start database
make python-run       # Run bot locally
make format           # Format before committing
```

### API development

```bash
make sync             # First time setup/sync
make database-run     # Start database
make api-reload       # Start API with hot-reload
```

Visit `localhost:8000/docs` to view the API documentation.

## Testing

This application uses `pytest` for automated unit testing.

- Run `make test` to execute the full suite.

### Ruff & Ty

We use **Ruff** for high-performance linting and formatting, and **Ty** for type checking.

- Run `make format` to clean up code style.
- Run `make ty` to check for type-safety issues.

Most IDEs (like VS Code) have extensions for Ruff that can format on save.

### Reseeding the Database

To reset and re-seed the local database, run `make database-clean` while the postgres container is active.

## Troubleshooting

### Lockfile Mismatch

If you see errors about `uv.lock` being out of sync, run:

```bash
make lock
# then
make sync
```

Or directly with uv:

```bash
uv lock && uv sync
```

### Environment Issues

If the virtual environment seems corrupted, try:

```bash
make clean-venv
make install
```

### Python Version

The project requires Python 3.14. `uv` will manage this for you, but you can check the version with:

```bash
uv run python --version
```

## Environment Variables

Postgres-related information is configured in the environment variables instead of a pre-created user/password. For local development, create a `.env` file to populate the following environment variables.

### Testing Variables

- GUILD_ID
- DISCORD_TOKEN
- MOD_LOGGING_CHANNEL_ID
- POSTGRES_USER
- POSTGRES_PASSWORD
- POSTGRES_DB
- INACTIVE_FORUM_CHANNEL_ID
- INACTIVE_FORUM_DURATION
- PF_RECRUITMENT_CHANNEL_ID
- PF_RECRUITMENT_CHANNEL_DURATION
- NOTIFY_CHANNEL_ID
- PTC_EVENT_FORUM_ID
- PTC_EVENT_FORUM_DURATION
- EVENT_BOT_ID

### Release Variables

- DISCORD_TOKEN
- POSTGRES_HOST
- POSTGRES_PORT
- POSTGRES_DB
- POSTGRES_USER
- POSTGRES_PASSWORD
- MODDINGWAY_ENVIRONMENT

Defaults are set for `POSTGRES_PORT` (5432) and `POSTGRES_DB` (moddingway) if not specified.
`INACTIVE_FORUM_CHANNEL_ID` and `INACTIVE_FORUM_DURATION` are optional. The relevant task will not run if those environment variables are not defined.
