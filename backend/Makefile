.PHONY: dev build tidy lint test up down logs infra

## Bootstrap — run once after cloning
tidy:
	go mod tidy

## Run both servers locally (no Docker) — loads .env if present
dev:
	@[ -f .env ] && export $$(grep -v '^#' .env | xargs) ; go run ./cmd/gateway

## Build the binary
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/gateway ./cmd/gateway

## Run tests
test:
	go test -v ./... -race -count=1

## Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## ── Docker Compose ──────────────────────────────────────────────────────────

# Services in the backend profile
BACKEND_SERVICES := gateway

## Start the full stack (builds gateway image)
up:
	docker compose -f deployments/docker-compose.yml --profile backend --profile infra up --build -d

## Start only infra dependencies (postgres, redis, otel, jaeger)
infra:
	docker compose -f deployments/docker-compose.yml --profile backend --profile infra stop $(BACKEND_SERVICES)
	docker compose -f deployments/docker-compose.yml --profile infra up -d

## Tail logs from all services
logs:
	docker compose -f deployments/docker-compose.yml --profile backend --profile infra logs -f

## Tear down and remove volumes
down:
	docker compose -f deployments/docker-compose.yml --profile backend --profile infra down 

downv:
	docker compose -f deployments/docker-compose.yml --profile backend --profile infra down -v 