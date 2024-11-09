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

### 安装 Helm

```bash
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

### 添加 Helm 仓库

```bash
helm repo add ops https://www.chenshaowen.com/ops/charts
```

### 安装 ops-controller-manager

```bash
helm install myops ops/ops --version 1.0.0 --namespace ops-system --create-namespace
```

### 查看安装结果

```bash
kubectl get pods -n ops-system
```

### 卸载

```bash
helm -n ops-system uninstall myops
```

## 需要注意的是

ops-controller-manager 默认只会处理 `ops-system` 命名空间下的 CRD 资源。

如果需要变更，可以修改 Env 中 `ACTIVE_NAMESPACE` 的值，指定某一个命令空间，如果为空，则表示处理所有命名空间。
