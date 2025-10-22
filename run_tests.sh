#!/bin/bash

# Test runner script for the Profile to Metrics Connector

echo "Running Profile to Metrics Connector Tests"
echo "=========================================="

# Run unit tests
echo "Running unit tests..."
go test -v ./internal/config/... -count=1
go test -v ./internal/connector/... -count=1

# Run integration tests
echo "Running integration tests..."
go test -v ./... -count=1

# Run tests with coverage
echo "Running tests with coverage..."
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "Test coverage report generated: coverage.html"
echo "All tests completed!"
