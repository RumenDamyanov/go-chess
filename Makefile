# Makefile for go-chess

.PHONY: build test clean lint fmt vet run-cli run-server install-deps docker-build docker-run docker-stop docker-compose-up docker-compose-down docker-dev help

# Variables
BINARY_NAME=go-chess
BINARY_CLI=go-chess-cli
BINARY_SERVER=go-chess-server
BUILD_DIR=build
MAIN_PACKAGE=./examples/cli
CLI_PACKAGE=./examples/cli
SERVER_PACKAGE=./examples/api-server

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Default target
all: test build

# Build the main application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(MAIN_PACKAGE)

# Build CLI example
build-cli:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_CLI) -v $(CLI_PACKAGE)

# Build server example
build-server:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_SERVER) -v $(SERVER_PACKAGE)

# Build all examples
build-examples: build-cli build-server

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	$(GOTEST) -bench=. ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Format code
fmt:
	$(GOFMT) ./...

# Vet code
vet:
	$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Install dependencies
install-deps:
	$(GOMOD) download
	$(GOMOD) verify

# Update dependencies
update-deps:
	$(GOMOD) tidy
	$(GOCMD) get -u ./...

# Run CLI example
run-cli: build-cli
	./$(BUILD_DIR)/$(BINARY_CLI)

# Run server example
run-server: build-server
	./$(BUILD_DIR)/$(BINARY_SERVER)

# Docker build
docker-build:
	docker build -t go-chess .

# Docker run
docker-run: docker-build
	docker run -d --name go-chess -p 8080:8080 -e CHESS_HOST=0.0.0.0 go-chess

# Docker stop
docker-stop:
	docker stop go-chess || true
	docker rm go-chess || true

# Docker compose up
docker-compose-up:
	docker-compose up --build -d

# Docker compose down
docker-compose-down:
	docker-compose down

# Docker development environment
docker-dev:
	docker-compose up --build chess-server

# Generate documentation
docs:
	godoc -http=:6060

# Security audit
security:
	gosec ./...

# Static analysis
analyze: fmt vet lint security

# Full CI pipeline
ci: install-deps analyze test build

# Development setup
dev-setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/github-action-gosec@latest
	$(MAKE) install-deps

# Show help
help:
	@echo "Available targets:"
	@echo "  build               - Build the main application"
	@echo "  build-cli           - Build CLI example"
	@echo "  build-server        - Build API server example"
	@echo "  build-examples      - Build all examples"
	@echo "  test                - Run tests"
	@echo "  test-coverage       - Run tests with coverage"
	@echo "  bench               - Run benchmarks"
	@echo "  clean               - Clean build artifacts"
	@echo "  fmt                 - Format code"
	@echo "  vet                 - Vet code"
	@echo "  lint                - Run linter"
	@echo "  install-deps        - Install dependencies"
	@echo "  update-deps         - Update dependencies"
	@echo "  run-cli             - Run CLI example"
	@echo "  run-server          - Run API server example"
	@echo "  docker-build        - Build Docker image"
	@echo "  docker-run          - Run Docker container"
	@echo "  docker-stop         - Stop Docker container"
	@echo "  docker-compose-up   - Start with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  docker-dev          - Run development environment"
	@echo "  docs                - Generate documentation"
	@echo "  security            - Run security audit"
	@echo "  analyze             - Run static analysis"
	@echo "  ci                  - Run full CI pipeline"
	@echo "  dev-setup           - Setup development environment"
	@echo "  help                - Show this help"


