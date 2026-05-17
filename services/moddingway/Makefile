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

TEST_REPORTS_DIR := ../../tests/moddingway

test:
	docker compose run --rm --build python-test

test-ci:
	@echo "Running containerized tests..."
	docker compose build python-test
	-docker rm -f moddingway-test-container
	docker create --name moddingway-test-container moddingway-python-test
	@python -c "import subprocess, sys; \
		r = subprocess.run(['docker', 'start', '-a', 'moddingway-test-container']); \
		exit_code = r.returncode; \
		print('Exporting test artifacts...'); \
		subprocess.run(['docker', 'cp', 'moddingway-test-container:/opt/app/tests/moddingway/coverage.xml', '$(TEST_REPORTS_DIR)/coverage.xml'], check=False); \
		subprocess.run(['docker', 'cp', 'moddingway-test-container:/opt/app/tests/moddingway/htmlcov', '$(TEST_REPORTS_DIR)/htmlcov'], check=False); \
		subprocess.run(['docker', 'rm', '-f', 'moddingway-test-container'], check=False); \
		sys.exit(exit_code)"

ty: sync
	@uv run ty check . ../../tests/moddingway

.PHONY: lock sync format lint stop clean clean-venv clean-docker build run database-run database-run-ephemeral database-clean database-clean-ephemeral api api-reload test test-ci ty python-run
