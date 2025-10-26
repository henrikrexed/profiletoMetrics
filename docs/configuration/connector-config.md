# Connector Configuration

The ProfileToMetrics connector supports comprehensive configuration options for metrics generation, attribute extraction, and filtering.

## Feature Gates

**⚠️ Important**: The ProfileToMetrics connector requires the `+service.profilesSupport` feature gate to be enabled:

```bash
# Command line
otelcol --feature-gates=+service.profilesSupport

# Docker
docker run --feature-gates=+service.profilesSupport otelcol

# Kubernetes (see deployment section)
```

## Basic Configuration

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
```

## Configuration Reference

### Metrics Configuration

#### CPU Metrics

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true                    # Enable CPU metrics
        metric_name: "cpu_time"         # Metric name
        description: "CPU time in seconds" # Metric description
        unit: "s"                       # Metric unit
```

#### Memory Metrics

```yaml
connectors:
  profiletometrics:
    metrics:
      memory:
        enabled: true                   # Enable memory metrics
        metric_name: "memory_allocation" # Metric name
        description: "Memory allocation in bytes" # Metric description
        unit: "bytes"                   # Metric unit
```

#### Function Metrics

Control whether to generate per-function metrics:

```yaml
connectors:
  profiletometrics:
    metrics:
      function:
        enabled: true                   # Enable function-level metrics (default: true)
```

**Function Metrics Behavior:**

When `function.enabled` is set to `true`, the connector generates metrics with **function names as attributes** rather than creating separate metrics for each function. This approach significantly reduces metric cardinality while still providing per-function visibility.

**Generated Metrics:**

1. **CPU Time Per Function**:
   - Metric: `cpu_time` (uses base metric name)
   - Attributes: `function.name="<function_name>"`
   - Example: `cpu_time{function.name="main"}`

2. **Memory Allocation Per Function**:
   - Metric: `memory_allocation` (uses base metric name)
   - Attributes: `function.name="<function_name>"`
   - Example: `memory_allocation{function.name="handler"}`

**Function Name Extraction:**

Function names are automatically extracted from profile stack traces:
- The connector traverses the profile's location, mapping, and function tables
- Each stack trace sample is analyzed to identify the called functions
- Function names are resolved from the profile's string table

**Cardinality Considerations:**

- **Low Cardinality**: All functions share the same base metrics (`cpu_time` and `memory_allocation`)
- The `function.name` attribute value determines the number of distinct time series
- A profile with 100 unique functions creates 200 time series (2 metrics × 100 functions)

**Benefits:**

- ✅ Reduced metric cardinality compared to per-function metrics
- ✅ Consistent metric structure for easier querying
- ✅ Automatic function discovery from stack traces
- ✅ Can be disabled to reduce cardinality if not needed

**Note**: Function-level metrics are automatically extracted from profile stack traces. When enabled, they can increase metric cardinality based on the number of unique functions in your profiles. Disable this feature if you don't need function-level visibility.

### Attribute Configuration

Extract attributes from the profiling data's string table.

#### Literal Values

```yaml
connectors:
  profiletometrics:
    attributes:
      - key: "service.name"
        value: "my-service"            # Literal value
      - key: "environment"
        value: "production"
```

#### Regular Expressions

```yaml
connectors:
  profiletometrics:
    attributes:
      - key: "service.name"
        value: "service-.*"           # Regex pattern
      - key: "version"
        value: "v[0-9]+\\.[0-9]+"      # Version pattern
```

### Filtering Configuration

#### Process Filtering

Filter metrics based on process names:

```yaml
connectors:
  profiletometrics:
    process_filter:
      enabled: true                     # Enable process filtering
      pattern: "my-app.*"              # Regex pattern for process names
```

#### Pattern Filtering

Filter metrics based on attribute patterns:

```yaml
connectors:
  profiletometrics:
    pattern_filter:
      enabled: true                    # Enable pattern filtering
      pattern: "service-.*"           # Regex pattern for attributes
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
    # Metrics configuration
    metrics:
      cpu:
        enabled: true
        metric_name: "cpu_time"
        description: "CPU time in seconds"
        unit: "s"
      memory:
        enabled: true
        metric_name: "memory_allocation"
        description: "Memory allocation in bytes"
        unit: "bytes"
      function:
        enabled: true                   # Enable function-level metrics
    
    # Attribute extraction
    attributes:
      - key: "service.name"
        value: "my-service"
      - key: "environment"
        value: "production"
      - key: "version"
        value: "v[0-9]+\\.[0-9]+"
      - key: "instance.id"
        value: "instance-.*"
    
    # Filtering
    process_filter:
      enabled: true
      pattern: "my-app.*"
    
    pattern_filter:
      enabled: true
      pattern: "service-.*"

exporters:
  debug:
    verbosity: detailed
  otlp:
    endpoint: "http://observability-platform:4317"

service:
  pipelines:
    traces:
      receivers: [otlp]
      connectors: [profiletometrics]
    metrics:
      receivers: [profiletometrics]
      exporters: [debug, otlp]
  
  telemetry:
    logs:
      level: debug
      development: true
```

## Configuration Validation

### Required Fields

- `metrics.cpu.enabled` or `metrics.memory.enabled` must be `true`
- At least one attribute must be configured
- Valid regex patterns for filters

### Optional Fields

- `metrics.cpu.metric_name` (default: "cpu_time")
- `metrics.memory.metric_name` (default: "memory_allocation")
- All filtering options are optional

## Advanced Configuration

### Multiple Attribute Rules

```yaml
connectors:
  profiletometrics:
    attributes:
      - key: "service.name"
        value: "my-service"
      - key: "service.version"
        value: "v[0-9]+\\.[0-9]+"
      - key: "deployment.environment"
        value: "production"
      - key: "k8s.pod.name"
        value: "pod-.*"
```

### Complex Filtering

```yaml
connectors:
  profiletometrics:
    process_filter:
      enabled: true
      pattern: "(my-app|worker|scheduler).*"
    
    thread_filter:
      enabled: true
      pattern: "(main|worker|background)-.*"
    
    pattern_filter:
      enabled: true
      pattern: "(service|deployment|k8s)-.*"
```

### Debug Configuration

```yaml
connectors:
  profiletometrics:
    # ... other configuration ...
    
    # Enable debug logging
    debug:
      enabled: true
      log_level: "debug"
      log_samples: true
      log_attributes: true
```

## Querying Function Metrics

When function metrics are enabled, you can query them using the `function.name` attribute:

### PromQL Examples

**Total CPU time by function:**
```promql
sum by (function.name) (cpu_time{function.name!=""})
```

**Top 10 functions by CPU time:**
```promql
topk(10, sum by (function.name) (rate(cpu_time{function.name!=""}[5m])))
```

**Total memory allocation per function:**
```promql
sum by (function.name) (memory_allocation{function.name!=""})
```

**Functions with highest memory usage:**
```promql
topk(5, sum by (function.name) (rate(memory_allocation{function.name!=""}[5m])))
```

**Filter specific function:**
```promql
cpu_time{function.name="myHandler"}
```

**Functions in a specific service:**
```promql
cpu_time{service.name="my-service", function.name!=""}
```

### Metrics Structure

With function metrics enabled, your metric hierarchy looks like:

```
Total Metrics:
├── cpu_time                           # Overall CPU time
├── cpu_time{function.name="main"}     # Per-function CPU (with function.name attribute)
├── cpu_time{function.name="handler"}  # Per-function CPU (with function.name attribute)
├── memory_allocation                  # Overall memory
├── memory_allocation{function.name="main"}     # Per-function memory (with function.name attribute)
└── memory_allocation{function.name="handler"}  # Per-function memory (with function.name attribute)

With Attributes:
├── cpu_time{service.name="app"}
├── cpu_time{service.name="app", function.name="main"}
├── cpu_time{service.name="app", function.name="handler"}
└── ...
```

## Configuration Examples

### Simple Setup

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
      memory:
        enabled: true
      function:
        enabled: false                  # Disable to reduce metric cardinality
    attributes:
      - key: "service.name"
        value: "my-service"
```

**Note**: In the simple setup example above, function-level metrics are disabled to reduce cardinality. Enable them if you need per-function metrics.

### Production Setup

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
      function:
        enabled: true                   # Enable function-level metrics
    
    attributes:
      - key: "service.name"
        value: "my-service"
      - key: "service.version"
        value: "v[0-9]+\\.[0-9]+"
      - key: "deployment.environment"
        value: "production"
      - key: "k8s.namespace"
        value: "default"
      - key: "k8s.pod.name"
        value: "pod-.*"
    
    process_filter:
      enabled: true
      pattern: "my-app.*"
    
    thread_filter:
      enabled: true
      pattern: "worker-.*"
```

## Troubleshooting Configuration

### Common Issues

#### 1. Invalid Regex Patterns

```yaml
# ❌ Invalid - unescaped dots
pattern: "service.*"

# ✅ Valid - escaped dots
pattern: "service\\..*"
```

#### 2. Missing Required Fields

```yaml
# ❌ Invalid - no metrics enabled
connectors:
  profiletometrics:
    attributes:
      - key: "service.name"
        value: "my-service"

# ✅ Valid - at least one metric enabled
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
    attributes:
      - key: "service.name"
        value: "my-service"
```

#### 3. Configuration Validation

```bash
# Validate configuration
otelcol --config config.yaml --dry-run
```

### Debug Configuration

Enable debug logging to troubleshoot configuration issues:

```yaml
service:
  telemetry:
    logs:
      level: debug
      development: true
```
