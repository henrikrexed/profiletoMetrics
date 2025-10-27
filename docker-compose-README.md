# Docker Compose Deployment

This directory contains Docker Compose files for easy deployment of the OpenTelemetry Profile-to-Metrics Connector.

## Files

- `docker-compose.yml` - Basic development setup
- `docker-compose.prod.yml` - Production-ready setup with Prometheus

## Quick Start

### Development Setup

```bash
# Start the collector with basic configuration
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the services
docker-compose down
```

### Production Setup

```bash
# Start the production stack (collector + Prometheus)
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Stop the services
docker-compose -f docker-compose.prod.yml down
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_SERVICE_NAME` | `otel-collector-profilemetrics` | Service name |
| `OTEL_SERVICE_VERSION` | `0.1.0` | Service version |
| `OTEL_LOG_LEVEL` | `info` | Log level |

### Volumes

- `./examples/simple-config.yaml:/etc/otelcol/config.yaml:ro` - Configuration file
- `otel-data:/var/lib/otelcol` - Persistent data storage
- `otel-tmp:/tmp/otelcol` - Temporary files

### Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 4317 | gRPC | OTLP gRPC receiver |
| 4318 | HTTP | OTLP HTTP receiver |
| 8888 | HTTP | Metrics endpoint |
| 8889 | HTTP | Prometheus exporter |

## Health Checks

The container includes built-in health checks that verify the collector is responding:

```bash
# Check container health
docker-compose ps

# Manual health check
curl http://localhost:8888/metrics
```

## Monitoring

### Metrics Endpoint

Access collector metrics at: http://localhost:8888/metrics

### Prometheus (Production Setup)

Access Prometheus at: http://localhost:9090

## Troubleshooting

### View Logs

```bash
# All services
docker-compose logs

# Specific service
docker-compose logs otel-collector

# Follow logs
docker-compose logs -f otel-collector
```

### Restart Services

```bash
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart otel-collector
```

### Clean Up

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (WARNING: This will delete data)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## Custom Configuration

To use a custom configuration file:

1. Copy your config file to the project root
2. Update the volume mapping in `docker-compose.yml`:

```yaml
volumes:
  - ./your-config.yaml:/etc/otelcol/config.yaml:ro
```

3. Restart the services:

```bash
docker-compose restart
```

## Resource Limits

The production setup includes resource limits:

- **CPU**: 4 cores limit, 2 cores reservation
- **Memory**: 8GB limit, 4GB reservation

Adjust these values in `docker-compose.prod.yml` based on your needs.
