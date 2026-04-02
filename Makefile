BINARY_NAME=w9s
BUILD_DIR=build
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION_PKG=github.com/mzaran/w9s/internal/version
LDFLAGS=-ldflags "-X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).Date=$(DATE)"

.DEFAULT_GOAL := help

.PHONY: build clean test test-unit test-integration test-smoke lint fmt vet ci dev mock coverage help

## build: Build the w9s binary
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/w9s

## clean: Remove build artifacts
clean:
	@rm -rf $(BUILD_DIR)

## test: Run all tests
test:
	go test -timeout 30m ./internal/... ./pkg/...

## test-unit: Run unit tests with verbose output
test-unit:
	go test -timeout 30m -v ./internal/...

## test-integration: Run integration tests (requires W9S_INTEGRATION_TESTS=1)
test-integration:
	W9S_INTEGRATION_TESTS=1 go test -timeout 45m -v ./test/integration/...

## test-smoke: Run smoke tests in mock mode (requires tmux)
test-smoke: build
	W9S_ENABLE_MOCK=1 ./test/smoke/verify-views.sh mock

## coverage: Generate HTML coverage report
coverage:
	go test -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code with gofumpt
fmt:
	gofumpt -w .

## vet: Run go vet
vet:
	go vet ./...

## ci: Full CI pipeline (lint + test + build)
ci: lint test build

## dev: Run with race detector and debug flag
dev:
	go run -race ./cmd/w9s --debug

## mock: Run in mock mode with sample data
mock:
	W9S_ENABLE_MOCK=1 go run ./cmd/w9s --mock

## help: Show this help message
help:
	@echo "w9s Makefile targets:"
	@echo ""
	@echo "  build            Build the binary"
	@echo "  clean            Remove build artifacts"
	@echo "  test             Run all tests"
	@echo "  test-unit        Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo "  test-smoke       Run smoke tests (mock mode, needs tmux)"
	@echo "  coverage         Generate HTML coverage report"
	@echo "  lint             Run linter"
	@echo "  fmt              Format code"
	@echo "  vet              Run go vet"
	@echo "  ci               Full CI (lint + test + build)"
	@echo "  dev              Run with race detector"
	@echo "  mock             Run in mock mode"
	@echo ""
