# Docker Deployment Guide

This guide covers how to build, configure, and deploy the OpenTelemetry Profile-to-Metrics Connector using Docker.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Building Docker Images](#building-docker-images)
- [Configuration Options](#configuration-options)
- [Running Containers](#running-containers)
- [Docker Compose](#docker-compose)
- [Monitoring Stack](#monitoring-stack)
- [Troubleshooting](#troubleshooting)

## 🛠 Prerequisites

### Required Software

- **Docker** or **Podman** (container runtime)
- **Make** (build automation)
- **Git** (version control)

### System Requirements

- **CPU**: 2+ cores recommended
- **Memory**: 4GB+ RAM recommended
- **Storage**: 2GB+ free space
- **Network**: Internet access for pulling base images

## 🏗 Building Docker Images

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd profiletoMetrics

# Build with default settings
make docker-build
```

### Advanced Configuration

The Makefile provides extensive configuration options:

```bash
# Show all available options
make help

# Configure build settings
make config DOCKER_BINARY=podman DOCKER_PLATFORM=linux/amd64

# Validate configuration
make validate-config
```

### Build Options

| Option | Default | Description |
|--------|---------|-------------|
| `DOCKER_BINARY` | `docker` | Container runtime (docker/podman) |
| `DOCKER_PLATFORM` | `linux/amd64` | Target platform |
| `DOCKER_IMAGE` | `hrexed/otel-collector-profilemetrics` | Image name |
| `DOCKER_TAG` | `0.1.0` | Image tag |
| `VERSION` | `0.1.0` | Project version |

### Platform-Specific Builds

#### x86_64 (Intel/AMD)
```bash
make docker-build DOCKER_BINARY=docker DOCKER_PLATFORM=linux/amd64
```

#### ARM64 (Apple Silicon, ARM servers)
```bash
make docker-build DOCKER_BINARY=docker DOCKER_PLATFORM=linux/arm64
```

#### Multi-Platform Builds
```bash
# Build for multiple platforms
make docker-build-multi

# Or manually with docker buildx
docker buildx build --platform linux/amd64,linux/arm64 \
  -t hrexed/otel-collector-profilemetrics:0.1.0 \
  -f docker/Dockerfile .
```

### Build Process Details

The Docker build process uses a multi-stage approach:

1. **Builder Stage**:
   - Uses `golang:1.24-alpine` as base
   - Installs OCB (OpenTelemetry Collector Builder)
   - Builds the collector binary using `ocb-simple.yaml`

2. **Runtime Stage**:
   - Uses `alpine:3.19` as base
   - Installs runtime dependencies
   - Creates non-root user (`otelcol`)
   - Copies the built binary
   - Sets up health checks and labels

## ⚙️ Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | OTLP exporter endpoint |
| `OTEL_SERVICE_NAME` | `profile-to-metrics` | Service name |
| `OTEL_LOG_LEVEL` | `info` | Log level |

### Volume Mounts

| Host Path | Container Path | Description |
|-----------|----------------|-------------|
| `/path/to/config.yaml` | `/etc/otelcol/config.yaml` | Custom configuration |
| `/path/to/data` | `/var/lib/otelcol` | Persistent data |
| `/path/to/logs` | `/tmp/otelcol` | Log files |

### Port Mappings

| Host Port | Container Port | Protocol | Description |
|-----------|----------------|----------|-------------|
| `4317` | `4317` | gRPC | OTLP gRPC receiver |
| `4318` | `4318` | HTTP | OTLP HTTP receiver |
| `8888` | `8888` | HTTP | Metrics endpoint |
| `8889` | `8889` | HTTP | Prometheus exporter |

## 🚀 Running Containers

### Basic Usage

```bash
# Run with default configuration
docker run -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0
```

### Advanced Usage

```bash
# Run with custom configuration
docker run -p 4317:4317 -p 8888:8888 \
  -v /path/to/config.yaml:/etc/otelcol/config.yaml \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
  hrexed/otel-collector-profilemetrics:0.1.0

# Run with persistent data
docker run -p 4317:4317 -p 8888:8888 \
  -v /path/to/data:/var/lib/otelcol \
  -v /path/to/logs:/tmp/otelcol \
  hrexed/otel-collector-profilemetrics:0.1.0

# Run in background
docker run -d --name otel-collector \
  -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0
```

### Podman Usage

```bash
# Build with Podman
make docker-build DOCKER_BINARY=podman

# Run with Podman
podman run -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0

# Run with rootless Podman
podman run --userns=keep-id \
  -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0
```

### Health Checks

The container includes built-in health checks:

```bash
# Check container health
docker ps
# Look for "healthy" status

# Manual health check
curl http://localhost:8888/metrics
```

## 🐳 Simple Docker Setup

The project provides a simple Docker setup focused on the OpenTelemetry Collector:

### Basic Container Usage

```bash
# Run the collector with default configuration
docker run -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0

# Run with custom configuration
docker run -p 4317:4317 -p 8888:8888 \
  -v /path/to/config.yaml:/etc/otelcol/config.yaml \
  hrexed/otel-collector-profilemetrics:0.1.0
```

### Accessing the Collector

| Endpoint | URL | Description |
|----------|-----|-------------|
| OTLP gRPC | `localhost:4317` | Receive telemetry data |
| OTLP HTTP | `localhost:4318` | Receive telemetry data via HTTP |
| Metrics | `http://localhost:8888/metrics` | Collector internal metrics |

## 🔧 Troubleshooting

### Common Issues

#### 1. Port Already in Use

```bash
# Check what's using the port
netstat -tulpn | grep :4317

# Kill the process
sudo kill -9 <PID>

# Or use different ports
docker run -p 4318:4317 -p 8889:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0
```

#### 2. Permission Denied

```bash
# Run with proper permissions
docker run --user 10001:10001 \
  -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0

# Or fix volume permissions
sudo chown -R 10001:10001 /path/to/data
```

#### 3. Configuration Errors

```bash
# Validate configuration
docker run --rm \
  hrexed/otel-collector-profilemetrics:0.1.0 \
  validate --config=/etc/otelcol/config.yaml

# Check logs
docker logs <container-name>
```

#### 4. Build Failures

```bash
# Clean build cache
docker system prune -a

# Build without cache
docker build --no-cache -t hrexed/otel-collector-profilemetrics:0.1.0 .

# Check build logs
docker build --progress=plain -t hrexed/otel-collector-profilemetrics:0.1.0 .
```

### Debugging

#### Enable Debug Logging

```bash
docker run -p 4317:4317 -p 8888:8888 \
  -e OTEL_LOG_LEVEL=debug \
  hrexed/otel-collector-profilemetrics:0.1.0
```

#### Access Container Shell

```bash
# Get shell access
docker exec -it <container-name> /bin/sh

# Check running processes
docker exec -it <container-name> ps aux

# Check file permissions
docker exec -it <container-name> ls -la /etc/otelcol/
```

#### Monitor Resource Usage

```bash
# Container stats
docker stats <container-name>

# Detailed resource usage
docker exec -it <container-name> top
```

### Performance Tuning

#### Memory Limits

```bash
docker run -p 4317:4317 -p 8888:8888 \
  --memory=2g --memory-swap=4g \
  hrexed/otel-collector-profilemetrics:0.1.0
```

#### CPU Limits

```bash
docker run -p 4317:4317 -p 8888:8888 \
  --cpus=2 \
  hrexed/otel-collector-profilemetrics:0.1.0
```

#### Network Optimization

```bash
docker run -p 4317:4317 -p 8888:8888 \
  --network=host \
  hrexed/otel-collector-profilemetrics:0.1.0
```

## 📚 Additional Resources

- [OpenTelemetry Collector Documentation](https://opentelemetry.io/docs/collector/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Prometheus Configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
- [Grafana Dashboard Creation](https://grafana.com/docs/grafana/latest/dashboards/)

## 🆘 Support

For issues related to Docker deployment:

1. Check the [troubleshooting section](#troubleshooting)
2. Review container logs: `docker logs <container-name>`
3. Validate configuration: `docker run --rm <image> validate`
4. Open an issue on GitHub with:
   - Docker version: `docker --version`
   - Container logs: `docker logs <container-name>`
   - Configuration file (if custom)
   - Steps to reproduce
