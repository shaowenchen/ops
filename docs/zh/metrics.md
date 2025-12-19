# Ops 监控指标

Ops 暴露 Prometheus 指标用于监控 controller 和 server 组件。

## 指标端点

- **Controller**: `:9090/metrics`
- **Server**: `:9090/metrics`

## Controller 指标

### 资源信息指标

这些指标在每次 reconcile 时暴露资源信息。

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_task_info` | namespace, name, desc, host, runtime_image | Task 资源信息 |
| `ops_controller_pipeline_info` | namespace, name, desc | Pipeline 资源信息 |
| `ops_controller_host_info` | namespace, name, address, hostname, distribution, arch, cpu_total, mem_total, disk_total, accelerator_vendor, accelerator_model, accelerator_count, heart_status | Host 资源信息 |
| `ops_controller_cluster_info` | namespace, name, server, version, node, pod, running_pod, heart_status | Cluster 资源信息 |
| `ops_controller_eventhooks_info` | namespace, name, type, subject, url | EventHooks 资源信息 |

### TaskRun 指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_taskrun_info` | namespace, name, taskref, crontab, status | TaskRun 资源信息和状态 |
| `ops_controller_taskrun_start_time` | namespace, name, taskref | TaskRun 开始时间（unix 时间戳） |
| `ops_controller_taskrun_end_time` | namespace, name, taskref | TaskRun 结束时间（unix 时间戳） |
| `ops_controller_taskrun_duration_seconds` | namespace, name, taskref, status | TaskRun 运行时长（秒） |

### PipelineRun 指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_pipelinerun_info` | namespace, name, pipelineref, crontab, status | PipelineRun 资源信息和状态 |
| `ops_controller_pipelinerun_start_time` | namespace, name, pipelineref | PipelineRun 开始时间（unix 时间戳） |
| `ops_controller_pipelinerun_end_time` | namespace, name, pipelineref | PipelineRun 结束时间（unix 时间戳） |
| `ops_controller_pipelinerun_duration_seconds` | namespace, name, pipelineref, status | PipelineRun 运行时长（秒） |

### 运行次数指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_taskref_run_total` | namespace, taskref, status | Task 运行总次数 |
| `ops_controller_pipelineref_run_total` | namespace, pipelineref, status | Pipeline 运行总次数 |

### EventHooks 指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_eventhooks_trigger_total` | namespace, eventhook_name, keyword, event_id, status | EventHooks 触发次数，包含匹配的关键字和事件 ID |

### Reconcile 指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_reconcile_total` | controller, namespace, result | reconcile 操作总次数 |
| `ops_controller_reconcile_errors_total` | controller, namespace, error_type | reconcile 错误总次数 |

### Controller 资源指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_controller_resource_goroutines` | pod | Controller goroutine 数量 |
| `ops_controller_resource_cpu_usage_seconds_total` | pod | Controller CPU 使用量（秒，累计值，从 cgroup 读取） |
| `ops_controller_resource_memory_usage_bytes` | pod | Controller 内存使用量（字节，从 cgroup 读取） |
| `ops_controller_uptime_seconds` | pod | Controller 运行时间（秒） |
| `ops_controller_info` | pod, version, build_date | Controller 信息 |

## Server 指标

### 资源指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_server_resource_goroutines` | pod | Server goroutine 数量 |
| `ops_server_resource_cpu_usage_seconds_total` | pod | Server CPU 使用量（秒，累计值，从 cgroup 读取） |
| `ops_server_resource_memory_usage_bytes` | pod | Server 内存使用量（字节，从 cgroup 读取） |

### 吞吐量指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_server_throughput_http_requests_total` | method, path, status_code | HTTP 请求总数 |
| `ops_server_throughput_api_requests_total` | endpoint, namespace, status | API 请求总数 |
| `ops_server_throughput_api_errors_total` | endpoint, namespace, error_type | API 错误总数 |

### Server 信息指标

| 指标 | 标签 | 描述 |
|------|------|------|
| `ops_server_info` | pod, version, build_date | Server 信息 |
| `ops_server_uptime_seconds` | pod | Server 运行时间（秒） |

## 查询示例

### 获取所有运行中的 TaskRun

```promql
ops_controller_taskrun_info{status="Running"}
```

### 按状态统计 Task 运行次数

```promql
sum by (taskref, status) (ops_controller_taskref_run_total)
```

### 获取 Host 列表及状态

```promql
ops_controller_host_info
```

### 按关键字统计 EventHooks 触发次数

```promql
sum by (eventhook_name, keyword) (ops_controller_eventhooks_trigger_total)
```

### 获取 TaskRun 运行时长

```promql
ops_controller_taskrun_duration_seconds
```

### 获取 Controller CPU 使用率

```promql
rate(ops_controller_resource_cpu_usage_seconds_total{pod="xxx"}[5m])
```

### 获取 Controller 内存使用量

```promql
ops_controller_resource_memory_usage_bytes{pod="xxx"}
```

### 获取 Server CPU 使用率

```promql
rate(ops_server_resource_cpu_usage_seconds_total{pod="xxx"}[5m])
```

### 获取 Server 内存使用量

```promql
ops_server_resource_memory_usage_bytes{pod="xxx"}
```

