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
  --set controller.image.repository="shaowenchen/ops-controller-manager" \
  --set controller.image.pullPolicy="Always" \
  --set controller.image.tag="latest" \
  --set controller.env.activeNamespace="ops-system" \
  --set controller.env.defaultRuntimeImage="ubuntu:22.04" \
  --set server.image.repository="shaowenchen/ops-server" \
  --set server.image.pullPolicy="Always" \
  --set server.image.tag="latest" \
  --set event.cluster="mycluster" \
  --set event.endpoint="http://app:password@nats-headless.ops-system.svc:4222"
```

### Key Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `EVENT_CLUSTER` | The cluster name for event identification | `mycluster` |
| `EVENT_ENDPOINT` | NATS server endpoint with credentials | `http://app:password@nats-headless.ops-system.svc:4222` |
| `ACTIVE_NAMESPACE` | Namespace to watch (empty = all namespaces) | `ops-system` |
| `DEFAULT_RUNTIME_IMAGE` | Default image for task execution | `ubuntu:22.04` |

Installation with values file:

```bash
# Create a custom values file
cat > my-values.yaml <<EOF
event:
  cluster: "mycluster"
  endpoint: "http://app:password@nats-headless.ops-system.svc:4222"
controller:
  replicaCount: 2
  image:
    repository: shaowenchen/ops-controller-manager
    pullPolicy: Always
    tag: "latest"
  env:
    activeNamespace: "ops-system"
    defaultRuntimeImage: "ubuntu:22.04"
server:
  replicaCount: 2
  image:
    repository: shaowenchen/ops-server
    pullPolicy: Always
    tag: "latest"
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 500m
      memory: 512Mi
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 4
    targetCPUUtilizationPercentage: 80
resources:
  limits:
    cpu: 1000m
    memory: 1024Mi
  requests:
    cpu: 500m
    memory: 512Mi
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
| `event.cluster` | Event cluster name (shared by controller and server) | `default` |
| `event.endpoint` | NATS event endpoint (shared by controller and server) | `http://app:mypassword@nats-headless.ops-system.svc:4222` |
| `controller.replicaCount` | Number of replicas for controller | `2` |
| `controller.image.repository` | Controller image repository | `shaowenchen/ops-controller-manager` |
| `controller.image.tag` | Controller image tag | `latest` |
| `controller.image.pullPolicy` | Controller image pull policy | `Always` |
| `controller.env.activeNamespace` | Active namespace for processing CRDs (empty = all namespaces) | `ops-system` |
| `controller.env.defaultRuntimeImage` | Default runtime image for tasks | `ubuntu:22.04` |
| `server.replicaCount` | Number of replicas for server | `2` |
| `server.image.repository` | Server image repository | `shaowenchen/ops-server` |
| `server.image.tag` | Server image tag | `latest` |
| `server.image.pullPolicy` | Server image pull policy | `Always` |
| `server.resources.limits.cpu` | Server CPU limit | `500m` |
| `server.resources.limits.memory` | Server memory limit | `512Mi` |
| `server.resources.requests.cpu` | Server CPU request | `500m` |
| `server.resources.requests.memory` | Server memory request | `512Mi` |
| `server.autoscaling.enabled` | Enable HPA for server | `true` |
| `server.autoscaling.minReplicas` | Minimum replicas for server HPA | `2` |
| `server.autoscaling.maxReplicas` | Maximum replicas for server HPA | `4` |
| `server.autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization for server HPA | `80` |
| `resources.limits.cpu` | Global CPU limit (for controller) | `1000m` |
| `resources.limits.memory` | Global memory limit (for controller) | `1024Mi` |
| `resources.requests.cpu` | Global CPU request (for controller) | `500m` |
| `resources.requests.memory` | Global memory request (for controller) | `512Mi` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `ingress.enabled` | Enable ingress | `false` |
| `prometheus.enabled` | Enable Prometheus monitoring | `true` |

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

- Controller: `http://<release-name>-ops-controller-metrics:9090/metrics`
- Server: `http://<release-name>-ops-server:9090/metrics`

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
kubectl port-forward -n ops-system svc/<release-name>-ops-controller-metrics 9090:9090
curl http://localhost:9090/metrics

# Server metrics
kubectl port-forward -n ops-system svc/<release-name>-ops-server 9090:9090
curl http://localhost:9090/metrics
```

## Additional Resources

- [Ops Documentation](https://www.chenshaowen.com/ops)
- [GitHub Repository](https://github.com/shaowenchen/ops)
- [NATS Installation Guide](https://www.chenshaowen.com/ops/nats)

## License

See the LICENSE file in the repository.
