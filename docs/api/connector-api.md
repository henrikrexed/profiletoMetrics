# Connector API Reference

This document provides a comprehensive reference for the ProfileToMetrics Connector configuration and usage.

## Overview

The ProfileToMetrics connector converts OpenTelemetry profiling data into metrics. It extracts attributes from the profile's string table and applies configurable filters to generate CPU and memory metrics.

## Configuration Reference

### Basic Configuration

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        metric_name: "cpu_time"
        description: "CPU time from profiling data"
        unit: "ns"
      memory:
        enabled: true
        metric_name: "memory_allocation"
        description: "Memory allocation from profiling data"
        unit: "bytes"
    attributes:
      - name: "service.name"
        value: "my-service"
        type: "literal"
      - name: "process.name"
        value: "main"
        type: "literal"
      - name: "function.name"
        value: ".*"
        type: "regex"
```

### Advanced Configuration

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        metric_name: "application_cpu_time"
        description: "Application CPU time"
        unit: "s"
      memory:
        enabled: true
        metric_name: "application_memory_allocation"
        description: "Application memory allocation"
        unit: "bytes"
    attributes:
      - name: "service.name"
        value: "my-service"
        type: "literal"
      - name: "environment"
        value: "production"
        type: "literal"
      - name: "function.name"
        value: ".*"
        type: "regex"
      - name: "thread.name"
        value: "worker-.*"
        type: "regex"
    process_filter:
      enabled: true
      pattern: "my-app.*"
    thread_filter:
      enabled: true
      pattern: "worker-.*"
    pattern_filter:
      enabled: true
      pattern: "production.*"
```

## Configuration Options

### Metrics Configuration

#### CPU Metrics

- **enabled**: Enable/disable CPU metrics generation
- **metric_name**: Name of the generated metric (default: "cpu_time")
- **description**: Description of the metric
- **unit**: Unit of measurement (default: "ns")

#### Memory Metrics

- **enabled**: Enable/disable memory metrics generation
- **metric_name**: Name of the generated metric (default: "memory_allocation")
- **description**: Description of the metric
- **unit**: Unit of measurement (default: "bytes")

### Attribute Configuration

#### Attribute Types

- **literal**: Direct string value
- **regex**: Regular expression pattern matching
- **string_table**: Direct string table index access

#### Attribute Examples

```yaml
attributes:
  # Literal value
  - name: "service.name"
    value: "my-service"
    type: "literal"
  
  # Regex pattern
  - name: "function.name"
    value: ".*worker.*"
    type: "regex"
  
  # String table index
  - name: "thread.name"
    value: "0"
    type: "string_table"
```

### Filter Configuration

#### Process Filter

Filter profiles based on process names:

```yaml
process_filter:
  enabled: true
  pattern: "my-app.*"  # Regex pattern
```

#### Thread Filter

Filter profiles based on thread names:

```yaml
thread_filter:
  enabled: true
  pattern: "worker-.*"  # Regex pattern
```

#### Pattern Filter

Filter profiles based on attribute values:

```yaml
pattern_filter:
  enabled: true
  pattern: "production.*"  # Regex pattern
```

## Feature Gates

**⚠️ Important**: The ProfileToMetrics connector requires the `+service.profilesSupport` feature gate to be enabled:

```bash
# Command line
otelcol --feature-gates=+service.profilesSupport

# Docker
docker run --feature-gates=+service.profilesSupport otelcol

# Kubernetes (see deployment section)
```

## Pipeline Configuration

The connector works with the profiles pipeline to convert profiling data to metrics:

```yaml
service:
  pipelines:
    traces:
      receivers: [otlp]
      e: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      exporters: [debug, otlp]
```

## Generated Metrics

### CPU Metrics

- **Name**: Configurable (default: "cpu_time")
- **Type**: Cumulative sum metric
- **Unit**: Time unit (default: "ns")
- **Attributes**: Extracted from string table and resource attributes

### Memory Metrics

- **Name**: Configurable (default: "memory_allocation")
- **Type**: Cumulative sum metric
- **Unit**: Memory unit (default: "bytes")
- **Attributes**: Extracted from string table and resource attributes

## Error Handling

The connector provides comprehensive error handling:

- **Configuration Errors**: Invalid configuration settings
- **Data Errors**: Invalid or malformed profiling data
- **Metric Generation Errors**: Failures in metric creation
- **Filter Errors**: Invalid regex patterns in filters

## Logging

The connector provides structured logging:

- **Input Statistics**: Number of profiles and samples processed
- **Processing Status**: Debug information about conversion process
- **Output Statistics**: Number of metrics generated
- **Error Context**: Detailed error information for troubleshooting

## Performance

The connector is optimized for performance:

- **Attribute Caching**: Caches frequently accessed string table attributes
- **Pattern Caching**: Caches compiled regex patterns for filters
- **Batch Processing**: Efficient processing of multiple profiles
- **Memory Management**: Optimized memory usage for large datasets

## Examples

### Basic Usage

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
      memory:
        enabled: true
```

### Advanced Usage with Filtering

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        metric_name: "application_cpu_time"
        unit: "s"
      memory:
        enabled: true
        metric_name: "application_memory_allocation"
        unit: "bytes"
    attributes:
      - name: "service.name"
        value: "my-service"
        type: "literal"
      - name: "function.name"
        value: ".*"
        type: "regex"
    process_filter:
      enabled: true
      pattern: "my-app.*"
    thread_filter:
      enabled: true
      pattern: "worker-.*"
```

### Complete Collector Configuration

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
        unit: "ns"
      memory:
        enabled: true
        metric_name: "memory_allocation"
        unit: "bytes"
    attributes:
      - name: "service.name"
        value: "my-service"
        type: "literal"
    process_filter:
      enabled: true
      pattern: "my-app.*"

exporters:
  debug:
    verbosity: detailed
  otlp:
    endpoint: http://localhost:4318

service:
  pipelines:
    profiles:
      receivers: [otlp]
      processors: [batch]
      exporters: [profiletometrics]
    metrics:
      connectors: [profiletometrics]
      processors: [batch]
      exporters: [debug, otlp]
```