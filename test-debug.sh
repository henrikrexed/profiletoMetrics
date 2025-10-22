#!/bin/bash

# Test script to demonstrate debug logging in the ProfileToMetrics connector

echo "Starting OpenTelemetry Collector with debug logging..."

# Run the collector in the background
podman run -d --name otelcol-debug \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 8888:8888 \
  -p 8889:8889 \
  -v $(pwd)/examples/final-debug-config.yaml:/etc/otelcol/config.yaml \
  localhost/hrexed/otel-collector-profilemetrics:0.1.0 \
  --config=/etc/otelcol/config.yaml

echo "Collector started. Waiting for it to initialize..."
sleep 5

echo "Checking collector logs for debug output..."
podman logs otelcol-debug

echo "Sending test data to the collector..."

# Send a simple trace to the collector
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{
    "resourceSpans": [{
      "resource": {
        "attributes": [{
          "key": "service.name",
          "value": {"stringValue": "test-service"}
        }]
      },
      "scopeSpans": [{
        "scope": {"name": "test-scope"},
        "spans": [{
          "traceId": "12345678901234567890123456789012",
          "spanId": "1234567890123456",
          "name": "test-span",
          "kind": "SPAN_KIND_INTERNAL",
          "startTimeUnixNano": "1640995200000000000",
          "endTimeUnixNano": "1640995201000000000"
        }]
      }]
    }]
  }'

echo "Waiting for processing..."
sleep 3

echo "Checking logs after sending data..."
podman logs otelcol-debug

echo "Cleaning up..."
podman stop otelcol-debug
podman rm otelcol-debug

echo "Test completed!"
