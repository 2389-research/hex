.PHONY: build test test-short test-coverage clean install run lint fmt vet release snapshot help

# Build configuration
BINARY_NAME=hex
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X github.com/harper/hex/internal/core.Version=$(VERSION) -X github.com/harper/hex/internal/core.Commit=$(COMMIT) -X github.com/harper/hex/internal/core.Date=$(DATE)"

## build: Build binary for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/hex

## test: Run all tests with race detection
test:
	go test -v -race ./...

## test-short: Run short tests only
test-short:
	go test -v -short ./...

## test-coverage: Run tests with coverage report
test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.txt coverage.html
	rm -rf dist/
	go clean

## install: Build and install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/hex

## run: Run without building (pass args with make run ARGS="...")
run:
	go run ./cmd/hex $(ARGS)

## lint: Run golangci-lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint"; \
		echo "  or go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## fmt: Format code with gofmt
fmt:
	gofmt -s -w .

## vet: Run go vet
vet:
	go vet ./...

## release: Test release build locally (requires goreleaser)
release:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed. Install with:"; \
		echo "  brew install goreleaser"; \
		echo "  or go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

## snapshot: Build snapshot release for testing
snapshot:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser build --snapshot --clean; \
	else \
		echo "goreleaser not installed. Install with:"; \
		echo "  brew install goreleaser"; \
		exit 1; \
	fi

## deps: Download dependencies
deps:
	go mod download
	go mod verify

## tidy: Tidy dependencies
tidy:
	go mod tidy

## verify: Run all verification steps (fmt, vet, lint, test)
verify: fmt vet lint test
	@echo "All checks passed!"

## help: Show this help message
help:
	@echo "Hex CLI Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

.DEFAULT_GOAL := build
