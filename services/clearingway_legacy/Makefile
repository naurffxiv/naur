.PHONY: dev up down logs watch build test

dev:
	docker compose up --build

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

watch:
	docker compose watch

build:
	docker compose build

test:
	go test ./... -race -covermode=atomic -coverprofile=coverage.out
