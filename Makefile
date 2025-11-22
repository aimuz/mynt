.PHONY: help test test-unit test-integration test-docker coverage build run clean

# Default target
help:
	@echo "Mynt NAS - Makefile Commands"
	@echo ""
	@echo "Testing:"
	@echo "  make test            - Run all tests"
	@echo "  make test-unit       - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-docker     - Run Docker-based system tests"
	@echo "  make coverage        - Generate test coverage report"
	@echo ""
	@echo "Building:"
	@echo "  make build           - Build the binary"
	@echo "  make run             - Run the application"
	@echo "  make clean           - Clean build artifacts"

# Run all tests
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v -race -timeout 30s ./...

# Run unit tests with coverage
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -race -timeout 60s ./tests/integration/...

# Run Docker tests
test-docker:
	@echo "Running Docker tests..."
	cd tests/docker && ./test.sh

# Quick test (no race detector)
test-fast:
	@echo "Running fast tests..."
	go test -timeout 15s ./...

# Build the binary
build:
	@echo "Building myntd..."
	go build -o bin/myntd ./cmd/myntd

# Run the application
run:
	@echo "Running myntd..."
	go run ./cmd/myntd

# Run integration tests (local)
test-integration:
	@echo "Running integration tests..."
	INTEGRATION_TESTS=1 go test -v -race -timeout 60s ./tests/integration/...

# Run all tests including integration
test-all: test test-integration

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f mynt.db
	rm -rf cmd/myntd/mynt.db

# Install development dependencies
deps:
	@echo "Installing dependencies..."
	go get github.com/stretchr/testify/assert
	go get github.com/stretchr/testify/require
	go get github.com/stretchr/testify/mock
	go mod tidy
