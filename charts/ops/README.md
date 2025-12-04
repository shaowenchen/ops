# Ops Helm Chart

This Helm chart deploys the Ops platform on a Kubernetes cluster. Ops is a Kubernetes Operator that provides automation for host management, cluster management, and task execution.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- NATS server (for event streaming, optional but recommended)

## Installation

### Add the Helm Repository

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
helm repo update
```

### Install Ops

Basic installation:

```bash
helm install myops ops/ops --version 2.0.0 --namespace ops-system --create-namespace
```

Installation with custom values:

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  --set controller.env.eventEndpoint="http://app:password@nats-headless.ops-system.svc:4222" \
  --set replicaCount=2
```

Installation with values file:

```bash
# Create a custom values file
cat > my-values.yaml <<EOF
replicaCount: 2
controller:
  env:
    activeNamespace: "ops-system"
    eventEndpoint: "http://app:password@nats-headless.ops-system.svc:4222"
server:
  env:
    eventEndpoint: "http://app:password@nats-headless.ops-system.svc:4222"
resources:
  limits:
    cpu: 2000m
    memory: 4096Mi
  requests:
    cpu: 1000m
    memory: 2048Mi
prometheus:
  enabled: true
EOF

helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  -f my-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas for controller and server | `1` |
| `image.repository` | Controller image repository | `registry.cn-beijing.aliyuncs.com/opshub/shaowenchen-ops-controller-manager` |
| `image.tag` | Controller image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `Always` |
| `server.image.repository` | Server image repository | `registry.cn-beijing.aliyuncs.com/opshub/shaowenchen-ops-server` |
| `server.image.tag` | Server image tag | `latest` |
| `server.image.pullPolicy` | Server image pull policy | `Always` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `ingress.enabled` | Enable ingress | `false` |
| `resources.limits.cpu` | CPU limit | `1000m` |
| `resources.limits.memory` | Memory limit | `2048Mi` |
| `resources.requests.cpu` | CPU request | `500m` |
| `resources.requests.memory` | Memory request | `1024Mi` |
| `autoscaling.enabled` | Enable HPA | `false` |
| `autoscaling.minReplicas` | Minimum replicas for HPA | `1` |
| `autoscaling.maxReplicas` | Maximum replicas for HPA | `100` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization | `80` |
| `prometheus.enabled` | Enable Prometheus monitoring | `true` |
| `controller.env.activeNamespace` | Active namespace for processing CRDs (empty = all namespaces) | `""` |
| `controller.env.defaultRuntimeImage` | Default runtime image for tasks | `registry.cn-beijing.aliyuncs.com/opshub/ubuntu:22.04` |
| `controller.env.eventCluster` | Event cluster name | `default` |
| `controller.env.eventEndpoint` | NATS event endpoint | `http://app:mypassword@nats-headless.ops-system.svc:4222` |
| `server.env.eventCluster` | Event cluster name | `default` |
| `server.env.eventEndpoint` | NATS event endpoint | `http://app:mypassword@nats-headless.ops-system.svc:4222` |

## Components

This chart deploys two main components:

### 1. Controller Manager

The controller manager is a Kubernetes Operator that manages CRDs:
- **Host**: Manages host machines
- **Cluster**: Manages cluster information
- **Task**: Manages one-time and scheduled tasks
- **TaskRun**: Executes tasks
- **Pipeline**: Manages pipelines
- **PipelineRun**: Executes pipelines
- **EventHooks**: Manages event hooks

### 2. Server

The server provides HTTP API and web interface for managing Ops resources.

## Monitoring

When `prometheus.enabled` is set to `true`, the chart creates:
- ServiceMonitor resources for Prometheus Operator
- Metrics services exposing `/metrics` endpoints

### Metrics Endpoints

- Controller: `http://<release-name>-ops-controller-metrics:8080/metrics`
- Server: `http://<release-name>-ops-server:80/metrics`

## Upgrading

To upgrade the release:

```bash
helm upgrade myops ops/ops --version 2.0.0 --namespace ops-system
```

## Uninstallation

To uninstall the release:

```bash
helm uninstall myops --namespace ops-system
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n ops-system
```

### Check Logs

```bash
# Controller logs
kubectl logs -n ops-system deployment/<release-name>-ops

# Server logs
kubectl logs -n ops-system deployment/<release-name>-ops-server
```

### Check Services

```bash
kubectl get svc -n ops-system
```

### Check ServiceMonitors

```bash
kubectl get servicemonitor -n ops-system
```

### Verify Metrics Endpoints

```bash
# Controller metrics
kubectl port-forward -n ops-system svc/<release-name>-ops-controller-metrics 8080:8080
curl http://localhost:8080/metrics

# Server metrics
kubectl port-forward -n ops-system svc/<release-name>-ops-server 8080:80
curl http://localhost:8080/metrics
```

## Additional Resources

- [Ops Documentation](https://www.chenshaowen.com/ops)
- [GitHub Repository](https://github.com/shaowenchen/ops)
- [NATS Installation Guide](https://www.chenshaowen.com/ops/nats)

## License

See the LICENSE file in the repository.
