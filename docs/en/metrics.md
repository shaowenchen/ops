# Ops Metrics

Ops exposes Prometheus metrics for monitoring controller and server components.

## Metrics Endpoints

- **Controller**: `:9090/metrics`
- **Server**: `:9090/metrics`

## Controller Metrics

### Resource Info Metrics

These metrics expose resource information during each reconcile.

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_task_info` | namespace, name, desc, host, runtime_image | Task resource info |
| `ops_controller_pipeline_info` | namespace, name, desc | Pipeline resource info |
| `ops_controller_host_info` | namespace, name, address, hostname, distribution, arch, cpu_total, mem_total, disk_total, accelerator_vendor, accelerator_model, accelerator_count, heart_status | Host resource info |
| `ops_controller_cluster_info` | namespace, name, server, version, node, pod, running_pod, heart_status | Cluster resource info |
| `ops_controller_eventhooks_info` | namespace, name, type, subject, url | EventHooks resource info |

### TaskRun Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_taskrun_info` | namespace, name, taskref, crontab, status | TaskRun resource info and status |
| `ops_controller_taskrun_start_time` | namespace, name, taskref | TaskRun start time (unix timestamp) |
| `ops_controller_taskrun_end_time` | namespace, name, taskref | TaskRun end time (unix timestamp) |
| `ops_controller_taskrun_duration_seconds` | namespace, name, taskref, status | TaskRun duration in seconds |

### PipelineRun Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_pipelinerun_info` | namespace, name, pipelineref, crontab, status | PipelineRun resource info and status |
| `ops_controller_pipelinerun_start_time` | namespace, name, pipelineref | PipelineRun start time (unix timestamp) |
| `ops_controller_pipelinerun_end_time` | namespace, name, pipelineref | PipelineRun end time (unix timestamp) |
| `ops_controller_pipelinerun_duration_seconds` | namespace, name, pipelineref, status | PipelineRun duration in seconds |

### Run Count Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_taskref_run_total` | namespace, taskref, status | Total number of TaskRef runs |
| `ops_controller_pipelineref_run_total` | namespace, pipelineref, status | Total number of PipelineRef runs |

### EventHooks Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_eventhooks_trigger_total` | namespace, eventhook_name, keyword, event_id, status | EventHooks trigger count with matched keyword and event ID |

### Reconcile Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_reconcile_total` | controller, namespace, result | Total number of reconcile operations |
| `ops_controller_reconcile_errors_total` | controller, namespace, error_type | Total number of reconcile errors |

### Controller Resource Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_controller_resource_goroutines` | pod | Controller number of goroutines |
| `ops_controller_resource_cpu_usage_seconds_total` | pod | Controller CPU usage in seconds (cumulative, read from cgroup) |
| `ops_controller_resource_memory_usage_bytes` | pod | Controller memory usage in bytes (read from cgroup) |
| `ops_controller_uptime_seconds` | pod | Controller uptime in seconds |
| `ops_controller_info` | pod, version, build_date | Controller information |

## Server Metrics

### Resource Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_server_resource_goroutines` | pod | Server number of goroutines |
| `ops_server_resource_cpu_usage_seconds_total` | pod | Server CPU usage in seconds (cumulative, read from cgroup) |
| `ops_server_resource_memory_usage_bytes` | pod | Server memory usage in bytes (read from cgroup) |

### Throughput Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_server_throughput_http_requests_total` | method, path, status_code | Total number of HTTP requests |
| `ops_server_throughput_api_requests_total` | endpoint, namespace, status | Total number of API requests |
| `ops_server_throughput_api_errors_total` | endpoint, namespace, error_type | Total number of API errors |

### Server Info Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `ops_server_info` | pod, version, build_date | Server information |
| `ops_server_uptime_seconds` | pod | Server uptime in seconds |

## Example Queries

### Get all running TaskRuns

```promql
ops_controller_taskrun_info{status="Running"}
```

### Get Task run count by status

```promql
sum by (taskref, status) (ops_controller_taskref_run_total)
```

### Get Host list with status

```promql
ops_controller_host_info
```

### Get EventHooks triggers by keyword

```promql
sum by (eventhook_name, keyword) (ops_controller_eventhooks_trigger_total)
```

### Get TaskRun duration

```promql
ops_controller_taskrun_duration_seconds
```

### Get Controller CPU usage rate

```promql
rate(ops_controller_resource_cpu_usage_seconds_total{pod="xxx"}[5m])
```

### Get Controller memory usage

```promql
ops_controller_resource_memory_usage_bytes{pod="xxx"}
```

### Get Server CPU usage rate

```promql
rate(ops_server_resource_cpu_usage_seconds_total{pod="xxx"}[5m])
```

### Get Server memory usage

```promql
ops_server_resource_memory_usage_bytes{pod="xxx"}
```

