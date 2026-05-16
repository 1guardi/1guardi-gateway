VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -X github.com/chaitanyabankanhal/ai-gateway/cli/internal/cmd.version=$(VERSION) \
           -X github.com/chaitanyabankanhal/ai-gateway/cli/internal/cmd.commit=$(COMMIT)

.PHONY: build install test tidy

build:
	go build -ldflags "$(LDFLAGS)" -o bin/aigw .

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./...

tidy:
	go mod tidy
