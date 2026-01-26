# EventHooks 使用指南

## 概述

EventHooks 是 Ops 提供的事件通知机制，允许您根据事件的关键词匹配规则，将事件转发到不同的通知渠道。支持多种通知类型，包括 Webhook、协作文档、事件转发和 Elasticsearch。

## 基本概念

### EventHooks 资源

EventHooks 是一个 Kubernetes CRD 资源，用于定义事件通知规则：

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: my-eventhook
  namespace: ops-system
spec:
  type: webhook              # 通知类型
  subject: ops.clusters.*.namespaces.*.pods.*.events  # 订阅的事件主题
  url: https://example.com/webhook  # 通知目标 URL
  keywords:                  # 关键词匹配规则
    include:
      - "Error"
    matchType: CONTAINS
```

### 字段说明

- **type**: 通知类型，支持的值：
  - `webhook`: HTTP Webhook 通知
  - `xiezuo`: 协作文档通知
  - `event`: 事件转发（转发到 NATS 主题）
  - `elasticsearch`: Elasticsearch 索引
- **subject**: 订阅的事件主题，支持通配符（如 `ops.clusters.*.namespaces.*.pods.*.events`）
- **url**: 通知目标地址，根据不同的通知类型有不同的格式要求
- **keywords**: 关键词匹配配置
- **options**: 额外的配置选项，不同通知类型支持不同的选项

## 关键词匹配

### 匹配模式 (matchMode)

- **ANY**（默认）: 只要包含列表中的任意一个关键词即匹配
- **ALL**: 必须包含列表中的所有关键词才匹配

### 匹配类型 (matchType)

- **CONTAINS**（默认）: 字符串包含匹配
- **EXACT**: 精确匹配
- **REGEX**: 正则表达式匹配

### 示例

#### 示例 1: 包含匹配

```yaml
keywords:
  include:
    - "Error"
    - "Failed"
  matchMode: ANY
  matchType: CONTAINS
```

匹配包含 "Error" 或 "Failed" 的事件。

#### 示例 2: 正则表达式匹配

```yaml
keywords:
  include:
    - "(?=.*(kube|etcd|calico|csi|fluid)).*(BackOff|OOMKilled|Evicted|NetworkNotReady|Unhealthy|Error|Failed|ImagePullBackOff).*"
  matchType: REGEX
```

使用正则表达式匹配：
- 必须包含：`kube`、`etcd`、`calico`、`csi` 或 `fluid` 中的任意一个
- 且必须包含：`BackOff`、`OOMKilled`、`Evicted` 等关键词中的任意一个

#### 示例 3: 排除匹配

```yaml
keywords:
  include:
    - "Error"
  exclude:
    - "healthcheck"
    - "test"
  matchType: CONTAINS
```

匹配包含 "Error" 但不包含 "healthcheck" 或 "test" 的事件。

## 通知类型详解

### 1. Webhook 通知 (webhook)

将事件数据以 JSON 格式发送到指定的 HTTP Webhook URL。

**配置示例：**

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: webhook-notification
  namespace: ops-system
spec:
  type: webhook
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: https://example.com/api/webhook
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

**发送的数据格式：**

事件的可读字符串格式（包含事件的所有字段信息）。

### 2. 协作文档通知 (xiezuo)

发送到协作文档系统的 Webhook。

**配置示例：**

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: xiezuo-notification
  namespace: ops-system
spec:
  type: xiezuo
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: https://365.kdocs.cn/woa/api/v1/webhook/send?key=xxxx
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

**数据格式：**

如果数据已经是 XiezuoBody 格式，直接发送；否则转换为文本格式发送。

### 3. 事件转发 (event)

将事件转发到另一个 NATS 主题。支持通配符替换。

**配置示例：**

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: event-forward
  namespace: ops-system
spec:
  type: event
  subject: ops.clusters.*.namespaces.*.pods.*.events
  # URL 中使用通配符 *，会自动替换为原始事件 subject 中对应位置的值
  url: ops.clusters.*.namespaces.*.pods.*.alerts
  keywords:
    include:
      - "(?=.*(kube|etcd|calico)).*(BackOff|OOMKilled|Evicted|Error|Failed).*"
    matchType: REGEX
```

**通配符替换示例：**

- 原始事件 subject: `ops.clusters.cluster1.namespaces.ns1.pods.pod1.events`
- URL 模板: `ops.clusters.*.namespaces.*.pods.*.alerts`
- 替换后: `ops.clusters.cluster1.namespaces.ns1.pods.pod1.alerts`

**节点事件支持：**

也支持节点事件格式：
- 原始事件: `ops.clusters.cluster1.nodes.node1.events`
- URL 模板: `ops.clusters.*.nodes.*.alerts`
- 替换后: `ops.clusters.cluster1.nodes.node1.alerts`

**注意事项：**

- 转发的事件会生成新的 ID 和时间戳
- 需要设置环境变量 `EVENT_ENDPOINT` 指定 NATS 服务器地址

### 4. Elasticsearch 通知 (elasticsearch)

将事件数据索引到 Elasticsearch。

**配置示例：**

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: elasticsearch-notification
  namespace: ops-system
spec:
  type: elasticsearch
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: http://elasticsearch:9200/ops-events/_doc
  options:
    username: elastic      # Elasticsearch 用户名（可选）
    password: changeme     # Elasticsearch 密码（可选）
    index: ops-events      # 覆盖 URL 中的索引名（可选）
    id: ""                 # 文档 ID（可选，不指定则自动生成）
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

**URL 格式：**

- `http://host:port/index/_doc` - 自动生成文档 ID
- `http://host:port/index/_doc/doc-id` - 指定文档 ID
- `http://host:port` - 需要在 options 中指定 index

**基于日期的索引命名：**

可以在索引名中使用日期占位符来创建基于时间的索引。支持的占位符：

- `{date}` 或 `{YYYY.MM.DD}` -> `2024.01.19`
- `{YYYY-MM-DD}` -> `2024-01-19`
- `{YYYYMMDD}` -> `20240119`
- `{YYYY.MM}` -> `2024.01` (月度索引)
- `{YYYY-MM}` -> `2024-01` (月度索引)
- `{YYYY}` -> `2024` (年度索引)

**使用日期索引的示例：**

```yaml
spec:
  type: elasticsearch
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: http://elasticsearch:9200
  options:
    index: ops-events-{YYYY.MM.DD}  # 创建类似 ops-events-2024.01.19 的索引
    username: elastic
    password: changeme
```

这将创建按日期命名的索引，例如：
- `ops-events-2024.01.19`
- `ops-events-2024.01.20`
- 等等

**索引的文档结构：**

索引的文档只包含来自 `event.Data()` 的原始事件数据，不添加任何额外的元数据或扩展字段。

**示例：**

如果原始事件数据是：
```json
{
  "cluster": "cluster1",
  "namespace": "ns1",
  "pod": "pod1",
  "type": "Warning",
  "reason": "BackOff",
  "message": "事件消息"
}
```

那么索引的文档将完全相同：
```json
{
  "cluster": "cluster1",
  "namespace": "ns1",
  "pod": "pod1",
  "type": "Warning",
  "reason": "BackOff",
  "message": "事件消息"
}
```

**注意：** 只索引原始事件数据。不会添加任何元数据字段（如 `@timestamp`、`event_id`、`event_type` 等）或扩展字段（如 `ext_*`）。

**支持的选项：**

- `username`: Elasticsearch 用户名（用于 Basic 认证）
- `password`: Elasticsearch 密码（用于 Basic 认证）
- `index`: 覆盖 URL 中的索引名（支持日期占位符，如 `{YYYY.MM.DD}`）
- `id`: 文档 ID（如果 URL 中未指定）

## 完整示例

### 示例 1: 将 Pod 错误事件发送到 Webhook

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: pod-errors-to-webhook
  namespace: ops-system
spec:
  type: webhook
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: https://alerting.example.com/webhook
  keywords:
    include:
      - "(?=.*(kube|etcd|calico|csi|fluid)).*(BackOff|OOMKilled|Evicted|NetworkNotReady|Unhealthy|Error|Failed|ImagePullBackOff).*"
    matchType: REGEX
```

### 示例 2: 将事件转发到告警主题

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: convert-events-to-alerts
  namespace: ops-system
spec:
  type: event
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: ops.clusters.*.namespaces.*.pods.*.alerts
  keywords:
    include:
      - "(?=.*(kube|etcd|calico|csi|fluid)).*(BackOff|OOMKilled|Evicted|NetworkNotReady|Unhealthy|Error|Failed|ImagePullBackOff).*"
    matchType: REGEX
```

### 示例 3: 将事件索引到 Elasticsearch

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: events-to-elasticsearch
  namespace: ops-system
spec:
  type: elasticsearch
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: http://elasticsearch:9200/ops-events/_doc
  options:
    username: elastic
    password: changeme
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

### 示例 3b: 使用日期索引将事件索引到 Elasticsearch

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: events-to-elasticsearch-daily
  namespace: ops-system
spec:
  type: elasticsearch
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: http://elasticsearch:9200
  options:
    index: ops-events-{YYYY.MM.DD}  # 创建按日期命名的索引，如 ops-events-2024.01.19
    username: elastic
    password: changeme
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

### 示例 4: 节点事件处理

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: node-events-to-alerts
  namespace: ops-system
spec:
  type: event
  subject: ops.clusters.*.nodes.*.events
  url: ops.clusters.*.nodes.*.alerts
  keywords:
    include:
      - "(?=.*(kube|etcd|calico)).*(BackOff|OOMKilled|Error|Failed).*"
    matchType: REGEX
```

## 最佳实践

1. **使用正则表达式进行复杂匹配**：当需要匹配多个条件时，使用正则表达式可以更精确地控制匹配规则。

2. **合理使用排除规则**：使用 `exclude` 可以过滤掉不需要的事件，减少误报。

3. **事件转发使用通配符**：在事件转发场景中，使用通配符可以保持事件的主题结构，便于后续处理。

4. **Elasticsearch 索引命名**：建议使用有意义的索引名，可以考虑按日期或集群名称组织索引。

5. **监控 EventHooks 状态**：通过 Prometheus 指标监控 EventHooks 的触发情况，及时发现配置问题。

## 故障排查

### 查看 EventHooks 状态

```bash
kubectl get eventhooks -n ops-system
kubectl describe eventhooks <name> -n ops-system
```

### 查看 Controller 日志

```bash
kubectl logs -n ops-system deployment/ops-controller-manager | grep eventhook
```

### 检查指标

```bash
# 查看 EventHooks 触发次数
curl http://localhost:8080/metrics | grep ops_controller_eventhooks_status
```

### 常见问题

1. **事件未触发**：
   - 检查 `subject` 是否正确匹配事件主题
   - 检查 `keywords` 配置是否正确
   - 查看 Controller 日志确认是否有错误

2. **Webhook 发送失败**：
   - 检查 URL 是否可访问
   - 检查网络连接
   - 查看 Controller 日志中的错误信息

3. **Elasticsearch 索引失败**：
   - 检查 Elasticsearch 连接
   - 验证认证信息
   - 检查索引权限

## 相关文档

- [事件系统文档](./opscontroller-events.md)
- [指标文档](./opscontroller-metrics.md)
