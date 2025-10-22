# Usage Guide

This guide provides comprehensive instructions on how to use the OpenTelemetry Profile-to-Metrics Connector in various scenarios.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Configuration Guide](#configuration-guide)
- [Usage Examples](#usage-examples)
- [Integration Patterns](#integration-patterns)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### 1. Basic Setup

Create a `config.yaml` file with the ProfileToMetrics connector:

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
        name: "cpu_time_seconds"
        unit: "s"
      memory:
        enabled: true
        name: "memory_allocation_bytes"
        unit: "bytes"
    
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"

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
```

### 2. Docker Usage

```bash
# Run the collector with profiles feature gate enabled
docker run -p 4317:4317 -p 8888:8888 \
  --feature-gates=+service.profilesSupport \
  hrexed/otel-collector-profilemetrics:0.1.0

# Send profile data
curl -X POST http://localhost:4317/v1/profiles \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @profile.pb
```

**⚠️ Important**: The `+service.profilesSupport` feature gate must be enabled to use the profiles pipeline.

## ⚙️ Configuration Guide

### Feature Gates

The ProfileToMetrics connector requires the `+service.profilesSupport` feature gate to be enabled:

```bash
# Command line
otelcol --feature-gates=+service.profilesSupport

# Docker
docker run --feature-gates=+service.profilesSupport otelcol

# Kubernetes (see deployment section)
```

### Core Configuration Structure

The ProfileToMetrics connector supports comprehensive configuration through YAML:

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        name: "cpu_time_seconds"
        unit: "s"
      memory:
        enabled: true
        name: "memory_allocation_bytes"
        unit: "bytes"
    
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"
    
    process_filter:
      enabled: true
      pattern: "my-app.*"
    
    thread_filter:
      enabled: true
      pattern: "worker-.*"
```

### Metrics Configuration

#### CPU Time Metrics

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        name: "cpu_time_seconds"
        unit: "s"
```

#### Memory Allocation Metrics

```yaml
connectors:
  profiletometrics:
    metrics:
      memory:
        enabled: true
        name: "memory_allocation_bytes"
        unit: "bytes"
```

### Attribute Extraction

#### Literal Value Extraction

```yaml
connectors:
  profiletometrics:
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"
      - name: "environment"
        value: "production"
        type: "literal"
```

#### Regular Expression Extraction

```yaml
connectors:
  profiletometrics:
    attributes:
      - name: "pod_name"
        value: "pod-(.*)"
        type: "regex"
      - name: "version"
        value: "v(\\d+\\.\\d+\\.\\d+)"
        type: "regex"
```

### Filtering Configuration

#### Process Filtering

```yaml
connectors:
  profiletometrics:
    process_filter:
      enabled: true
      pattern: "my-app-.*"
```

#### Thread Filtering

```yaml
connectors:
  profiletometrics:
    thread_filter:
      enabled: true
      pattern: "worker-.*"
```

#### Complete Filtering Example

```yaml
connectors:
  profiletometrics:
    process_filter:
      enabled: true
      pattern: "my-app-.*"
    
    thread_filter:
      enabled: true
      pattern: "worker-.*"
```

## 📖 Usage Examples

### Example 1: Basic CPU and Memory Metrics

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
        name: "cpu_time_seconds"
        unit: "s"
      memory:
        enabled: true
        name: "memory_allocation_bytes"
        unit: "bytes"
    
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"

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
```

### Example 2: Advanced Filtering

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
        name: "cpu_time_seconds"
        unit: "s"
    
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"
      - name: "environment"
        value: "production"
        type: "literal"
    
    process_filter:
      enabled: true
      pattern: "my-service-.*"
    
    thread_filter:
      enabled: true
      pattern: "worker-.*"

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
```

### Example 3: Attribute Extraction

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
        name: "cpu_time_seconds"
        unit: "s"
    
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"
      - name: "pod_name"
        value: "pod-(.*)"
        type: "regex"
      - name: "version"
        value: "v(\\d+\\.\\d+\\.\\d+)"
        type: "regex"
      - name: "environment"
        value: "production"
        type: "literal"

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
```

## 🔗 Integration Patterns

### Pattern 1: Basic Collector Integration

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
        name: "cpu_time_seconds"
        unit: "s"
      memory:
        enabled: true
        name: "memory_allocation_bytes"
        unit: "bytes"

exporters:
  debug:
    verbosity: detailed
  otlp:
    endpoint: http://localhost:4317

service:
  pipelines:
    profiles:
      receivers: [otlp]
      exporters: [profiletometrics]
    
    metrics:
      receivers: [profiletometrics]
      exporters: [debug, otlp]
```

### Pattern 2: Production Collector Setup

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
  resource:
    attributes:
      - key: service.name
        value: profile-to-metrics
        action: upsert

connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        name: "application_cpu_time_seconds"
        unit: "s"
      memory:
        enabled: true
        name: "application_memory_allocation_bytes"
        unit: "bytes"
    
    attributes:
      - name: "service_name"
        value: "my-service"
        type: "literal"
      - name: "environment"
        value: "production"
        type: "literal"
    
    process_filter:
      enabled: true
      pattern: "my-app.*"

exporters:
  debug:
    verbosity: detailed
  otlp:
    endpoint: http://observability-platform:4317
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: "profile_metrics"

service:
  pipelines:
    profiles:
      receivers: [otlp]
      processors: [batch, resource]
      exporters: [profiletometrics]
    
    metrics:
      receivers: [profiletometrics]
      processors: [batch]
      exporters: [debug, otlp, prometheus]
```

### Pattern 3: Microservices Architecture

```go
// Profile processing service
type ProfileProcessor struct {
    converter *profiletometrics.Converter
    exporter  MetricsExporter
}

func (p *ProfileProcessor) ProcessProfile(profile pprofile.Profiles) error {
    metrics, err := p.converter.ConvertProfilesToMetrics(context.Background(), profile)
    if err != nil {
        return err
    }
    
    return p.exporter.Export(metrics)
}

// Metrics exporter interface
type MetricsExporter interface {
    Export(metrics pmetric.Metrics) error
}
```

### Pattern 4: Batch Processing

```go
// Batch profile processor
type BatchProcessor struct {
    converter *profiletometrics.Converter
    batchSize int
}

func (b *BatchProcessor) ProcessBatch(profiles []pprofile.Profiles) error {
    for _, profile := range profiles {
        metrics, err := b.converter.ConvertProfilesToMetrics(context.Background(), profile)
        if err != nil {
            return err
        }
        
        // Process metrics in batch
        b.processMetrics(metrics)
    }
    
    return nil
}
```

## 🏆 Best Practices

### 1. Configuration Management

```go
// Use environment variables for configuration
func loadConfig() profiletometrics.Config {
    return profiletometrics.Config{
        Metrics: profiletometrics.MetricsConfig{
            CPUTime: profiletometrics.MetricConfig{
                Enabled:     getEnvBool("CPU_METRICS_ENABLED", true),
                Name:        getEnvString("CPU_METRICS_NAME", "cpu_time_seconds"),
                Description: getEnvString("CPU_METRICS_DESC", "CPU time spent in seconds"),
            },
        },
        ProcessFilter: profiletometrics.ProcessFilterConfig{
            Enabled:            getEnvBool("PROCESS_FILTER_ENABLED", false),
            ProcessNamePattern: getEnvString("PROCESS_FILTER_PATTERN", ""),
        },
    }
}
```

### 2. Error Handling

```go
func processProfiles(profiles pprofile.Profiles) error {
    converter := profiletometrics.NewConverter(config)
    
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    if err != nil {
        log.Printf("Failed to convert profiles: %v", err)
        return err
    }
    
    // Validate metrics before processing
    if metrics.ResourceMetrics().Len() == 0 {
        log.Printf("No metrics generated from profiles")
        return nil
    }
    
    return exportMetrics(metrics)
}
```

### 3. Performance Optimization

```go
// Reuse converter instance
var globalConverter *profiletometrics.Converter

func init() {
    config := loadConfig()
    globalConverter = profiletometrics.NewConverter(config)
}

func processProfiles(profiles pprofile.Profiles) error {
    // Reuse the global converter
    metrics, err := globalConverter.ConvertProfilesToMetrics(context.Background(), profiles)
    // ... rest of processing
}
```

### 4. Monitoring and Observability

```go
// Add metrics for the converter itself
func processProfiles(profiles pprofile.Profiles) error {
    start := time.Now()
    
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    
    // Record processing time
    processingTime := time.Since(start)
    log.Printf("Profile conversion took %v", processingTime)
    
    // Record metrics count
    metricCount := 0
    resourceMetrics := metrics.ResourceMetrics()
    for i := 0; i < resourceMetrics.Len(); i++ {
        scopeMetrics := resourceMetrics.At(i).ScopeMetrics()
        for j := 0; j < scopeMetrics.Len(); j++ {
            metricCount += scopeMetrics.At(j).Metrics().Len()
        }
    }
    
    log.Printf("Generated %d metrics", metricCount)
    
    return err
}
```

## 🔧 Troubleshooting

### Common Issues

#### 1. No Metrics Generated

**Symptoms**: Converter returns empty metrics
**Causes**: 
- Profile data doesn't match filter criteria
- Invalid configuration
- Empty profile data

**Solutions**:
```go
// Enable debug logging
config := profiletometrics.Config{
    // ... your config
}

// Check profile data
if profiles.Len() == 0 {
    log.Printf("No profiles to process")
    return
}

// Validate configuration
if err := config.Validate(); err != nil {
    log.Printf("Invalid configuration: %v", err)
    return
}
```

#### 2. Filter Not Working

**Symptoms**: Metrics include filtered-out data
**Causes**:
- Invalid regex patterns
- Filter configuration not enabled
- Sample attributes don't match expected format

**Solutions**:
```go
// Test regex patterns
pattern := "my-service-.*"
regex, err := regexp.Compile(pattern)
if err != nil {
    log.Printf("Invalid regex pattern: %v", err)
    return
}

// Enable debug logging to see filter results
log.Printf("Testing filter pattern: %s", pattern)
```

#### 3. Attribute Extraction Issues

**Symptoms**: Missing or incorrect attributes
**Causes**:
- Invalid string table indices
- Regex patterns don't match
- String table is empty

**Solutions**:
```go
// Validate string table indices
if stringTableIndex >= stringTable.Len() {
    log.Printf("String table index %d out of range (length: %d)", 
        stringTableIndex, stringTable.Len())
    return
}

// Test regex patterns
matched, err := regexp.MatchString(pattern, testString)
if err != nil {
    log.Printf("Regex error: %v", err)
    return
}
```

### Debugging Tips

#### 1. Enable Verbose Logging

```go
// Add debug logging
log.Printf("Processing %d profiles", profiles.Len())
log.Printf("Configuration: %+v", config)
```

#### 2. Validate Input Data

```go
// Check profile structure
for i := 0; i < profiles.Len(); i++ {
    profile := profiles.At(i)
    log.Printf("Profile %d: %d samples, %d locations", 
        i, profile.Sample().Len(), profile.Location().Len())
}
```

#### 3. Test Individual Components

```go
// Test converter without filters
config := profiletometrics.Config{
    Metrics: profiletometrics.MetricsConfig{
        CPUTime: profiletometrics.MetricConfig{
            Enabled: true,
            Name:    "cpu_time_seconds",
        },
    },
    // No filters for testing
}

converter := profiletometrics.NewConverter(config)
```

## 📚 Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Go Profiling Guide](https://golang.org/pkg/runtime/pprof/)
- [Regular Expressions in Go](https://golang.org/pkg/regexp/)
- [Performance Best Practices](https://golang.org/doc/effective_go.html#performance)
