# MailSentinel Test Harness Makefile

.PHONY: test test-unit test-integration test-e2e test-bench test-coverage test-all clean build help

# Default target
help:
	@echo "MailSentinel Test Harness"
	@echo ""
	@echo "Available targets:"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  test-bench     - Run benchmark tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean test artifacts"
	@echo ""

# Build the binary
build:
	@echo "Building MailSentinel..."
	@go build -o bin/mailsentinel ./cmd/mailsentinel
	@echo "Binary built: bin/mailsentinel"

# Run all tests
test: test-unit test-integration

# Run unit tests only (pkg/ and internal/ packages with _test.go files)
test-unit:
	@echo "Running unit tests..."
	@go test -v ./pkg/config ./pkg/types ./internal/... -run "Test"
	@go test -v ./test/simple_test.go

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test -v ./test/integration_test.go -tags=integration

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	@go test -v ./test/e2e_test.go -tags=e2e

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	@go test -v -bench=. -benchmem ./test/benchmark_test.go

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./pkg/... ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	@go test -v -race ./pkg/... ./internal/... ./test/...

# Run tests with verbose output
test-verbose:
	@echo "Running all tests with verbose output..."
	@go test -v -race -coverprofile=coverage.out ./pkg/... ./internal/... ./test/...

# Run all test suites
test-all: test-unit test-integration test-e2e test-bench

# Validate test data
test-validate:
	@echo "Validating test data..."
	@go run scripts/validate_testdata.go

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -rf tmp/
	@go clean -testcache

# Quick smoke test
smoke:
	@echo "Running smoke tests..."
	@go test -v -short ./pkg/config ./pkg/types

# Test with specific timeout
test-timeout:
	@echo "Running tests with 5 minute timeout..."
	@go test -v -timeout=5m ./...

# Continuous testing (requires entr or similar)
test-watch:
	@echo "Starting continuous testing..."
	@find . -name "*.go" | entr -c make test-unit
