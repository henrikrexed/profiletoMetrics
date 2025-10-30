# ProfileToMetrics Connector

ProfileToMetrics is an OpenTelemetry Collector connector that converts profiling data into metrics with optional filtering and attribute extraction. It is inspired by connectors in `opentelemetry-collector-contrib` and designed to be easy to run and configure.

## Features

- Convert OpenTelemetry profiles to metrics (CPU time, memory allocation)
- Optional per-function metrics with `function.name` and `file.name` attributes
- Process filtering with one or more regex patterns
- Attribute extraction from the profile string table (literal or regex)
- Ready-to-use Docker and Kubernetes examples

## Quick Start

- Example collector configuration is available in `examples/`.
- Minimal snippet:

```yaml
connectors:
  profiletometrics:
    metrics:
      cpu:
        enabled: true
        metric_name: cpu_time
      memory:
        enabled: true
        metric_name: memory_allocation
    process_filter:
      enabled: true
      patterns: ["my-app.*"]
```

## Documentation

The full documentation (installation, configuration, deployment, and API reference) is hosted on the project website:

- Project Docs: https://henrikrexed.github.io/profiletoMetrics/

## Examples

- Local examples: see the `examples/` directory
- Kubernetes manifests: see the `k8s/` directory

## License

Apache 2.0. See `LICENSE` for details.
