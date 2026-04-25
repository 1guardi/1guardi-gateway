.PHONY: test test-verbose test-cover lint build

BACKEND_DIR := ./backend

test:
	cd $(BACKEND_DIR) && go test ./...

test-verbose:
	cd $(BACKEND_DIR) && go test -v ./...

test-cover:
	cd $(BACKEND_DIR) && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

lint:
	cd $(BACKEND_DIR) && go vet ./...

build:
	cd $(BACKEND_DIR) && go build ./...
