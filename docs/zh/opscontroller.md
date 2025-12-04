## Ops-controller-manager

### 功能简介

ops-controller-manager 是一个 Kubernetes Operator。它提供了三种对象：`Host`, `Cluster`, `Task`。

### Host

Host 对象用于描述一个主机，比如主机名，IP 地址，SSH 用户名，SSH 密码，SSH 私钥等。

### Cluster

Cluster 对象用于描述一个集群，比如集群名，集群的主机数量、Pod 数量、负载、CPU、内存等。

### Task

Task 对象用于描述一个任务，比如一次性任务、周期任务。

## 安装

### 前置要求

- Kubernetes 1.19+
- Helm 3.0+
- NATS 服务器（可选，但推荐用于事件流）

### 安装 Helm

```bash
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

### 添加 Helm 仓库

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
helm repo update
```

### 安装 ops-controller-manager

**基础安装：**

使用默认配置安装：

```bash
helm install myops ops/ops --version 2.0.0 --namespace ops-system --create-namespace
```

**使用自定义值安装：**

可以使用 `--set` 参数自定义配置：

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  --set controller.env.eventEndpoint="http://app:password@nats-headless.ops-system.svc:4222" \
  --set replicaCount=2 \
  --set resources.limits.memory=4096Mi
```

**使用 values 文件安装：**

创建自定义 values 文件以进行更复杂的配置：

```yaml
# my-values.yaml
replicaCount: 2

controller:
  env:
    activeNamespace: "ops-system"
    eventEndpoint: "http://app:password@nats-headless.ops-system.svc:4222"
    defaultRuntimeImage: "registry.cn-beijing.aliyuncs.com/opshub/ubuntu:22.04"

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

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

然后使用 values 文件安装：

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  -f my-values.yaml
```

### 查看安装结果

安装后，检查 Pod 状态以确保一切正常运行：

```bash
# 检查 Pod
kubectl get pods -n ops-system

# 检查 Service
kubectl get svc -n ops-system

# 检查 Deployment
kubectl get deployments -n ops-system
```

### 升级安装

升级现有安装：

```bash
helm upgrade myops ops/ops --version 2.0.0 --namespace ops-system
```

或使用自定义值：

```bash
helm upgrade myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  -f my-values.yaml
```

### 卸载

卸载 ops-controller-manager：

```bash
helm uninstall myops --namespace ops-system
```

## 配置

### 命名空间配置

默认情况下，`ops-controller-manager` 只会处理 `ops-system` 命名空间下的 CRD 资源。

如果需要变更，可以修改 Helm values 中的 `controller.env.activeNamespace` 值。如果为空，则会处理所有命名空间的资源。

**使用 --set：**

```bash
helm install myops ops/ops --version 2.0.0 \
  --namespace ops-system \
  --create-namespace \
  --set controller.env.activeNamespace=""
```

**使用 values 文件：**

```yaml
controller:
  env:
    activeNamespace: ""  # 空值表示所有命名空间
```

### 事件配置

为 controller 和 server 配置 NATS 事件端点：

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

### 资源配置

配置资源限制和请求：

```yaml
resources:
  limits:
    cpu: 2000m
    memory: 4096Mi
  requests:
    cpu: 1000m
    memory: 2048Mi
```

也可以为 server 单独配置资源：

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

### 监控配置

启用 Prometheus 监控（创建 ServiceMonitor 资源）：

```yaml
prometheus:
  enabled: true
```

### 自动扩缩容配置

启用水平 Pod 自动扩缩容：

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
```

## 故障排查

### 查看 Pod 日志

```bash
# Controller 日志
kubectl logs -n ops-system deployment/myops-ops

# Server 日志
kubectl logs -n ops-system deployment/myops-ops-server
```

### 检查指标端点

```bash
# Controller 指标
kubectl port-forward -n ops-system svc/myops-ops-controller-metrics 8080:8080
curl http://localhost:8080/metrics

# Server 指标
kubectl port-forward -n ops-system svc/myops-ops-server 8080:80
curl http://localhost:8080/metrics
```

### 检查 ServiceMonitor

```bash
kubectl get servicemonitor -n ops-system
```

更多详细配置选项，请参阅 [Helm Chart README](../../charts/ops/README.md)。
