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

```go
package main

import (
    "context"
    "log"
    "example.com/profiletoMetrics/pkg/profiletometrics"
    "go.opentelemetry.io/collector/pdata/pprofile"
)

func main() {
    // Create configuration
    config := profiletometrics.Config{
        Metrics: profiletometrics.MetricsConfig{
            CPUTime: profiletometrics.MetricConfig{
                Enabled:     true,
                Name:        "cpu_time_seconds",
                Description: "CPU time spent in seconds",
            },
            MemoryAllocation: profiletometrics.MetricConfig{
                Enabled:     true,
                Name:        "memory_allocation_bytes",
                Description: "Memory allocated in bytes",
            },
        },
    }
    
    // Create converter
    converter := profiletometrics.NewConverter(config)
    
    // Your profiling data
    profiles := pprofile.NewProfiles()
    // ... populate with your profile data
    
    // Convert to metrics
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the metrics
    // ... export or process metrics
}
```

### 2. Docker Usage

```bash
# Run the collector
docker run -p 4317:4317 -p 8888:8888 \
  hrexed/otel-collector-profilemetrics:0.1.0

# Send profile data
curl -X POST http://localhost:4317/v1/profiles \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @profile.pb
```

## ⚙️ Configuration Guide

### Core Configuration Structure

```go
type Config struct {
    Metrics            MetricsConfig            `mapstructure:"metrics"`
    AttributeExtraction AttributeExtractionConfig `mapstructure:"attribute_extraction"`
    ProcessFilter       ProcessFilterConfig       `mapstructure:"process_filter"`
    ThreadFilter       ThreadFilterConfig         `mapstructure:"thread_filter"`
    PatternFilter       PatternFilterConfig       `mapstructure:"pattern_filter"`
}
```

### Metrics Configuration

#### CPU Time Metrics

```go
CPUTime: profiletometrics.MetricConfig{
    Enabled:     true,
    Name:        "cpu_time_seconds",
    Description: "CPU time spent in seconds",
}
```

#### Memory Allocation Metrics

```go
MemoryAllocation: profiletometrics.MetricConfig{
    Enabled:     true,
    Name:        "memory_allocation_bytes",
    Description: "Memory allocated in bytes",
}
```

### Attribute Extraction

#### String Table Index Extraction

```go
AttributeExtraction: profiletometrics.AttributeExtractionConfig{
    Rules: []profiletometrics.AttributeRule{
        {
            Name:             "service_name",
            Source:           "string_table",
            StringTableIndex: 0,
        },
        {
            Name:             "pod_name",
            Source:           "string_table", 
            StringTableIndex: 1,
        },
    },
}
```

#### Regular Expression Extraction

```go
AttributeExtraction: profiletometrics.AttributeExtractionConfig{
    Rules: []profiletometrics.AttributeRule{
        {
            Name:             "extracted_value",
            Source:           "regex",
            Pattern:          "service-(.*)-v\\d+",
            StringTableIndex: 2,
        },
    },
}
```

#### Literal Value Extraction

```go
AttributeExtraction: profiletometrics.AttributeExtractionConfig{
    Rules: []profiletometrics.AttributeRule{
        {
            Name:   "environment",
            Source: "literal",
            Value:  "production",
        },
    },
}
```

### Filtering Configuration

#### Process Filtering

```go
ProcessFilter: profiletometrics.ProcessFilterConfig{
    Enabled:             true,
    ProcessNamePattern:  "my-app-.*",
    CompiledProcessRegex: regexp.MustCompile("my-app-.*"),
}
```

#### Thread Filtering

```go
ThreadFilter: profiletometrics.ThreadFilterConfig{
    Enabled:              true,
    ThreadNamePattern:    "worker-.*",
    ProcessNamePattern:   "app-.*",
    CompiledThreadRegex:   regexp.MustCompile("worker-.*"),
    CompiledProcessRegex: regexp.MustCompile("app-.*"),
}
```

#### Pattern Filtering

```go
PatternFilter: profiletometrics.PatternFilterConfig{
    Enabled: true,
    AttributePatterns: []string{
        "service.name=my-service",
        "k8s.pod.name=pod-.*",
        "environment=production",
    },
    CompiledAttributePatterns: []*regexp.Regexp{
        regexp.MustCompile("service.name=my-service"),
        regexp.MustCompile("k8s.pod.name=pod-.*"),
        regexp.MustCompile("environment=production"),
    },
}
```

## 📖 Usage Examples

### Example 1: Basic CPU and Memory Metrics

```go
package main

import (
    "context"
    "log"
    "example.com/profiletoMetrics/pkg/profiletometrics"
    "go.opentelemetry.io/collector/pdata/pprofile"
)

func main() {
    config := profiletometrics.Config{
        Metrics: profiletometrics.MetricsConfig{
            CPUTime: profiletometrics.MetricConfig{
                Enabled:     true,
                Name:        "cpu_time_seconds",
                Description: "CPU time spent in seconds",
            },
            MemoryAllocation: profiletometrics.MetricConfig{
                Enabled:     true,
                Name:        "memory_allocation_bytes",
                Description: "Memory allocated in bytes",
            },
        },
    }
    
    converter := profiletometrics.NewConverter(config)
    
    // Create test profile data
    profiles := createTestProfiles()
    
    // Convert to metrics
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process metrics
    resourceMetrics := metrics.ResourceMetrics()
    for i := 0; i < resourceMetrics.Len(); i++ {
        resourceMetric := resourceMetrics.At(i)
        scopeMetrics := resourceMetric.ScopeMetrics()
        
        for j := 0; j < scopeMetrics.Len(); j++ {
            scopeMetric := scopeMetrics.At(j)
            metricSlice := scopeMetric.Metrics()
            
            for k := 0; k < metricSlice.Len(); k++ {
                metric := metricSlice.At(k)
                log.Printf("Metric: %s", metric.Name())
            }
        }
    }
}

func createTestProfiles() pprofile.Profiles {
    // Implementation to create test profile data
    // ... (see testdata/profile_test_data.go for examples)
    return pprofile.NewProfiles()
}
```

### Example 2: Advanced Filtering

```go
package main

import (
    "context"
    "log"
    "regexp"
    "example.com/profiletoMetrics/pkg/profiletometrics"
    "go.opentelemetry.io/collector/pdata/pprofile"
)

func main() {
    config := profiletometrics.Config{
        Metrics: profiletometrics.MetricsConfig{
            CPUTime: profiletometrics.MetricConfig{
                Enabled:     true,
                Name:        "cpu_time_seconds",
                Description: "CPU time spent in seconds",
            },
        },
        
        // Filter by process name
        ProcessFilter: profiletometrics.ProcessFilterConfig{
            Enabled:             true,
            ProcessNamePattern:  "my-service-.*",
            CompiledProcessRegex: regexp.MustCompile("my-service-.*"),
        },
        
        // Filter by thread name
        ThreadFilter: profiletometrics.ThreadFilterConfig{
            Enabled:              true,
            ThreadNamePattern:    "worker-.*",
            ProcessNamePattern:   "my-service-.*",
            CompiledThreadRegex:   regexp.MustCompile("worker-.*"),
            CompiledProcessRegex: regexp.MustCompile("my-service-.*"),
        },
        
        // Filter by attribute patterns
        PatternFilter: profiletometrics.PatternFilterConfig{
            Enabled: true,
            AttributePatterns: []string{
                "service.name=my-service",
                "environment=production",
            },
            CompiledAttributePatterns: []*regexp.Regexp{
                regexp.MustCompile("service.name=my-service"),
                regexp.MustCompile("environment=production"),
            },
        },
    }
    
    converter := profiletometrics.NewConverter(config)
    
    // Your profile data
    profiles := loadProfiles()
    
    // Convert with filtering
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Generated %d resource metrics", metrics.ResourceMetrics().Len())
}

func loadProfiles() pprofile.Profiles {
    // Load your profile data from file, API, etc.
    // ... implementation
    return pprofile.NewProfiles()
}
```

### Example 3: Attribute Extraction

```go
package main

import (
    "context"
    "log"
    "example.com/profiletoMetrics/pkg/profiletometrics"
    "go.opentelemetry.io/collector/pdata/pprofile"
)

func main() {
    config := profiletometrics.Config{
        Metrics: profiletometrics.MetricsConfig{
            CPUTime: profiletometrics.MetricConfig{
                Enabled:     true,
                Name:        "cpu_time_seconds",
                Description: "CPU time spent in seconds",
            },
        },
        
        // Extract attributes from string table
        AttributeExtraction: profiletometrics.AttributeExtractionConfig{
            Rules: []profiletometrics.AttributeRule{
                {
                    Name:             "service_name",
                    Source:           "string_table",
                    StringTableIndex: 0,
                },
                {
                    Name:             "pod_name",
                    Source:           "string_table",
                    StringTableIndex: 1,
                },
                {
                    Name:             "version",
                    Source:           "regex",
                    Pattern:          "v(\\d+\\.\\d+\\.\\d+)",
                    StringTableIndex: 2,
                },
                {
                    Name:   "environment",
                    Source: "literal",
                    Value:  "production",
                },
            },
        },
    }
    
    converter := profiletometrics.NewConverter(config)
    
    // Your profile data
    profiles := loadProfiles()
    
    // Convert with attribute extraction
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process metrics with extracted attributes
    processMetrics(metrics)
}

func processMetrics(metrics pmetric.Metrics) {
    resourceMetrics := metrics.ResourceMetrics()
    for i := 0; i < resourceMetrics.Len(); i++ {
        resourceMetric := resourceMetrics.At(i)
        
        // Access resource attributes
        attributes := resourceMetric.Resource().Attributes()
        serviceName, _ := attributes.Get("service_name")
        podName, _ := attributes.Get("pod_name")
        
        log.Printf("Processing metrics for service: %s, pod: %s", 
            serviceName.AsString(), podName.AsString())
        
        // Process scope metrics
        scopeMetrics := resourceMetric.ScopeMetrics()
        for j := 0; j < scopeMetrics.Len(); j++ {
            scopeMetric := scopeMetrics.At(j)
            metricSlice := scopeMetric.Metrics()
            
            for k := 0; k < metricSlice.Len(); k++ {
                metric := metricSlice.At(k)
                log.Printf("Metric: %s", metric.Name())
            }
        }
    }
}
```

## 🔗 Integration Patterns

### Pattern 1: Standalone Library Usage

```go
// In your application
import "example.com/profiletoMetrics/pkg/profiletometrics"

func processProfiles(profiles pprofile.Profiles) {
    config := profiletometrics.Config{
        // ... your configuration
    }
    
    converter := profiletometrics.NewConverter(config)
    metrics, err := converter.ConvertProfilesToMetrics(context.Background(), profiles)
    
    // Export metrics to your preferred backend
    exportMetrics(metrics)
}
```

### Pattern 2: OpenTelemetry Collector Integration

```yaml
# collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:
  # Your custom profile-to-metrics processor would go here
  # (when fully integrated with the collector framework)

exporters:
  otlp:
    endpoint: http://localhost:4317
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp, debug]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp, debug]
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
