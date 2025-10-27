# Collector Configuration

This guide covers configuring the OpenTelemetry Collector with the ProfileToMetrics connector.

## Basic Collector Setup

### Minimal Configuration

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
      memory:
        enabled: true

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      profiles: [otlp]
      exporters: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      exporters: [debug]
```

## Receivers

### OTLP Receiver

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
        max_recv_msg_size: 4194304
        max_concurrent_streams: 16
      http:
        endpoint: 0.0.0.0:4318
        max_request_body_size: 4194304
```

### Prometheus Receiver

```yaml
receivers:
  prometheus:
    config:
      scrape_configs:
        - job_name: 'otel-collector'
          scrape_interval: 10s
          static_configs:
            - targets: ['localhost:8888']
```

## Connectors

### ProfileToMetrics Connector

```yaml
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
```

## Processors

### Batch Processor

```yaml
processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
    send_batch_max_size: 2048
```

### Resource Processor

```yaml
processors:
  resource:
    attributes:
      - key: "deployment.environment"
        value: "production"
        action: "upsert"
      - key: "service.name"
        from_attribute: "service.name"
        action: "upsert"
```

### Filter Processor

```yaml
processors:
  filter:
    traces:
      span:
        - 'attributes["http.method"] == "GET"'
    metrics:
      metric:
        - 'name == "cpu_time"'
```

### Transform Processor

```yaml
processors:
  transform:
    trace_statements:
      - context: "span"
        statements:
          - set(attributes["processed"], true)
    metric_statements:
      - context: "metric"
        statements:
          - set(attributes["source"], "profiletometrics")
```

## Exporters

### Debug Exporter

```yaml
exporters:
  debug:
    verbosity: detailed
    sampling_initial: 2
    sampling_thereafter: 500
```

### OTLP Exporter

```yaml
exporters:
  otlp:
    endpoint: "http://observability-platform:4317"
    tls:
      insecure: true
    compression: gzip
    timeout: 30s
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 300s
```



## Service Configuration

### Pipelines

```yaml
service:
  pipelines:
    profiles:
      receivers: [otlp]
       exporters: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      processors: [batch, resource, transform]
      exporters: [debug, otlp]
    logs:
      receivers: [otlp]
      processors: [batch, resource]
      exporters: [debug, otlp]
```

### Telemetry

```yaml
service:
  telemetry:
    logs:
      level: debug
      development: true
    metrics:
      level: detailed
      address: 0.0.0.0:8888
    traces:
      level: detailed
      address: 0.0.0.0:8888
```

## Complete Configuration Example

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

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

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  resource:
    attributes:
      - key: "deployment.environment"
        value: "production"
        action: "upsert"
  filter:
    metrics:
      metric:
        - 'name == "cpu_time"'
        - 'name == "memory_allocation"'

exporters:
  debug:
    verbosity: detailed
  otlp:
    endpoint: "http://observability-platform:4317"
    tls:
      insecure: true
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: "otel"

service:
  pipelines:
    profiles:
      receivers: [otlp]
      connectors: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      processors: [batch, resource, filter]
      exporters: [debug, otlp]
  
  telemetry:
    logs:
      level: debug
      development: true
    metrics:
      level: detailed
      address: 0.0.0.0:8888
```

## Environment-Specific Configurations

### Development

```yaml
service:
  telemetry:
    logs:
      level: debug

exporters:
  debug:
    verbosity: detailed
```

### Production

```yaml
service:
  telemetry:
    logs:
      level: info

exporters:
  otlp:
    endpoint: "http://observability-platform:4317"
    tls:
      insecure: false
    compression: gzip
```

### Kubernetes

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

connectors:
  profiletometrics:
    attributes:
      - key: "k8s.namespace"
        value: "default"
      - key: "k8s.pod.name"
        value: "pod-.*"
      - key: "k8s.container.name"
        value: "container-.*"

processors:
  k8sattributes:
    auth_type: "serviceAccount"
    passthrough: false
    filter:
      node_from_env_var: "KUBE_NODE_NAME"
```

## Configuration Validation

### Command Line Validation

```bash
# Validate configuration
otelcol --config config.yaml --dry-run

# Check configuration syntax
otelcol --config config.yaml --check-config
```


## Troubleshooting

### Debug Configuration

```yaml
service:
  telemetry:
    logs:
      level: debug
      development: true
    metrics:
      level: detailed
      address: 0.0.0.0:8888
```

### Health Checks

```bash
# Check collector health
curl http://localhost:8888/

# Check metrics
curl http://localhost:8888/metrics

# Check configuration
curl http://localhost:8888/config
```

### Log Analysis

```bash
# Check logs for errors
docker logs <container-id> | grep ERROR

# Check debug logs
docker logs <container-id> | grep DEBUG

# Check connector logs
docker logs <container-id> | grep profiletometrics
```
