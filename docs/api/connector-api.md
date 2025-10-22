# Connector API Reference

This document provides a comprehensive API reference for the ProfileToMetrics Connector.

## Connector Interface

### Factory

```go
type Factory struct {
    component.MustNewType
}

func (f *Factory) CreateDefaultConfig() component.Config
func (f *Factory) CreateConnector(
    ctx context.Context,
    params connector.CreateSettings,
    cfg component.Config,
    nextConsumer consumer.Metrics,
) (connector.Connector, error)
```

### Connector Implementation

```go
type profileToMetricsConnector struct {
    config       *Config
    nextConsumer consumer.Metrics
    logger       *zap.Logger
    converter    *profiletometrics.ConverterConnector
}
```

## Methods

### Start

```go
func (c *profileToMetricsConnector) Start(
    ctx context.Context,
    host component.Host,
) error
```

**Description**: Initializes the connector and starts processing.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `host`: OpenTelemetry Collector host interface

**Returns**:
- `error`: Any initialization error

**Example**:
```go
err := connector.Start(ctx, host)
if err != nil {
    log.Fatal("Failed to start connector:", err)
}
```

### Shutdown

```go
func (c *profileToMetricsConnector) Shutdown(
    ctx context.Context,
) error
```

**Description**: Gracefully shuts down the connector.

**Parameters**:
- `ctx`: Context for cancellation and timeout

**Returns**:
- `error`: Any shutdown error

**Example**:
```go
err := connector.Shutdown(ctx)
if err != nil {
    log.Printf("Error during shutdown: %v", err)
}
```

### Capabilities

```go
func (c *profileToMetricsConnector) Capabilities() consumer.Capabilities
```

**Description**: Returns the capabilities of the connector.

**Returns**:
- `consumer.Capabilities`: Connector capabilities

**Example**:
```go
caps := connector.Capabilities()
fmt.Printf("Mutates data: %v\n", caps.MutatesData)
```

### ConsumeTraces

```go
func (c *profileToMetricsConnector) ConsumeTraces(
    ctx context.Context,
    td ptrace.Traces,
) error
```

**Description**: Processes trace data and converts it to metrics.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `td`: Trace data to process

**Returns**:
- `error`: Any processing error

**Example**:
```go
err := connector.ConsumeTraces(ctx, traceData)
if err != nil {
    log.Printf("Failed to process traces: %v", err)
}
```

### ConsumeLogs

```go
func (c *profileToMetricsConnector) ConsumeLogs(
    ctx context.Context,
    ld plog.Logs,
) error
```

**Description**: Processes log data and converts it to metrics.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `ld`: Log data to process

**Returns**:
- `error`: Any processing error

**Example**:
```go
err := connector.ConsumeLogs(ctx, logData)
if err != nil {
    log.Printf("Failed to process logs: %v", err)
}
```

### ConsumeMetrics

```go
func (c *profileToMetricsConnector) ConsumeMetrics(
    ctx context.Context,
    md pmetric.Metrics,
) error
```

**Description**: Processes metric data (pass-through).

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `md`: Metric data to process

**Returns**:
- `error`: Any processing error

**Example**:
```go
err := connector.ConsumeMetrics(ctx, metricData)
if err != nil {
    log.Printf("Failed to process metrics: %v", err)
}
```

## Configuration API

### Config Structure

```go
type Config struct {
    Metrics       MetricsConfig       `mapstructure:"metrics"`
    Attributes    []AttributeConfig   `mapstructure:"attributes"`
    ProcessFilter ProcessFilterConfig `mapstructure:"process_filter"`
    ThreadFilter  ThreadFilterConfig  `mapstructure:"thread_filter"`
    PatternFilter PatternFilterConfig  `mapstructure:"pattern_filter"`
}
```

### MetricsConfig

```go
type MetricsConfig struct {
    CPU    CPUMetricsConfig    `mapstructure:"cpu"`
    Memory MemoryMetricsConfig `mapstructure:"memory"`
}
```

### CPUMetricsConfig

```go
type CPUMetricsConfig struct {
    Enabled     bool   `mapstructure:"enabled"`
    MetricName  string `mapstructure:"metric_name"`
    Description string `mapstructure:"description"`
    Unit        string `mapstructure:"unit"`
}
```

### MemoryMetricsConfig

```go
type MemoryMetricsConfig struct {
    Enabled     bool   `mapstructure:"enabled"`
    MetricName  string `mapstructure:"metric_name"`
    Description string `mapstructure:"description"`
    Unit        string `mapstructure:"unit"`
}
```

### AttributeConfig

```go
type AttributeConfig struct {
    Key   string `mapstructure:"key"`
    Value string `mapstructure:"value"`
}
```

### FilterConfig

```go
type ProcessFilterConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    Pattern string `mapstructure:"pattern"`
}

type ThreadFilterConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    Pattern string `mapstructure:"pattern"`
}

type PatternFilterConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    Pattern string `mapstructure:"pattern"`
}
```

## Converter API

### ConverterConnector

```go
type ConverterConnector struct {
    config ConverterConfig
    logger *zap.Logger
}
```

### NewConverterConnector

```go
func NewConverterConnector(config ConverterConfig) *ConverterConnector
```

**Description**: Creates a new converter connector instance.

**Parameters**:
- `config`: Converter configuration

**Returns**:
- `*ConverterConnector`: New converter instance

**Example**:
```go
config := ConverterConfig{
    Metrics: MetricsConfig{
        CPU: CPUMetricsConfig{Enabled: true},
        Memory: MemoryMetricsConfig{Enabled: true},
    },
}
converter := NewConverterConnector(config)
```

### ConvertTracesToMetrics

```go
func (c *ConverterConnector) ConvertTracesToMetrics(
    traces ptrace.Traces,
) (pmetric.Metrics, error)
```

**Description**: Converts trace data to metrics.

**Parameters**:
- `traces`: Trace data to convert

**Returns**:
- `pmetric.Metrics`: Generated metrics
- `error`: Any conversion error

**Example**:
```go
metrics, err := converter.ConvertTracesToMetrics(traces)
if err != nil {
    log.Printf("Conversion failed: %v", err)
    return
}
```

### ConvertLogsToMetrics

```go
func (c *ConverterConnector) ConvertLogsToMetrics(
    logs plog.Logs,
) (pmetric.Metrics, error)
```

**Description**: Converts log data to metrics.

**Parameters**:
- `logs`: Log data to convert

**Returns**:
- `pmetric.Metrics`: Generated metrics
- `error`: Any conversion error

**Example**:
```go
metrics, err := converter.ConvertLogsToMetrics(logs)
if err != nil {
    log.Printf("Conversion failed: %v", err)
    return
}
```

### ConvertProfilesToMetrics

```go
func (c *ConverterConnector) ConvertProfilesToMetrics(
    ctx context.Context,
    profiles pprofile.Profiles,
) (pmetric.Metrics, error)
```

**Description**: Converts profiling data to metrics.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `profiles`: Profiling data to convert

**Returns**:
- `pmetric.Metrics`: Generated metrics
- `error`: Any conversion error

**Example**:
```go
metrics, err := converter.ConvertProfilesToMetrics(ctx, profiles)
if err != nil {
    log.Printf("Conversion failed: %v", err)
    return
}
```

## Utility Functions

### CalculateCPUTime

```go
func CalculateCPUTime(samples []pprofile.Sample) float64
```

**Description**: Calculates total CPU time from samples.

**Parameters**:
- `samples`: Profiling samples

**Returns**:
- `float64`: Total CPU time in seconds

**Example**:
```go
cpuTime := CalculateCPUTime(samples)
fmt.Printf("CPU time: %.3f seconds\n", cpuTime)
```

### CalculateMemoryAllocation

```go
func CalculateMemoryAllocation(samples []pprofile.Sample) float64
```

**Description**: Calculates total memory allocation from samples.

**Parameters**:
- `samples`: Profiling samples

**Returns**:
- `float64`: Total memory allocation in bytes

**Example**:
```go
memoryAllocation := CalculateMemoryAllocation(samples)
fmt.Printf("Memory allocation: %.0f bytes\n", memoryAllocation)
```

### ExtractFromStringTable

```go
func ExtractFromStringTable(
    profile pprofile.Profile,
    attributes map[string]string,
) map[string]string
```

**Description**: Extracts attributes from profiling string table.

**Parameters**:
- `profile`: Profiling data
- `attributes`: Existing attributes map

**Returns**:
- `map[string]string`: Updated attributes map

**Example**:
```go
attributes := make(map[string]string)
attributes = ExtractFromStringTable(profile, attributes)
fmt.Printf("Extracted attributes: %v\n", attributes)
```

## Error Handling

### Error Types

```go
type ConversionError struct {
    Type    string
    Message string
    Cause   error
}

func (e *ConversionError) Error() string
func (e *ConversionError) Unwrap() error
```

### Common Errors

```go
var (
    ErrInvalidConfiguration = &ConversionError{
        Type:    "ConfigurationError",
        Message: "invalid configuration",
    }
    
    ErrInvalidProfilingData = &ConversionError{
        Type:    "DataError",
        Message: "invalid profiling data",
    }
    
    ErrMetricGeneration = &ConversionError{
        Type:    "MetricError",
        Message: "failed to generate metrics",
    }
)
```

### Error Handling Example

```go
metrics, err := converter.ConvertProfilesToMetrics(ctx, profiles)
if err != nil {
    var convErr *ConversionError
    if errors.As(err, &convErr) {
        switch convErr.Type {
        case "ConfigurationError":
            log.Printf("Configuration error: %v", convErr.Message)
        case "DataError":
            log.Printf("Data error: %v", convErr.Message)
        case "MetricError":
            log.Printf("Metric generation error: %v", convErr.Message)
        default:
            log.Printf("Unknown error: %v", convErr.Message)
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

## Logging

### Log Levels

- `DEBUG`: Detailed debugging information
- `INFO`: General information
- `WARN`: Warning messages
- `ERROR`: Error messages

### Log Fields

```go
c.logger.Debug("Processing traces",
    zap.Int("resource_spans_count", resourceSpansCount),
    zap.Int("total_spans", totalSpans),
)

c.logger.Error("Failed to convert traces to metrics",
    zap.Error(err),
    zap.Int("input_spans", totalSpans),
)
```

### Log Configuration

```yaml
service:
  telemetry:
    logs:
      level: debug
      development: true
```

## Performance

### Metrics

- **Throughput**: Number of profiles processed per second
- **Latency**: Time to process a single profile
- **Memory Usage**: Memory consumption during processing
- **CPU Usage**: CPU consumption during processing

### Optimization

- **Batch Processing**: Process multiple profiles in batches
- **Caching**: Cache frequently accessed data
- **Parallel Processing**: Use goroutines for concurrent processing
- **Memory Pooling**: Reuse memory allocations

### Monitoring

```go
// Performance metrics
c.logger.Info("Performance metrics",
    zap.Duration("processing_time", processingTime),
    zap.Int("profiles_processed", profilesProcessed),
    zap.Float64("throughput", throughput),
)
```

## Testing

### Unit Tests

```go
func TestConvertProfilesToMetrics(t *testing.T) {
    // Test implementation
}
```

### Integration Tests

```go
func TestIntegration(t *testing.T) {
    // Integration test implementation
}
```

### Benchmark Tests

```go
func BenchmarkConvertProfilesToMetrics(b *testing.B) {
    // Benchmark implementation
}
```

## Examples

### Basic Usage

```go
// Create configuration
config := Config{
    Metrics: MetricsConfig{
        CPU: CPUMetricsConfig{
            Enabled: true,
            MetricName: "cpu_time",
        },
        Memory: MemoryMetricsConfig{
            Enabled: true,
            MetricName: "memory_allocation",
        },
    },
}

// Create converter
converter := NewConverterConnector(ConverterConfig{
    Metrics: config.Metrics,
})

// Convert profiles
metrics, err := converter.ConvertProfilesToMetrics(ctx, profiles)
if err != nil {
    log.Fatal(err)
}

// Use metrics
fmt.Printf("Generated %d metrics\n", metrics.ResourceMetrics().Len())
```

### Advanced Usage

```go
// Advanced configuration
config := Config{
    Metrics: MetricsConfig{
        CPU: CPUMetricsConfig{
            Enabled: true,
            MetricName: "application_cpu_time",
            Description: "Application CPU time",
            Unit: "s",
        },
        Memory: MemoryMetricsConfig{
            Enabled: true,
            MetricName: "application_memory_allocation",
            Description: "Application memory allocation",
            Unit: "bytes",
        },
    },
    Attributes: []AttributeConfig{
        {Key: "service.name", Value: "my-service"},
        {Key: "environment", Value: "production"},
    },
    ProcessFilter: ProcessFilterConfig{
        Enabled: true,
        Pattern: "my-app.*",
    },
    ThreadFilter: ThreadFilterConfig{
        Enabled: true,
        Pattern: "worker-.*",
    },
}

// Create converter with advanced configuration
converter := NewConverterConnector(ConverterConfig{
    Metrics: config.Metrics,
    Attributes: config.Attributes,
    ProcessFilter: config.ProcessFilter,
    ThreadFilter: config.ThreadFilter,
})

// Convert with error handling
metrics, err := converter.ConvertProfilesToMetrics(ctx, profiles)
if err != nil {
    log.Printf("Conversion failed: %v", err)
    return
}

// Process metrics
resourceMetrics := metrics.ResourceMetrics()
for i := 0; i < resourceMetrics.Len(); i++ {
    resourceMetric := resourceMetrics.At(i)
    scopeMetrics := resourceMetric.ScopeMetrics()
    
    for j := 0; j < scopeMetrics.Len(); j++ {
        scopeMetric := scopeMetrics.At(j)
        metrics := scopeMetric.Metrics()
        
        for k := 0; k < metrics.Len(); k++ {
            metric := metrics.At(k)
            fmt.Printf("Metric: %s\n", metric.Name())
        }
    }
}
```
