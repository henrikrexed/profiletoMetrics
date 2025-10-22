# Multi-stage Dockerfile for OpenTelemetry Collector with Profile to Metrics Connector
# This Dockerfile creates a minimal, secure image for production use

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy all source code first (including go.mod, go.sum, and all source files)
COPY . .

# Download dependencies
RUN go mod download

# Build the collector using OCB
RUN go install go.opentelemetry.io/collector/cmd/builder@v0.137.0
RUN mv $(go env GOPATH)/bin/builder $(go env GOPATH)/bin/ocb

# Run OCB with the new manifest
RUN ocb --config ocb/manifest.yaml

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN adduser -D -s /bin/sh -u 10001 otelcol

# Create directories
RUN mkdir -p /etc/otelcol /var/lib/otelcol /tmp/otelcol \
    && chown -R otelcol:otelcol /etc/otelcol /var/lib/otelcol /tmp/otelcol

# Copy the collector binary from build stage
COPY --from=builder /app/dist/otelcol-profiletometrics /otelcol-profiletometrics

# Set ownership
RUN chown otelcol:otelcol /otelcol-profiletometrics \
    && chmod +x /otelcol-profiletometrics

# Switch to non-root user
USER otelcol

# Expose ports
# 4317: OTLP gRPC receiver
# 4318: OTLP HTTP receiver  
# 8888: Prometheus metrics endpoint
# 8889: Prometheus exporter endpoint
EXPOSE 4317 4318 8888 8889

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8888/metrics || exit 1

# Set working directory
WORKDIR /etc/otelcol

# Default configuration
COPY examples/simple-config.yaml /etc/otelcol/config.yaml

# Set entrypoint
ENTRYPOINT ["/otelcol-profiletometrics"]
CMD ["--config=/etc/otelcol/config.yaml"]

# Labels
LABEL maintainer="profile-to-metrics-team"
LABEL description="OpenTelemetry Collector with Profile to Metrics Connector"
LABEL version="0.1.0"
LABEL collector.version="0.137.0"
