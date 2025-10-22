#!/bin/bash

# Example build scripts for Profile to Metrics Connector
# This script demonstrates different ways to build the connector with various configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Profile to Metrics Connector - Build Examples${NC}"
echo "=================================================="

# Function to show current config
show_config() {
    echo -e "${BLUE}Current Configuration:${NC}"
    make config
    echo ""
}

# Example 1: Default build
echo -e "${YELLOW}Example 1: Default Build${NC}"
echo "Building with default settings..."
make config
make build-collector
echo -e "${GREEN}✓ Default build completed${NC}"
echo ""

# Example 2: Custom version and image name
echo -e "${YELLOW}Example 2: Custom Version and Image Name${NC}"
echo "Building with custom version and image name..."
make config VERSION=1.2.3 DOCKER_IMAGE=myregistry/profiletometrics
make build-collector VERSION=1.2.3 DOCKER_IMAGE=myregistry/profiletometrics
echo -e "${GREEN}✓ Custom build completed${NC}"
echo ""

# Example 3: Using Podman instead of Docker
echo -e "${YELLOW}Example 3: Using Podman${NC}"
echo "Building with Podman..."
make config DOCKER_BINARY=podman
make validate-config DOCKER_BINARY=podman
echo -e "${GREEN}✓ Podman configuration validated${NC}"
echo ""

# Example 4: ARM64 platform build
echo -e "${YELLOW}Example 4: ARM64 Platform Build${NC}"
echo "Building for ARM64 platform..."
make config DOCKER_PLATFORM=linux/arm64
make validate-config DOCKER_PLATFORM=linux/arm64
echo -e "${GREEN}✓ ARM64 configuration validated${NC}"
echo ""

# Example 5: Multi-platform build
echo -e "${YELLOW}Example 5: Multi-Platform Build${NC}"
echo "Building for multiple platforms..."
make config DOCKER_IMAGE=myregistry/profiletometrics VERSION=1.0.0
echo -e "${BLUE}Note: Multi-platform build requires Docker Buildx and will push to registry${NC}"
echo -e "${BLUE}Command: make docker-build-multi DOCKER_IMAGE=myregistry/profiletometrics VERSION=1.0.0${NC}"
echo ""

# Example 6: Production build with validation
echo -e "${YELLOW}Example 6: Production Build with Validation${NC}"
echo "Running production build pipeline..."
make validate-config
make test
make build-collector
echo -e "${GREEN}✓ Production build pipeline completed${NC}"
echo ""

# Show all available environment variables
echo -e "${YELLOW}Available Environment Variables:${NC}"
echo "VERSION          - Version of the connector (default: 0.1.0)"
echo "DOCKER_IMAGE     - Docker image name (default: hrexed/otel-collector-profilemetrics)"
echo "DOCKER_TAG       - Docker image tag (default: VERSION)"
echo "DOCKER_BINARY    - Docker binary to use: docker or podman (default: docker)"
echo "DOCKER_PLATFORM  - Target platform: linux/amd64 or linux/arm64 (default: linux/amd64)"
echo ""

# Show usage examples
echo -e "${YELLOW}Usage Examples:${NC}"
echo "# Build with custom version"
echo "make docker-build VERSION=2.0.0"
echo ""
echo "# Build with custom image name and registry"
echo "make docker-build DOCKER_IMAGE=myregistry/myimage VERSION=1.0.0"
echo ""
echo "# Build with Podman for ARM64"
echo "make docker-build DOCKER_BINARY=podman DOCKER_PLATFORM=linux/arm64"
echo ""
echo "# Multi-platform build and push"
echo "make docker-build-multi DOCKER_IMAGE=myregistry/myimage VERSION=1.0.0"
echo ""
echo "# Show current configuration"
echo "make config"
echo ""
echo "# Validate configuration"
echo "make validate-config"
echo ""

echo -e "${GREEN}Build examples completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Run tests: make test"
echo "2. Build collector: make build-collector"
echo "3. Build Docker image: make docker-build"
echo "4. Run locally: make run-simple"
