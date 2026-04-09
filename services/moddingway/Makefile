lock:
	@uv lock

sync: lock
	@uv sync --all-groups --link-mode=copy --quiet

format: sync
	@uv run ruff format . ../../tests/moddingway
	@uv run ruff check --fix . ../../tests/moddingway

lint: format ty

stop:
	docker compose down


clean-venv:
	@uv venv --clear

clean-docker:
	docker image prune -a

clean: clean-docker clean-venv

build:
	docker compose build python-app

run:
	docker compose down --remove-orphans
	docker compose up python-app-local --build

python-run: sync
	@uv run python main.py

database-run:
	docker compose -f postgres.yml down
	docker compose -f postgres.yml up -d postgres_local

database-clean:
	docker exec postgres_db rm scripts -r -f
	docker exec postgres_db mkdir scripts
	docker cp postgres postgres_db:scripts
	docker exec postgres_db psql -f scripts/postgres/drop_all_tables.sql -U moddingwayLocalDB moddingway
	docker exec postgres_db psql -f scripts/postgres/create_tables.sql -U moddingwayLocalDB moddingway
	docker exec postgres_db psql -f scripts/postgres/seed_data.sql -U moddingwayLocalDB moddingway

database-run-ephemeral:
	docker compose -f postgres.yml down
	docker compose -f postgres.yml up -d postgres_ephemeral

database-clean-ephemeral:
	docker exec postgres_db_ephemeral rm scripts -r -f
	docker exec postgres_db_ephemeral mkdir scripts
	docker cp postgres postgres_db_ephemeral:scripts
	docker exec postgres_db_ephemeral psql -f scripts/postgres/drop_all_tables.sql -U moddingwayLocalDB moddingway
	docker exec postgres_db_ephemeral psql -f scripts/postgres/create_tables.sql -U moddingwayLocalDB moddingway
	docker exec postgres_db_ephemeral psql -f scripts/postgres/seed_data.sql -U moddingwayLocalDB moddingway

api: sync
	@uv run uvicorn moddingway_api.main:app --port 8000

api-reload: sync
	@uv run uvicorn moddingway_api.main:app --port 8000 --reload

test: sync
	@uv run pytest ../../tests/moddingway/unit

ty: sync
	@uv run ty check . ../../tests/moddingway

.PHONY: lock sync format lint stop clean clean-venv clean-docker build run database-run database-run-ephemeral database-clean database-clean-ephemeral api api-reload test ty python-run
