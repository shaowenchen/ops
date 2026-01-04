# Controller Service 说明

## 问题

在 Helm chart 中，controller-manager 缺少对应的 Service，导致 ServiceMonitor 无法匹配到 Service 进行 metrics 采集。

## 解决方案

### 1. 创建 Controller Service

创建了 `charts/ops/templates/service-controller.yaml`，包含：
- Service 名称: `{release-name}-ops-controller-metrics`
- Metrics 端口: 8080 (HTTP)
- Selector: 匹配 controller deployment 的 labels

### 2. 更新 Deployment Labels

在 `charts/ops/templates/deployment.yaml` 中添加了 `control-plane: controller-manager` label，确保 Service 可以正确匹配到 Pod。

### 3. 更新 ServiceMonitor

更新了 `charts/ops/templates/servicemonitor-controller.yaml`：
- 使用 HTTP 协议（因为 Helm chart 部署不使用 kube-rbac-proxy）
- 使用 `metrics` 端口（8080）
- Selector 匹配新创建的 Service

## 两种部署方式的区别

### Kustomize 部署（使用 kube-rbac-proxy）
- Service: `controller-manager-metrics-service`
- 端口: 8443 (HTTPS)
- 协议: HTTPS，需要认证
- 位置: `config/rbac/auth_proxy_service.yaml`

### Helm Chart 部署（直接暴露）
- Service: `{release-name}-ops-controller-metrics`
- 端口: 8080 (HTTP)
- 协议: HTTP，无需认证
- 位置: `charts/ops/templates/service-controller.yaml`

## 验证

部署后验证：

```bash
# 检查 Service
kubectl get svc -n <namespace> | grep controller-metrics

# 检查 ServiceMonitor
kubectl get servicemonitor -n <namespace> | grep controller

# 检查 metrics 端点
kubectl port-forward -n <namespace> svc/<release-name>-ops-controller-metrics 8080:8080
curl http://localhost:8080/metrics
```

