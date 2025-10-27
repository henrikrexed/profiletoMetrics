# Docker Deployment

Deploy the ProfileToMetrics Connector using Docker containers.

## Feature Gates

**⚠️ Important**: The ProfileToMetrics connector requires the `+service.profilesSupport` feature gate to be enabled:

```bash
# Docker run with feature gate
docker run --feature-gates=+service.profilesSupport hrexed/otel-collector-profilemetrics:0.1.0
```

## Quick Start

### Pull and Run

```bash
# Pull the latest image
docker pull hrexed/otel-collector-profilemetrics:0.1.0

# Run with basic configuration
docker run -p 4317:4317 -p 8888:8888 \
  --feature-gates=+service.profilesSupport \
  hrexed/otel-collector-profilemetrics:0.1.0
```

### With Configuration File

```bash
# Create configuration file
cat > config.yaml << EOF
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
      memory:
        enabled: true

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    profiles:
      receivers: [otlp]
      exporters: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      exporters: [debug]
EOF

# Run with configuration
docker run -p 4317:4317 -p 8888:8888 \
  --feature-gates=+service.profilesSupport \
  -v $(pwd)/config.yaml:/etc/otelcol/config.yaml \
  hrexed/otel-collector-profilemetrics:0.1.0
```

## Image Variants

### Latest (Recommended)

```bash
docker pull hrexed/otel-collector-profilemetrics:0.1.0
```

### Specific Version

```bash
docker pull hrexed/otel-collector-profilemetrics:0.1.0
```

### Multi-Platform

```bash
# AMD64
docker pull hrexed/otel-collector-profilemetrics:0.1.0-amd64

# ARM64
docker pull hrexed/otel-collector-profilemetrics:0.1.0-arm64
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_CONFIG` | `/etc/otelcol/config.yaml` | Path to configuration file |
| `OTEL_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `OTEL_LOG_FORMAT` | `json` | Log format (json, console) |

### Volume Mounts

```bash
# Configuration file
-v $(pwd)/config.yaml:/etc/otelcol/config.yaml

# Logs directory
-v $(pwd)/logs:/var/log/otelcol

# Data directory
-v $(pwd)/data:/var/lib/otelcol
```

### Port Mapping

| Container Port | Host Port | Description |
|------------------|----------|-------------|
| 4317 | 4317 | OTLP gRPC receiver |
| 4318 | 4318 | OTLP HTTP receiver |
| 8888 | 8888 | Health check and metrics |
| 8889 | 8889 | Prometheus metrics (if enabled) |

## Docker Compose

### Basic Setup

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    ports:
      - "4317:4317"
      - "4318:4318"
      - "8888:8888"
    volumes:
      - ./config.yaml:/etc/otelcol/config.yaml
    environment:
      - OTEL_LOG_LEVEL=debug
    restart: unless-stopped
```

### With Observability Platform

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    ports:
      - "4317:4317"
      - "4318:4318"
      - "8888:8888"
    volumes:
      - ./config.yaml:/etc/otelcol/config.yaml
    environment:
      - OTEL_LOG_LEVEL=info
    depends_on:
      - jaeger
      - prometheus
    restart: unless-stopped

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    restart: unless-stopped
```

## Production Deployment

### Resource Limits

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    ports:
      - "4317:4317"
      - "4318:4318"
      - "8888:8888"
    volumes:
      - ./config.yaml:/etc/otelcol/config.yaml
    environment:
      - OTEL_LOG_LEVEL=info
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
    restart: unless-stopped
```

### Health Checks

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    ports:
      - "4317:4317"
      - "4318:4318"
      - "8888:8888"
    volumes:
      - ./config.yaml:/etc/otelcol/config.yaml
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8888/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    restart: unless-stopped
```

### Logging Configuration

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    ports:
      - "4317:4317"
      - "4318:4318"
      - "8888:8888"
    volumes:
      - ./config.yaml:/etc/otelcol/config.yaml
    environment:
      - OTEL_LOG_LEVEL=info
      - OTEL_LOG_FORMAT=json
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    restart: unless-stopped
```

## Building Custom Images

### Build from Source

```bash
# Clone repository
git clone https://github.com/henrikrexed/profiletoMetrics.git
cd profiletoMetrics

# Build Docker image
make docker-build

# Build with specific version
make docker-build VERSION=0.1.0
```



## Monitoring

### Metrics Endpoint

```bash
# Prometheus metrics
curl http://localhost:8888/metrics

# Health check
curl http://localhost:8888/
```

### Log Monitoring

```bash
# Follow logs
docker logs -f <container-id>

# Filter error logs
docker logs <container-id> | grep ERROR

# Filter debug logs
docker logs <container-id> | grep DEBUG
```

## Security

### Non-Root User

The container runs as a non-root user (`otelcol`) for security.

### Network Security

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    ports:
      - "127.0.0.1:4317:4317"  # Bind to localhost only
      - "127.0.0.1:4318:4318"
      - "127.0.0.1:8888:8888"
    volumes:
      - ./config.yaml:/etc/otelcol/config.yaml
    restart: unless-stopped
```

### TLS Configuration

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
        tls:
          cert_file: /etc/otelcol/tls/server.crt
          key_file: /etc/otelcol/tls/server.key
```

## Performance Tuning

### Resource Allocation

```yaml
version: '3.8'

services:
  otel-collector:
    image: hrexed/otel-collector-profilemetrics:0.1.0
    deploy:
      resources:
        limits:
          cpus: '4.0'
          memory: 8G
        reservations:
          cpus: '2.0'
          memory: 4G
    environment:
      - GOMAXPROCS=4
    restart: unless-stopped
```

### Batch Processing

```yaml
processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
    send_batch_max_size: 2048
```
