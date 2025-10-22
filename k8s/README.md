# Kubernetes Deployment for ProfileToMetrics Connector

This directory contains Kubernetes manifests for deploying the OpenTelemetry ProfileToMetrics connector in a Kubernetes cluster.

## 📁 Files Overview

| File | Description |
|------|-------------|
| `namespace.yaml` | Namespace definitions for the collector |
| `rbac.yaml` | ServiceAccount, ClusterRole, and ClusterRoleBinding for Kubernetes API access |
| `configmap.yaml` | OpenTelemetry Collector configuration with ProfileToMetrics connector |
| `deployment.yaml` | Deployment manifest for the collector |
| `service.yaml` | Service definitions for internal access |

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster (v1.21+)
- kubectl configured
- OpenTelemetry Collector image available

### ⚠️ Important: Feature Gate Required

The ProfileToMetrics connector requires the `+service.profilesSupport` feature gate to be enabled. This is already configured in the deployment manifest.

### 1. Deploy with kubectl

```bash
# Create namespace and RBAC
kubectl apply -f namespace.yaml
kubectl apply -f rbac.yaml

# Deploy configuration
kubectl apply -f configmap.yaml

# Deploy the collector
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

### 2. Deploy all at once

```bash
# Deploy everything
kubectl apply -f namespace.yaml
kubectl apply -f rbac.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

### 3. Verify Deployment

```bash
# Check pods
kubectl get pods -n otel-collector

# Check services
kubectl get svc -n otel-collector

# Check logs
kubectl logs -n otel-collector -l app.kubernetes.io/name=otel-collector

# Check metrics endpoint
kubectl port-forward -n otel-collector svc/otel-collector 8888:8888
curl http://localhost:8888/metrics
```

## 🏗 Architecture

### Deployment
- **Use case**: Centralized collection
- **Benefits**: Resource efficiency, easier management
- **Scaling**: Horizontal scaling with replicas

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `KUBE_NODE_NAME` | Auto-detected | Kubernetes node name |
| `OTEL_RESOURCE_ATTRIBUTES` | Set in manifest | Resource attributes |
| `OTEL_SERVICE_NAME` | `otel-collector` | Service name |
| `OTEL_SERVICE_VERSION` | `0.1.0` | Service version |

### Resource Requirements

| Component | CPU Request | Memory Request | CPU Limit | Memory Limit |
|-----------|-------------|----------------|-----------|--------------|
| Deployment | 100m | 256Mi | 500m | 1Gi |
| DaemonSet | 50m | 128Mi | 200m | 512Mi |

### Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 4317 | TCP | OTLP gRPC receiver |
| 4318 | TCP | OTLP HTTP receiver |
| 8888 | TCP | Health check endpoint |

## 🔧 Customization

### 1. Modify Configuration

Edit `configmap.yaml` to customize the OpenTelemetry Collector configuration:

```yaml
data:
  config.yaml: |
    # Your custom configuration here
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
    # ... rest of configuration
```

### 2. Update Image

```bash
# Update image tag
kubectl set image deployment/otel-collector otel-collector=hrexed/otel-collector-profilemetrics:latest -n otel-collector

# Or edit the deployment
kubectl edit deployment otel-collector -n otel-collector
```

### 3. Scale Deployment

```bash
# Scale to 3 replicas
kubectl scale deployment otel-collector --replicas=3 -n otel-collector

# Or edit the deployment
kubectl edit deployment otel-collector -n otel-collector
```

### 4. Customize with Kustomize

Edit `kustomization.yaml`:

```yaml
# Change image tag
images:
  - name: hrexed/otel-collector-profilemetrics
    newTag: "latest"

# Add patches
patchesStrategicMerge:
  - custom-patch.yaml
```

## 📊 Monitoring

### Basic Health Checks

The deployment includes basic health checks:

- **Liveness Probe**: HTTP GET on `/metrics` endpoint
- **Readiness Probe**: HTTP GET on `/metrics` endpoint
- **Resource Monitoring**: CPU and memory limits enforced

## 🔒 Security

### RBAC Permissions

The collector requires the following permissions:

- **Pods**: Read access for k8sattributes processor
- **Nodes**: Read access for node information
- **Namespaces**: Read access for namespace information
- **Services**: Read access for service discovery
- **ReplicaSets/Deployments**: Read access for workload information

### Security Context

- **Non-root user**: Runs as user 10001
- **Read-only filesystem**: Prevents write access
- **No privilege escalation**: Dropped all capabilities
- **Resource limits**: CPU and memory limits enforced

## 🌐 Networking

### Service Types

- **ClusterIP**: Internal cluster access
- **Headless**: DNS discovery for stateful services

## 🚨 Troubleshooting

### Common Issues

1. **Pod not starting**
   ```bash
   kubectl describe pod -n otel-collector -l app.kubernetes.io/name=otel-collector
   kubectl logs -n otel-collector -l app.kubernetes.io/name=otel-collector
   ```

2. **Configuration errors**
   ```bash
   kubectl get configmap otel-collector-config -n otel-collector -o yaml
   ```

3. **RBAC issues**
   ```bash
   kubectl auth can-i get pods --as=system:serviceaccount:otel-collector:otel-collector
   ```

4. **Resource constraints**
   ```bash
   kubectl top pods -n otel-collector
   kubectl describe nodes
   ```

### Debug Commands

```bash
# Check collector status
kubectl get pods -n otel-collector -o wide

# Check logs
kubectl logs -n otel-collector -l app.kubernetes.io/name=otel-collector --tail=100

# Check configuration
kubectl exec -n otel-collector deployment/otel-collector -- cat /etc/otelcol/config.yaml

# Test endpoints
kubectl port-forward -n otel-collector svc/otel-collector 8888:8888
curl http://localhost:8888/metrics
```

## 📚 Examples

### 1. Basic Deployment

```bash
# Deploy basic collector
kubectl apply -f namespace.yaml
kubectl apply -f rbac.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

### 2. Development Deployment

```bash
# Deploy with debug configuration
kubectl apply -f namespace.yaml
kubectl apply -f rbac.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the deployment
5. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../../LICENSE) file for details.
