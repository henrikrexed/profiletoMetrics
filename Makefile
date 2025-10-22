# Makefile for Profile to Metrics Connector
# This Makefile provides commands to build, test, and create Docker images
# for the OpenTelemetry Collector with the profile-to-metrics connector
#
# Environment Variables:
#   VERSION          - Version of the connector (default: 0.1.0)
#   DOCKER_IMAGE     - Docker image name (default: hrexed/otel-collector-profilemetrics)
#   DOCKER_TAG       - Docker image tag (default: VERSION)
#   DOCKER_BINARY    - Docker binary to use: docker or podman (default: docker)
#   DOCKER_PLATFORM  - Target platform: linux/amd64 or linux/arm64 (default: linux/amd64)
#
# Examples:
#   make docker-build DOCKER_IMAGE=myregistry/myimage VERSION=1.0.0
#   make docker-build DOCKER_BINARY=podman DOCKER_PLATFORM=linux/arm64
#   make docker-build-multi DOCKER_IMAGE=myregistry/myimage VERSION=1.0.0

# Variables
PROJECT_NAME := profiletoMetrics
VERSION := $(or $(VERSION),0.1.0)
COLLECTOR_VERSION := 0.137.0
DOCKER_IMAGE := $(or $(DOCKER_IMAGE),hrexed/otel-collector-profilemetrics)
DOCKER_TAG := $(or $(DOCKER_TAG),$(VERSION))
DOCKER_BINARY := $(or $(DOCKER_BINARY),docker)
DOCKER_PLATFORM := $(or $(DOCKER_PLATFORM),linux/amd64)
GO_VERSION := 1.23
OCB_VERSION := 0.137.0

# Directories
DIST_DIR := ./dist
BUILD_DIR := ./build
DOCKER_DIR := ./docker

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Docker parameters
DOCKER := $(DOCKER_BINARY)
DOCKER_BUILD := $(DOCKER) build --platform $(DOCKER_PLATFORM)
DOCKER_PUSH := $(DOCKER) push
DOCKER_TAG_CMD := $(DOCKER) tag

# OCB parameters
OCB := ocb
OCB_BUILD := $(OCB) build

.PHONY: all build test clean docker-build docker-push help install-deps install-ocb

# Default target
all: clean install-deps test build

# Help target
help:
	@echo "Profile to Metrics Connector - Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  install-deps    - Install Go dependencies"
	@echo "  install-ocb     - Install OpenTelemetry Collector Builder (OCB)"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linters"
	@echo "  format         - Format Go code"
	@echo ""
	@echo "Building:"
	@echo "  build          - Build the connector library"
	@echo "  build-collector - Build the OpenTelemetry Collector with OCB"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-build-multi - Build multi-platform Docker image"
	@echo "  docker-push       - Push Docker image to registry"
	@echo "  docker-run        - Run Docker container locally"
	@echo ""
	@echo "Examples:"
	@echo "  run-example    - Run the collector with example configuration"
	@echo "  run-simple     - Run the collector with simple configuration"
	@echo ""
	@echo "Configuration:"
	@echo "  config         - Show current configuration"
	@echo "  validate-config - Validate configuration settings"

# Install dependencies
install-deps:
	@echo "Installing Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install OpenTelemetry Collector Builder (OCB)
install-ocb:
	@echo "Installing OpenTelemetry Collector Builder..."
	@if ! command -v ocb &> /dev/null; then \
		go install go.opentelemetry.io/collector/cmd/builder@v$(OCB_VERSION); \
		mv $(GOPATH)/bin/builder $(GOPATH)/bin/ocb; \
	fi
	@echo "OCB installed successfully"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format Go code
format:
	@echo "Formatting Go code..."
	$(GOCMD) fmt ./...
	@if command -v goimports &> /dev/null; then \
		goimports -w .; \
	fi

# Build the connector library
build:
	@echo "Building connector library..."
	$(GOBUILD) -v ./pkg/profiletometrics/...

# Build the OpenTelemetry Collector with OCB
build-collector: install-ocb
	@echo "Building OpenTelemetry Collector with OCB..."
	@mkdir -p $(DIST_DIR)
	$(OCB_BUILD) --config ocb.yaml
	@echo "Collector built successfully in $(DIST_DIR)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(DIST_DIR)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f $(PROJECT_NAME)

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@echo "  Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "  Platform: $(DOCKER_PLATFORM)"
	@echo "  Binary: $(DOCKER_BINARY)"
	$(DOCKER_BUILD) -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f docker/Dockerfile.simple .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Push Docker image
docker-push: docker-build
	@echo "Pushing Docker image..."
	$(DOCKER_PUSH) $(DOCKER_IMAGE):$(DOCKER_TAG)

# Build multi-platform Docker image
docker-build-multi: build-collector
	@echo "Building multi-platform Docker image..."
	@echo "  Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "  Platforms: linux/amd64,linux/arm64"
	@mkdir -p $(DOCKER_DIR)
	@cp $(DIST_DIR)/otelcol-profiletometrics $(DOCKER_DIR)/
	$(DOCKER) buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_IMAGE):$(DOCKER_TAG) --push $(DOCKER_DIR)
	@echo "Multi-platform Docker image built and pushed: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run Docker container locally
docker-run: docker-build
	@echo "Running Docker container..."
	@echo "  Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "  Binary: $(DOCKER_BINARY)"
	$(DOCKER) run -d \
		--name otelcol-profiletometrics \
		-p 4317:4317 \
		-p 4318:4318 \
		-p 8888:8888 \
		-p 8889:8889 \
		-v $(PWD)/examples/collector-config.yaml:/etc/otelcol/config.yaml \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Stop and remove Docker container
docker-stop:
	@echo "Stopping Docker container..."
	$(DOCKER) stop otelcol-profiletometrics || true
	$(DOCKER) rm otelcol-profiletometrics || true

# Run the collector with example configuration
run-example: build-collector
	@echo "Running collector with example configuration..."
	$(DIST_DIR)/otelcol-profiletometrics --config examples/collector-config.yaml

# Run the collector with simple configuration
run-simple: build-collector
	@echo "Running collector with simple configuration..."
	$(DIST_DIR)/otelcol-profiletometrics --config examples/simple-config.yaml

# Development targets
dev-setup: install-deps install-ocb
	@echo "Development environment setup complete"

dev-test: test-coverage lint
	@echo "Development tests complete"

# Release targets
release-check: test lint
	@echo "Release checks complete"

release-build: clean release-check build-collector docker-build
	@echo "Release build complete"

# Utility targets
show-config:
	@echo "Project Configuration:"
	@echo "  Project Name: $(PROJECT_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Collector Version: $(COLLECTOR_VERSION)"
	@echo "  Docker Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "  Docker Binary: $(DOCKER_BINARY)"
	@echo "  Docker Platform: $(DOCKER_PLATFORM)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  OCB Version: $(OCB_VERSION)"

# Check if required tools are installed
check-tools:
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
	@command -v $(DOCKER_BINARY) >/dev/null 2>&1 || { echo "$(DOCKER_BINARY) is required but not installed. Aborting." >&2; exit 1; }
	@echo "All required tools are installed"

# Install all dependencies and tools
install-all: check-tools install-deps install-ocb
	@echo "All dependencies installed successfully"

# Quick start - build and run with simple config
quick-start: install-all build-collector run-simple

# Full pipeline - test, build, and create Docker image
full-pipeline: clean test build-collector docker-build
	@echo "Full pipeline completed successfully"

# Validate configuration
validate-config:
	@echo "Validating configuration..."
	@if [ "$(DOCKER_BINARY)" != "docker" ] && [ "$(DOCKER_BINARY)" != "podman" ]; then \
		echo "Error: DOCKER_BINARY must be 'docker' or 'podman', got: $(DOCKER_BINARY)"; \
		exit 1; \
	fi
	@if [ "$(DOCKER_PLATFORM)" != "linux/amd64" ] && [ "$(DOCKER_PLATFORM)" != "linux/arm64" ]; then \
		echo "Error: DOCKER_PLATFORM must be 'linux/amd64' or 'linux/arm64', got: $(DOCKER_PLATFORM)"; \
		exit 1; \
	fi
	@echo "Configuration is valid"

# Show current configuration
config:
	@echo "Current Configuration:"
	@echo "====================="
	@echo "VERSION: $(VERSION)"
	@echo "DOCKER_IMAGE: $(DOCKER_IMAGE)"
	@echo "DOCKER_TAG: $(DOCKER_TAG)"
	@echo "DOCKER_BINARY: $(DOCKER_BINARY)"
	@echo "DOCKER_PLATFORM: $(DOCKER_PLATFORM)"
	@echo "COLLECTOR_VERSION: $(COLLECTOR_VERSION)"
	@echo "GO_VERSION: $(GO_VERSION)"
	@echo "OCB_VERSION: $(OCB_VERSION)"
