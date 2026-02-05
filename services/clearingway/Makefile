.PHONY: test build format lint lint-fix

test:
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

test-html: test
	go tool cover -html=coverage.out

build:
	go build -o bin/clearingway ./cmd/clearingway

run: build
	./bin/clearingway

format:
	goimports -w .

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix
