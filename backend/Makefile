.PHONY: dev build tidy lint test up down logs

## Bootstrap — run once after cloning
tidy:
	go mod tidy

## Run both servers locally (no Docker)
dev:
	go run ./cmd/gateway

## Build the binary
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/gateway ./cmd/gateway

## Run tests
test:
	go test ./... -race -count=1

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## ── Docker Compose ──────────────────────────────────────────────────────────

## Start the full stack (builds gateway image)
up:
	docker compose -f deployments/docker-compose.yml up --build -d

## Tail logs from all services
logs:
	docker compose -f deployments/docker-compose.yml logs -f

## Tear down and remove volumes
down:
	docker compose -f deployments/docker-compose.yml down -v
