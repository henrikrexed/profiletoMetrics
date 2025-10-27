# Kubernetes Deployment

Deploy the ProfileToMetrics Connector on Kubernetes using the provided manifests.

## Feature Gates

**⚠️ Important**: The ProfileToMetrics connector requires the `+service.profilesSupport` feature gate to be enabled. This is configured in the deployment manifest.

## Quick Start

### Apply All Manifests

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Apply RBAC
kubectl apply -f k8s/rbac.yaml

# Apply configuration
kubectl apply -f k8s/configmap.yaml

# Deploy collector
kubectl apply -f k8s/deployment.yaml

# Create service
kubectl apply -f k8s/service.yaml
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -n otel-collector

# Check logs
kubectl logs -n otel-collector deployment/otel-collector

# Check service
kubectl get svc -n otel-collector
```

## Manifest Details

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: otel-collector
  labels:
    name: otel-collector
```

### RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: otel-collector
  namespace: otel-collector
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: otel-collector
rules:
  - apiGroups: [""]
    resources: ["nodes", "nodes/proxy", "nodes/stats", "nodes/metrics"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods", "services", "endpoints"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: otel-collector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: otel-collector
subjects:
  - kind: ServiceAccount
    name: otel-collector
    namespace: otel-collector
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: otel-collector
data:
  config.yaml: |
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
          - key: "k8s.namespace"
            value: "otel-collector"
          - key: "k8s.pod.name"
            value: "pod-.*"

    exporters:
      debug:
        verbosity: detailed
      otlp:
        endpoint: "http://observability-platform:4317"

    service:
      pipelines:
        profiles:
          receivers: [otlp]
          exporters: [profiletometrics]
        metrics:
          receivers: [profiletometrics]
          exporters: [debug, otlp]
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector
  namespace: otel-collector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: otel-collector
  template:
    metadata:
      labels:
        app: otel-collector
    spec:
      serviceAccountName: otel-collector
      containers:
        - name: otel-collector
          image: hrexed/otel-collector-profilemetrics:0.1.0
          ports:
            - containerPort: 4317
              name: otlp-grpc
            - containerPort: 4318
              name: otlp-http
            - containerPort: 8888
              name: health-check
          env:
            - name: OTEL_LOG_LEVEL
              value: "info"
          volumeMounts:
            - name: config
              mountPath: /etc/otelcol
          livenessProbe:
            httpGet:
              path: /
              port: 8888
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /
              port: 8888
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            limits:
              cpu: 1000m
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 512Mi
      volumes:
        - name: config
          configMap:
            name: otel-collector-config
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: otel-collector
  namespace: otel-collector
spec:
  selector:
    app: otel-collector
  ports:
    - name: otlp-grpc
      port: 4317
      targetPort: 4317
    - name: otlp-http
      port: 4318
      targetPort: 4318
    - name: health-check
      port: 8888
      targetPort: 8888
  type: ClusterIP
```

## Troubleshooting

### Common Issues

#### 1. Pod Not Starting

```bash
# Check pod status
kubectl get pods -n otel-collector

# Check pod events
kubectl describe pod -n otel-collector <pod-name>

# Check logs
kubectl logs -n otel-collector <pod-name>
```
