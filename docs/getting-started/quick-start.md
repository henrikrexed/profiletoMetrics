# Quick Start

Get up and running with the ProfileToMetrics Connector in minutes.

## Prerequisites

- Docker installed
- Basic understanding of OpenTelemetry

## Step 1: Create Configuration

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
    attributes:
      - key: "service.name"
        value: "my-service"
    process_filter:
      enabled: true
      pattern: "my-app.*"

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    profiles:
      receivers: [otlp]
      processors: [batch]
      exporters: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      processors: [batch]
      exporters: [debug]
```

## Step 2: Run the Collector

```bash
# Pull the image
docker pull hrexed/otel-collector-profilemetrics:latest

# Run the collector with profiles feature gate enabled
docker run -p 4317:4317 -p 8888:8888 \
  --feature-gates=+service.profilesSupport \
  -v $(pwd)/config.yaml:/etc/otelcol/config.yaml \
  hrexed/otel-collector-profilemetrics:latest
```

**⚠️ Important**: The `+service.profilesSupport` feature gate must be enabled to use the profiles pipeline.

## Step 3: Send Test Data

In another terminal, send test profiling data:

```bash
# Clone the repository
git clone https://github.com/henrikrexed/profiletoMetrics.git
cd profiletoMetrics

# Send test data
./scripts/send-test-data.sh
```

## Step 4: Verify Output

You should see debug output showing:

```
2024-01-15T10:30:00.000Z	info	ProfileToMetrics connector started
2024-01-15T10:30:00.000Z	debug	Processing profiles	{"resource_profiles_count": 1, "total_profiles": 1}
2024-01-15T10:30:00.000Z	debug	Profiles converted to metrics	{"input_profiles": 1, "output_metrics": 2}
```

## Step 5: Check Metrics

Visit the metrics endpoint:

```bash
curl http://localhost:8888/metrics
```

You should see metrics like:

```
# HELP cpu_time CPU time in seconds
# TYPE cpu_time counter
cpu_time{service_name="my-service"} 0.123

# HELP memory_allocation Memory allocation in bytes
# TYPE memory_allocation counter
memory_allocation{service_name="my-service"} 1024
```

## Advanced Configuration

### Enable Debug Logging

```yaml
service:
  telemetry:
    logs:
      level: debug
      development: true
```

### Add More Attributes

```yaml
connectors:
  profiletometrics:
    attributes:
      - key: "service.name"
        value: "my-service"
      - key: "environment"
        value: "production"
      - key: "version"
        value: "1.0.0"
```

### Filter by Process

```yaml
connectors:
  profiletometrics:
    process_filter:
      enabled: true
      pattern: "my-app.*"
    thread_filter:
      enabled: true
      pattern: "worker-.*"
```

## Kubernetes Quick Start

### 1. Apply Manifests

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### 2. Check Status

```bash
kubectl get pods -n otel-collector
kubectl logs -n otel-collector deployment/otel-collector
```

### 3. Port Forward

```bash
kubectl port-forward -n otel-collector svc/otel-collector 4317:4317 8888:8888
```

## Troubleshooting

### Common Issues

#### 1. Connection Refused

```bash
# Check if collector is running
docker ps | grep otel-collector

# Check logs
docker logs <container-id>
```

#### 2. Configuration Errors

```bash
# Validate configuration
docker run --rm -v $(pwd)/config.yaml:/etc/otelcol/config.yaml \
  hrexed/otel-collector-profilemetrics:latest --config=/etc/otelcol/config.yaml --dry-run
```

#### 3. No Metrics Generated

- Check if profiling data is being sent
- Verify connector configuration
- Enable debug logging

### Debug Commands

```bash
# Check collector health
curl http://localhost:8888/

# Check metrics
curl http://localhost:8888/metrics

# Check logs with debug level
docker run -e OTEL_LOG_LEVEL=debug \
  -v $(pwd)/config.yaml:/etc/otelcol/config.yaml \
  hrexed/otel-collector-profilemetrics:latest
```

## Next Steps

- [Configuration Guide](configuration/connector-config.md)
- [Docker Deployment](deployment/docker.md)
- [Kubernetes Deployment](deployment/kubernetes.md)
- [Testing Guide](testing/unit-tests.md)
