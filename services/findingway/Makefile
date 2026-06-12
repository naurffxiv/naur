.PHONY: test build tokens

test:
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

tokens:
	docker compose run --rm --env-file .env -e TOKENS_ONLY=true findingway

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.buildSha=`git rev-parse HEAD` -X main.buildTime=`date +'%Y-%m-%d_%T'`"

format:
	goimports -w .

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

redis-docker:
	docker run --name findingway-redis -d -p 6379:6379 redis:latest