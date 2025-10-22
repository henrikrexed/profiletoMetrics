# Project Overview

This document provides a high-level overview of the OpenTelemetry Profile-to-Metrics Connector project structure and components.

## 📁 Project Structure

```
profiletoMetrics/
├── pkg/profiletometrics/          # Core converter library
│   ├── converter.go               # Main converter implementation
│   └── converter_test.go          # Unit tests
├── examples/                      # Configuration examples
│   ├── simple-config.yaml        # Basic collector configuration
│   └── collector-config.yaml      # Advanced collector configuration
├── docker/                        # Docker configuration
│   └── Dockerfile                 # Multi-stage Docker build
├── scripts/                       # Utility scripts
│   ├── build-examples.sh         # Build demonstration script
│   └── send-test-data.sh          # Test data sending script
├── testdata/                      # Test data and mocks
│   └── profile_test_data.go       # Mock profile data for testing
├── data-example/                  # Example profile data
│   └── profile.log                # Sample profile log file
├── Makefile                       # Build automation
├── ocb-simple.yaml               # OpenTelemetry Collector Builder config
├── README.md                      # Main documentation
├── USAGE.md                       # Detailed usage guide
├── DOCKER.md                      # Docker deployment guide
├── TESTING.md                     # Testing documentation
└── PROJECT_OVERVIEW.md           # This file
```

## 🧩 Core Components

### 1. Profile-to-Metrics Converter (`pkg/profiletometrics/`)

The heart of the project - a Go library that converts OpenTelemetry profiling data into metrics.

**Key Features:**
- CPU time metric generation
- Memory allocation metric generation
- Flexible attribute extraction (literal, regex, string table)
- Process and thread filtering
- Pattern-based filtering

**Main Types:**
- `Config`: Configuration structure
- `Converter`: Main converter implementation
- `MetricsConfig`: Metric generation settings
- `AttributeExtractionConfig`: Attribute extraction rules
- `ProcessFilterConfig`: Process filtering settings
- `ThreadFilterConfig`: Thread filtering settings
- `PatternFilterConfig`: Pattern filtering settings

### 2. Configuration Examples (`examples/`)

Ready-to-use configuration files for different scenarios:

- **`simple-config.yaml`**: Basic collector setup with OTLP receiver/exporter
- **`collector-config.yaml`**: Advanced setup with processors and multiple exporters

### 3. Docker Support (`docker/`)

Containerized deployment using multi-stage builds:

- **Base Image**: Alpine Linux 3.19
- **Builder**: Go 1.24 with OCB (OpenTelemetry Collector Builder)
- **Runtime**: Non-root user, health checks, proper labels
- **Size**: ~63.5 MB

### 4. Build Automation (`Makefile`)

Comprehensive build system with configurable options:

**Key Targets:**
- `build`: Build the project
- `test`: Run tests with coverage
- `docker-build`: Build Docker image
- `docker-build-multi`: Multi-platform builds
- `clean`: Clean build artifacts

**Configuration Variables:**
- `DOCKER_BINARY`: docker/podman
- `DOCKER_PLATFORM`: linux/amd64, linux/arm64
- `DOCKER_IMAGE`: Image name
- `DOCKER_TAG`: Image tag

### 5. Testing Infrastructure

**Unit Tests:**
- `pkg/profiletometrics/converter_test.go`: Core converter tests
- `testdata/profile_test_data.go`: Mock data generation

**Test Scripts:**
- `run_tests.sh`: Test execution script
- `scripts/send-test-data.sh`: Integration testing

## 🔧 Build System

### OpenTelemetry Collector Builder (OCB)

The project uses OCB to build a custom OpenTelemetry Collector:

```yaml
# ocb-simple.yaml
dist:
  name: otelcol-profiletometrics
  description: "OpenTelemetry Collector with Profile to Metrics Connector"
  output_path: ./dist

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.118.0
  - gomod: go.opentelemetry.io/collector/exporter/debugexporter v0.118.0

receivers:
  - gomod: go.opentelemetry.io/collector/receiver/otlpreceiver v0.118.0

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.118.0
```

### Docker Build Process

1. **Builder Stage**: Go 1.24 + OCB → Build collector binary
2. **Runtime Stage**: Alpine 3.19 + Binary → Production image

## 📊 Key Metrics Generated

### CPU Time Metrics
- **Name**: `cpu_time_seconds`
- **Type**: Counter
- **Description**: CPU time spent in seconds
- **Attributes**: Extracted from profile data

### Memory Allocation Metrics
- **Name**: `memory_allocation_bytes`
- **Type**: Counter
- **Description**: Memory allocated in bytes
- **Attributes**: Extracted from profile data

## 🎯 Use Cases

### 1. Application Performance Monitoring
Convert application profiles into metrics for monitoring dashboards.

### 2. Resource Usage Tracking
Track CPU and memory usage patterns across services.

### 3. Performance Optimization
Identify performance bottlenecks through metric analysis.

### 4. Capacity Planning
Use historical metrics for capacity planning decisions.

## 🔄 Data Flow

```
Profile Data → Converter → Metrics → Exporters → Monitoring Backend
     ↓              ↓         ↓         ↓
  pprofile.    Config    pmetric.   OTLP/Debug
  Profiles              Metrics
```

## 🛠 Development Workflow

### 1. Local Development
```bash
# Install dependencies
make install-deps

# Run tests
make test

# Build locally
make build
```

### 2. Docker Development
```bash
# Build Docker image
make docker-build

# Run container
docker run -p 4317:4317 hrexed/otel-collector-profilemetrics:0.1.0
```

### 3. Testing
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific tests
go test -v ./pkg/profiletometrics/...
```

## 📚 Documentation Structure

| Document | Purpose |
|----------|---------|
| `README.md` | Main project documentation |
| `USAGE.md` | Detailed usage examples and patterns |
| `DOCKER.md` | Docker deployment guide |
| `TESTING.md` | Testing strategies and examples |
| `PROJECT_OVERVIEW.md` | This overview document |

## 🚀 Getting Started

### Quick Start
1. Clone the repository
2. Run `make install-deps`
3. Run `make test`
4. Build Docker image: `make docker-build`
5. Run container: `docker run -p 4317:4317 hrexed/otel-collector-profilemetrics:0.1.0`

### Next Steps
- Read `USAGE.md` for detailed usage examples
- Check `DOCKER.md` for deployment options
- Review `TESTING.md` for testing strategies

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## 📄 License

Apache License 2.0 - see LICENSE file for details.
