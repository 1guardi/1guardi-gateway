.PHONY: test test-verbose test-cover lint build infra

BACKEND_DIR := ./backend

test:
	cd $(BACKEND_DIR) && go test -v ./... -race -count=1

test-cover:
	cd $(BACKEND_DIR) && go test -v -coverprofile=coverage.out ./...
	cd $(BACKEND_DIR) && go tool cover -func=coverage.out

lint:
	cd $(BACKEND_DIR) && go vet ./...

build:
	cd $(BACKEND_DIR) && go build ./...

infra:
	cd $(BACKEND_DIR) && $(MAKE) infra
