# Ops 事件系统

## 概述

Ops 项目通过 NATS JetStream 发布各种事件，用于监控、告警和集成。所有事件都遵循 CloudEvents 标准，并通过 NATS 主题（Subject）进行路由。

## 事件主题格式

### 标准格式

大多数事件主题遵循以下格式：

```
ops.clusters.{cluster}.namespaces.{namespace}.{resourceType}.{resourceName}.{observation}
```

其中：
- `{cluster}`: 集群名称（从环境变量 `EVENT_CLUSTER` 获取）
- `{namespace}`: Kubernetes 命名空间
- `{resourceType}`: 资源类型（如 `hosts`, `clusters`, `taskruns` 等）
- `{resourceName}`: 资源名称
- `{observation}`: 观测类型，可能的值及含义：
  - `status`: 资源状态信息（如运行状态、健康状态等）
  - `metrics`: 指标数据（如 CPU、内存、磁盘使用率等）
  - `logs`: 日志信息
  - `events`: 事件信息（如 Kubernetes 事件）
  - `traces`: 追踪信息（如分布式追踪数据）
  - `alerts`: 告警信息
  - `findings`: 主动上报的信息和状态

### 节点事件特殊格式

节点（Node）事件使用特殊格式，不包含 `namespaces` 部分：

```
ops.clusters.{cluster}.nodes.{nodeName}.{observation}
```

**示例：**
```
ops.clusters.mycluster.nodes.mynode.events
```

**说明：**
- 节点是集群级别的资源，不属于任何命名空间
- 节点事件的主题格式省略了 `namespaces` 部分
- 其他格式规则与标准格式相同

## 事件类型列表

### 1. Controller 设置事件

**主题格式：**
```
ops.clusters.{cluster}.namespaces.{namespace}.controllers.{resourceType}.status
```

**触发时机：**
- 各个 Controller 启动时发布，表示该 Controller 已就绪

**发布位置：**
- `controllers/host_controller.go`: Host Controller 启动
- `controllers/cluster_controller.go`: Cluster Controller 启动
- `controllers/task_controller.go`: Task Controller 启动
- `controllers/taskrun_controller.go`: TaskRun Controller 启动
- `controllers/pipeline_controller.go`: Pipeline Controller 启动
- `controllers/pipelinerun_controller.go`: PipelineRun Controller 启动
- `controllers/event_controller.go`: Event Controller 启动
- `controllers/eventhooks_controller.go`: EventHook Controller 启动

**事件数据结构：**
```json
{
  "cluster": "string",
  "kind": "string"  // 资源类型，如 "Hosts", "Clusters", "TaskRuns" 等
}
```

---

### 2. Host 状态事件

**主题格式：**
```
ops.clusters.{cluster}.namespaces.{namespace}.hosts.{hostName}.status
```

**触发时机：**
- Host 资源状态更新时（心跳状态、磁盘使用率等）

**发布位置：**
- `controllers/host_controller.go`: Host 状态变更时

**事件数据结构：**
```json
{
  "cluster": "string",
  "address": "string",
  "port": 0,
  "username": "string",
  "status": {
    "hostname": "string",
    "diskUsagePercent": "string",
    "heartStatus": "string",
    "heartTime": "string",
    // ... 其他 HostStatus 字段
  }
}
```

---

### 3. Cluster 状态事件

**主题格式：**
```
ops.clusters.{cluster}.namespaces.{namespace}.clusters.{clusterName}.status
```

**触发时机：**
- Cluster 资源状态更新时（Kubernetes 版本、证书过期时间、心跳状态等）

**发布位置：**
- `controllers/cluster_controller.go`: Cluster 状态变更时

**事件数据结构：**
```json
{
  "cluster": "string",
  "server": "string",
  "status": {
    "version": "string",
    "certNotAfterDays": 0,
    "heartStatus": "string",
    "heartTime": "string",
    // ... 其他 ClusterStatus 字段
  }
}
```

---

### 4. TaskRun 状态事件

**主题格式：**
```
ops.clusters.{cluster}.namespaces.{namespace}.taskruns.{taskRunName}.status
```

**触发时机：**
- TaskRun 执行完成或状态变更时

**发布位置：**
- `controllers/taskrun_controller.go`: 
  - TaskRun 执行完成时（`run` 函数中）
  - TaskRun 状态变更时（Reconcile 中）

**事件数据结构：**
```json
{
  "cluster": "string",
  "taskRef": "string",
  "desc": "string",
  "variables": {
    "key": "value"
  },
  "runStatus": "string",
  "startTime": "string",
  "taskRunNodeStatus": [
    {
      "nodeName": "string",
      "runStatus": "string",
      "taskRunStep": [
        {
          "stepName": "string",
          "stepContent": "string",
          "stepOutput": "string",
          "stepStatus": "string"
        }
      ]
    }
  ]
}
```

---

### 5. PipelineRun 状态事件

**主题格式：**
```
ops.clusters.{cluster}.namespaces.{namespace}.pipelineruns.{pipelineRunName}.status
```

**触发时机：**
- PipelineRun 执行完成或状态变更时
- 跨集群 PipelineRun 状态同步完成时

**发布位置：**
- `controllers/pipelinerun_controller.go`: 
  - PipelineRun 执行完成时（`run` 函数中）
  - PipelineRun 状态变更时（Reconcile 中）
  - 跨集群 PipelineRun 状态同步完成时（goroutine 中）

**事件数据结构：**
```json
{
  "cluster": "string",
  "pipelineRef": "string",
  "desc": "string",
  "variables": {
    "key": "value"
  },
  "runStatus": "string",
  "startTime": "string",
  "pipelineRunStatus": [
    {
      "taskName": "string",
      "taskRef": "string",
      "taskRunStatus": {
        "runStatus": "string",
        "taskRunNodeStatus": []
      }
    }
  ]
}
```

---

### 6. Kubernetes 事件

**主题格式：**

对于命名空间资源：
```
ops.clusters.{cluster}.namespaces.{namespace}.{resourceKind}s.{resourceName}.events
```

对于节点（Node）资源（特殊格式，无 namespaces）：
```
ops.clusters.{cluster}.nodes.{nodeName}.events
```

**触发时机：**
- Kubernetes 原生事件（events.k8s.io/v1）的创建、更新、删除时
- 仅发布最近 2 分钟内的事件

**发布位置：**
- `controllers/event_controller.go`: 监听 Kubernetes Event 资源

**事件数据结构：**
```json
{
  "cluster": "string",
  "type": "string",        // Normal, Warning 等
  "reason": "string",      // 事件原因
  "eventTime": "string",   // ISO 8601 格式
  "from": "string",        // 事件来源（Manager）
  "message": "string"      // 事件消息
}
```

**示例主题：**

命名空间资源：
```
ops.clusters.mycluster.namespaces.ops-system.pods.my-pod.events
ops.clusters.mycluster.namespaces.ops-system.deployments.my-deployment.events
```

节点资源（特殊格式）：
```
ops.clusters.mycluster.nodes.mynode.events
```

---

### 7. 自定义事件（通过 API）

**主题格式：**
```
ops.clusters.{cluster}.namespaces.{namespace}.{eventName}
```

**触发时机：**
- 通过 API 接口 `/api/v1/namespaces/{namespace}/events/{event}` 手动发布

**发布位置：**
- `pkg/server/api.go`: `CreateEvent` 函数

**事件数据结构：**
- 由调用方自定义，可以是任意 JSON 对象

**使用示例：**
```bash
curl -X POST http://localhost:80/api/v1/namespaces/ops-system/events/my-custom-event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "message": "Custom event data",
    "level": "info"
  }'
```

---

## 事件订阅示例

### 订阅所有事件

```bash
nats --user=app --password=${apppassword} sub "ops.>"
```

### 订阅特定命名空间的事件

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.ops-system.>"
```

### 订阅 Host 状态事件

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.hosts.*.status"
```

### 订阅 TaskRun 状态事件

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.taskruns.*.status"
```

### 订阅 PipelineRun 状态事件

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.pipelineruns.*.status"
```

### 订阅 Kubernetes 事件

```bash
# 订阅所有命名空间资源的 Kubernetes 事件
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.*.events"

# 订阅节点事件（特殊格式）
nats --user=app --password=${apppassword} sub "ops.clusters.*.nodes.*.events"
```

## 事件格式说明

所有事件都遵循 [CloudEvents](https://cloudevents.io/) 标准，包含以下标准字段：

- `id`: 事件唯一标识符（UUID）
- `source`: 事件来源（固定为 `https://github.com/shaowenchen/ops`）
- `type`: 事件类型（如 `Host`, `Cluster`, `PipelineRun` 等）
- `specversion`: CloudEvents 规范版本（`1.0`）
- `time`: 事件时间戳（ISO 8601 格式）
- `data`: 事件数据（JSON 格式）
- `subject`: NATS 主题（自动设置）

## 环境变量配置

事件系统依赖以下环境变量：

- `EVENT_CLUSTER`: 集群名称，用于构建事件主题
- `EVENT_ENDPOINT`: NATS 服务器地址（如 `nats://nats:4222`）

## 注意事项

1. **事件去重**: NATS JetStream 支持消息去重，通过 `dupe-window` 配置（默认 2 分钟）
2. **事件保留**: 根据 Stream 配置决定事件保留时间（默认 24 小时）
3. **事件顺序**: 同一主题的事件按时间顺序发布
4. **异步发布**: 大部分事件通过 goroutine 异步发布，不会阻塞主流程
5. **事件过滤**: Kubernetes 事件仅发布最近 2 分钟内的事件，避免历史事件干扰

## 相关文档

- [NATS 配置文档](./nats.md)
- [EventHook 使用文档](./opscontroller.md#eventhook)

