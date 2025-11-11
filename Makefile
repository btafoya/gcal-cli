.PHONY: build test clean install help coverage lint

# Variables
BINARY_NAME=gcal-cli
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-X github.com/btafoya/gcal-cli/internal/commands.Version=$(VERSION) \
        -X github.com/btafoya/gcal-cli/internal/commands.Commit=$(COMMIT) \
        -X github.com/btafoya/gcal-cli/internal/commands.BuildDate=$(BUILD_DATE)

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/gcal-cli
	@echo "Build complete: ./$(BINARY_NAME)"

## build-release: Build optimized release binary
build-release:
	@echo "Building release binary..."
	@go build -ldflags "$(LDFLAGS) -s -w" -o $(BINARY_NAME) ./cmd/gcal-cli
	@echo "Release build complete: ./$(BINARY_NAME)"

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install -ldflags "$(LDFLAGS)" ./cmd/gcal-cli
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

## test-short: Run tests without verbose output
test-short:
	@go test ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## tidy: Tidy and verify module dependencies
tidy:
	@echo "Tidying module dependencies..."
	@go mod tidy
	@go mod verify

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## run: Build and run the application
run: build
	@./$(BINARY_NAME)

## run-help: Build and show help
run-help: build
	@./$(BINARY_NAME) --help

## run-version: Build and show version
run-version: build
	@./$(BINARY_NAME) version

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download

## verify: Run all verification steps (fmt, vet, test)
verify: fmt vet test
	@echo "All verification steps passed!"
