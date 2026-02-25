VENV := venv

ifeq ($(OS),Windows_NT)
    VENV_PYTHON := $(VENV)/Scripts/python.exe
    BASE_PYTHON := py -3.14
    RM_VENV := if exist $(VENV) rmdir /s /q $(VENV)
else
    VENV_PYTHON := $(VENV)/bin/python
    BASE_PYTHON := python3
    RM_VENV := rm -rf $(VENV)
endif

PY := $(if $(wildcard $(VENV_PYTHON)),$(VENV_PYTHON),$(BASE_PYTHON))

format:
	@$(PY) -m ruff format .
	@$(PY) -m ruff check --fix .

lint: format ty

stop:
	docker compose down

install:
	@echo "Creating virtual environment and installing dependencies..."
	@$(RM_VENV)
	@$(BASE_PYTHON) -m venv $(VENV)
	@$(VENV_PYTHON) -m pip install --upgrade pip
	@$(VENV_PYTHON) -m pip install -r requirements-dev.txt

clean-venv:
	@$(RM_VENV)

clean-docker:
	docker image prune -a

clean: clean-docker clean-venv

build:
	docker compose build python-app

run:
	docker compose down --remove-orphans
	docker compose up python-app-local --build

python-run:
	$(PY) main.py

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

api:
	$(PY) -m uvicorn moddingway_api.main:app --port 8000

api-reload:
	$(PY) -m uvicorn moddingway_api.main:app --port 8000 --reload

test:
	$(PY) -m pytest ../../tests/moddingway/unit

ty:
	@$(PY) -m ty check .

.PHONY: format lint stop install clean clean-venv clean-docker build run database-run database-run-ephemeral database-clean database-clean-ephemeral api api-reload test ty python-run
