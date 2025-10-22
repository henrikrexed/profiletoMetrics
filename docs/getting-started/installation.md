# Installation

This guide covers different ways to install and use the ProfileToMetrics Connector.

## Prerequisites

- **OpenTelemetry Collector**: Version 0.137.0 or later
- **Go**: Version 1.23.0 or later (for building from source)
- **Docker**: For containerized deployment
- **Kubernetes**: For K8s deployment (optional)

## Installation Methods

### 1. Docker (Recommended)

The easiest way to get started is using the pre-built Docker image.

#### Pull the Image

```bash
docker pull hrexed/otel-collector-profilemetrics:latest
```

#### Run with Configuration

```bash
docker run -p 4317:4317 -p 8888:8888 \
  -v $(pwd)/config.yaml:/etc/otelcol/config.yaml \
  hrexed/otel-collector-profilemetrics:latest
```

#### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_CONFIG` | `/etc/otelcol/config.yaml` | Path to configuration file |
| `OTEL_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### 2. Kubernetes

Deploy using the provided Kubernetes manifests.

#### Quick Deploy

```bash
# Apply all manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

#### Verify Deployment

```bash
# Check pods
kubectl get pods -n otel-collector

# Check logs
kubectl logs -n otel-collector deployment/otel-collector

# Check service
kubectl get svc -n otel-collector
```

### 3. Build from Source

#### Prerequisites

- Go 1.23.0+
- Git
- Make

#### Build Steps

```bash
# Clone the repository
git clone https://github.com/henrikrexed/profiletoMetrics.git
cd profiletoMetrics

# Install dependencies
go mod tidy

# Run tests
make test

# Build the collector
make build

# Build Docker image
make docker-build
```

#### Using Makefile

The project includes a comprehensive Makefile with the following targets:

```bash
# Build the collector
make build

# Run tests
make test

# Build Docker image
make docker-build

# Install OCB (OpenTelemetry Collector Builder)
make install-ocb

# Clean build artifacts
make clean
```

### 4. OpenTelemetry Collector Builder (OCB)

Build a custom collector with the ProfileToMetrics connector.

#### Install OCB

```bash
# Install OCB
go install go.opentelemetry.io/collector/cmd/builder@v0.137.0

# Verify installation
builder --version
```

#### Build Custom Collector

```bash
# Build with OCB
ocb --config ocb/manifest.yaml
```

## Configuration

### Basic Configuration

Create a `config.yaml` file:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        metric_name: "cpu_time"
      memory:
        enabled: true
        metric_name: "memory_allocation"

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      connectors: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      exporters: [debug]
```

### Advanced Configuration

See the [Configuration Guide](configuration/connector-config.md) for detailed configuration options.

## Verification

### 1. Check Collector Status

```bash
# Health check endpoint
curl http://localhost:8888/

# Metrics endpoint
curl http://localhost:8888/metrics
```

### 2. Send Test Data

```bash
# Send test profiling data
./scripts/send-test-data.sh
```

### 3. Check Logs

```bash
# Docker logs
docker logs <container-id>

# Kubernetes logs
kubectl logs -n otel-collector deployment/otel-collector
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

```bash
# Check what's using the port
lsof -i :4317

# Kill the process
kill -9 <PID>
```

#### 2. Configuration Errors

```bash
# Validate configuration
otelcol --config config.yaml --dry-run
```

#### 3. Permission Issues (Kubernetes)

```bash
# Check RBAC
kubectl auth can-i create pods --as=system:serviceaccount:otel-collector:otel-collector
```

### Debug Mode

Enable debug logging for troubleshooting:

```yaml
service:
  telemetry:
    logs:
      level: debug
      development: true
```

## Next Steps

- [Quick Start Guide](quick-start.md)
- [Configuration Reference](configuration/connector-config.md)
- [Deployment Options](deployment/docker.md)
