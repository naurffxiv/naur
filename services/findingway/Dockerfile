# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o findingway .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D appuser
WORKDIR /findingway
COPY --from=builder /src/findingway .
COPY --from=builder /src/config.yaml .
RUN chown -R appuser:appuser /findingway
USER appuser
ENTRYPOINT ["/findingway/findingway"]
