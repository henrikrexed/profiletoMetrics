#!/bin/bash

# Script to send test profile data to the OpenTelemetry Collector
# This script demonstrates how to send profiling data to the collector

set -e

# Configuration
COLLECTOR_HOST=${COLLECTOR_HOST:-localhost}
COLLECTOR_GRPC_PORT=${COLLECTOR_GRPC_PORT:-4317}
COLLECTOR_HTTP_PORT=${COLLECTOR_HTTP_PORT:-4318}
PROFILE_FILE=${PROFILE_FILE:-data-example/profile.log}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Profile to Metrics Connector - Test Data Sender${NC}"
echo "=================================================="

# Check if collector is running
echo -e "${YELLOW}Checking if collector is running...${NC}"
if ! curl -s http://${COLLECTOR_HOST}:8888/metrics > /dev/null; then
    echo -e "${RED}Error: Collector is not running on ${COLLECTOR_HOST}:8888${NC}"
    echo "Please start the collector first:"
    echo "  make run-simple"
    echo "  or"
    echo "  docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}Collector is running!${NC}"

# Check if profile file exists
if [ ! -f "$PROFILE_FILE" ]; then
    echo -e "${RED}Error: Profile file not found: $PROFILE_FILE${NC}"
    echo "Please ensure the profile file exists or set PROFILE_FILE environment variable"
    exit 1
fi

echo -e "${YELLOW}Sending profile data from: $PROFILE_FILE${NC}"

# Function to send profile data via gRPC
send_profile_grpc() {
    local file=$1
    echo -e "${YELLOW}Sending profile data via gRPC (${COLLECTOR_HOST}:${COLLECTOR_GRPC_PORT})...${NC}"
    
    # Use curl to send HTTP request (simulating gRPC)
    curl -X POST \
        -H "Content-Type: application/x-protobuf" \
        -H "Content-Encoding: gzip" \
        --data-binary @"$file" \
        "http://${COLLECTOR_HOST}:${COLLECTOR_HTTP_PORT}/v1/profiles" \
        --max-time 30 \
        --connect-timeout 10 \
        --retry 3 \
        --retry-delay 1 \
        --silent \
        --show-error \
        --fail-with-body
}

# Function to send profile data via HTTP
send_profile_http() {
    local file=$1
    echo -e "${YELLOW}Sending profile data via HTTP (${COLLECTOR_HOST}:${COLLECTOR_HTTP_PORT})...${NC}"
    
    curl -X POST \
        -H "Content-Type: application/x-protobuf" \
        -H "Content-Encoding: gzip" \
        --data-binary @"$file" \
        "http://${COLLECTOR_HOST}:${COLLECTOR_HTTP_PORT}/v1/profiles" \
        --max-time 30 \
        --connect-timeout 10 \
        --retry 3 \
        --retry-delay 1 \
        --silent \
        --show-error \
        --fail-with-body
}

# Function to check metrics endpoint
check_metrics() {
    echo -e "${YELLOW}Checking metrics endpoint...${NC}"
    
    # Wait a bit for processing
    sleep 2
    
    # Check if profile metrics are available
    if curl -s "http://${COLLECTOR_HOST}:8889/metrics" | grep -q "profile_cpu_time_seconds"; then
        echo -e "${GREEN}✓ Profile CPU time metrics found!${NC}"
    else
        echo -e "${YELLOW}⚠ Profile CPU time metrics not found yet${NC}"
    fi
    
    if curl -s "http://${COLLECTOR_HOST}:8889/metrics" | grep -q "profile_memory_allocation_bytes"; then
        echo -e "${GREEN}✓ Profile memory allocation metrics found!${NC}"
    else
        echo -e "${YELLOW}⚠ Profile memory allocation metrics not found yet${NC}"
    fi
}

# Main execution
echo -e "${YELLOW}Starting profile data transmission...${NC}"

# Try HTTP first (more reliable for testing)
if send_profile_http "$PROFILE_FILE"; then
    echo -e "${GREEN}✓ Profile data sent successfully via HTTP${NC}"
    check_metrics
else
    echo -e "${RED}✗ Failed to send profile data via HTTP${NC}"
    echo "Trying gRPC..."
    
    if send_profile_grpc "$PROFILE_FILE"; then
        echo -e "${GREEN}✓ Profile data sent successfully via gRPC${NC}"
        check_metrics
    else
        echo -e "${RED}✗ Failed to send profile data via gRPC${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${GREEN}Test completed successfully!${NC}"
echo ""
echo "You can now:"
echo "1. Check metrics at: http://${COLLECTOR_HOST}:8889/metrics"
echo "2. View Prometheus at: http://${COLLECTOR_HOST}:9090"
echo "3. View Grafana at: http://${COLLECTOR_HOST}:3000 (admin/admin)"
echo "4. View Jaeger at: http://${COLLECTOR_HOST}:16686"
