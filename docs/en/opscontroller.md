### **Ops-controller-manager Overview**

`ops-controller-manager` is a Kubernetes Operator that provides three core objects: `Host`, `Cluster`, and `Task`.

#### **Objects in ops-controller-manager**

- **Host**: Describes a host machine, including hostname, IP address, SSH username, password, private key, etc.
- **Cluster**: Describes a cluster, including details such as cluster name, number of hosts, number of pods, load, CPU, memory, etc.
- **Task**: Describes a task, which can be one-time or scheduled (cron) tasks.

### **Installation**

#### **Prerequisites**

- Kubernetes 1.19+
- Helm 3.0+
- NATS server (optional but recommended for event streaming)

#### **Install Helm**

If you haven't installed Helm yet, you can install it using the following command:

```bash
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

#### **Add Helm Repository**

Add the `ops` Helm repository to your Helm configuration:

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
helm repo update
```

#### **Install ops-controller-manager**

**Basic Installation:**

To install the `ops-controller-manager` using Helm with default values:

```bash
helm install myops ops/ops --version 2.0.0 --namespace ops-system --create-namespace
```

**Installation with Custom Values:**

You can customize the installation using `--set` flags:

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  --set controller.image.repository="shaowenchen/ops-controller-manager" \
  --set controller.image.pullPolicy="Always" \
  --set controller.image.tag="latest" \
  --set controller.env.activeNamespace="ops-system" \
  --set controller.env.defaultRuntimeImage="ubuntu:22.04" \
  --set controller.replicaCount=2 \
  --set server.image.repository="shaowenchen/ops-server" \
  --set server.image.pullPolicy="Always" \
  --set server.image.tag="latest" \
  --set server.replicaCount=2 \
  --set server.autoscaling.minReplicas=2 \
  --set server.autoscaling.maxReplicas=4 \
  --set event.cluster="mycluster" \
  --set event.endpoint="http://app:password@nats-headless.ops-system.svc:4222"
```

**Installation with Values File:**

Create a custom values file for more complex configurations:

```yaml
# my-values.yaml
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
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 4
    targetCPUUtilizationPercentage: 80

prometheus:
  enabled: true
```

Then install with the values file:

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  -f my-values.yaml
```

#### **Verify Installation**

After the installation, check the status of the pods to ensure everything is running:

```bash
# Check pods
kubectl get pods -n ops-system

# Check services
kubectl get svc -n ops-system

# Check deployments
kubectl get deployments -n ops-system
```

#### **Upgrade Installation**

To upgrade an existing installation:

```bash
helm upgrade myops ops/ops --version 2.0.0 --namespace ops-system
```

Or with custom values:

```bash
helm upgrade myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  -f my-values.yaml
```

#### **Uninstall ops-controller-manager**

To uninstall the `ops-controller-manager`:

```bash
helm uninstall myops --namespace ops-system
```

### **Configuration**

#### **Namespace Configuration**

By default, `ops-controller-manager` will only process CRD resources within the `ops-system` namespace.

To change this behavior, you can modify the `controller.env.activeNamespace` value in the Helm values. If left empty, it will process resources from all namespaces.

**Using --set:**

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  --set controller.env.activeNamespace=""
```

**Using values file:**

```yaml
controller:
  env:
    activeNamespace: "" # Empty means all namespaces
```

#### **Event Configuration**

Configure NATS event endpoint for both controller and server:

```yaml
controller:
  env:
    eventCluster: "default"
    eventEndpoint: "http://app:password@nats-headless.ops-system.svc:4222"

server:
  env:
    eventCluster: "default"
    eventEndpoint: "http://app:password@nats-headless.ops-system.svc:4222"
```

#### **Resource Configuration**

Configure resource limits and requests:

```yaml
resources:
  limits:
    cpu: 2000m
    memory: 4096Mi
  requests:
    cpu: 1000m
    memory: 2048Mi
```

You can also configure separate resources for server:

```yaml
server:
  resources:
    limits:
      cpu: 1000m
      memory: 2048Mi
    requests:
      cpu: 500m
      memory: 1024Mi
```

#### **Monitoring Configuration**

Enable Prometheus monitoring (creates ServiceMonitor resources):

```yaml
prometheus:
  enabled: true
```

#### **Autoscaling Configuration**

Enable Horizontal Pod Autoscaler:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
```

### **Troubleshooting**

#### **Check Pod Logs**

```bash
# Controller logs
kubectl logs -n ops-system deployment/myops-ops

# Server logs
kubectl logs -n ops-system deployment/myops-ops-server
```

#### **Check Metrics Endpoints**

```bash
# Controller metrics
kubectl port-forward -n ops-system svc/myops-ops-controller-metrics 8080:8080
curl http://localhost:8080/metrics

# Server metrics
kubectl port-forward -n ops-system svc/myops-ops-server 8080:80
curl http://localhost:8080/metrics
```

#### **Check ServiceMonitors**

```bash
kubectl get servicemonitor -n ops-system
```

For more detailed configuration options, see the [Helm Chart README](../../charts/ops/README.md).
