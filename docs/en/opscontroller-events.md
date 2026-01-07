# Ops Event System

## Overview

The Ops project publishes various events through NATS JetStream for monitoring, alerting, and integration. All events follow the CloudEvents standard and are routed through NATS subjects.

## Event Subject Format

### Standard Format

Most event subjects follow this format:

```
ops.clusters.{cluster}.namespaces.{namespace}.{resourceType}.{resourceName}.{observation}
```

Where:
- `{cluster}`: Cluster name (from environment variable `EVENT_CLUSTER`)
- `{namespace}`: Kubernetes namespace
- `{resourceType}`: Resource type (e.g., `hosts`, `clusters`, `taskruns`, etc.)
- `{resourceName}`: Resource name
- `{observation}`: Observation type, possible values and meanings:
  - `status`: Resource status information (e.g., running status, health status, etc.)
  - `metrics`: Metrics data (e.g., CPU, memory, disk usage, etc.)
  - `logs`: Log information
  - `events`: Event information (e.g., Kubernetes events)
  - `traces`: Tracing information (e.g., distributed tracing data)
  - `alerts`: Alert information
  - `findings`: Proactively reported information and status

### Node Event Special Format

Node events use a special format without the `namespaces` part:

```
ops.clusters.{cluster}.nodes.{nodeName}.{observation}
```

**Example:**
```
ops.clusters.mycluster.nodes.mynode.events
```

**Note:**
- Nodes are cluster-level resources and do not belong to any namespace
- Node event subjects omit the `namespaces` part
- Other format rules are the same as the standard format

## Event Types

### 1. Controller Setup Events

**Subject Format:**
```
ops.clusters.{cluster}.namespaces.{namespace}.controllers.{resourceType}.status
```

**Trigger:**
- Published when each Controller starts, indicating the Controller is ready

**Published From:**
- `controllers/host_controller.go`: Host Controller startup
- `controllers/cluster_controller.go`: Cluster Controller startup
- `controllers/task_controller.go`: Task Controller startup
- `controllers/taskrun_controller.go`: TaskRun Controller startup
- `controllers/pipeline_controller.go`: Pipeline Controller startup
- `controllers/pipelinerun_controller.go`: PipelineRun Controller startup
- `controllers/event_controller.go`: Event Controller startup
- `controllers/eventhooks_controller.go`: EventHook Controller startup

**Event Data Structure:**
```json
{
  "cluster": "string",
  "kind": "string"  // Resource type, e.g., "Hosts", "Clusters", "TaskRuns", etc.
}
```

---

### 2. Host Status Events

**Subject Format:**
```
ops.clusters.{cluster}.namespaces.{namespace}.hosts.{hostName}.status
```

**Trigger:**
- When Host resource status is updated (heartbeat status, disk usage, etc.)

**Published From:**
- `controllers/host_controller.go`: When Host status changes

**Event Data Structure:**
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
    // ... other HostStatus fields
  }
}
```

---

### 3. Cluster Status Events

**Subject Format:**
```
ops.clusters.{cluster}.namespaces.{namespace}.clusters.{clusterName}.status
```

**Trigger:**
- When Cluster resource status is updated (Kubernetes version, certificate expiration, heartbeat status, etc.)

**Published From:**
- `controllers/cluster_controller.go`: When Cluster status changes

**Event Data Structure:**
```json
{
  "cluster": "string",
  "server": "string",
  "status": {
    "version": "string",
    "certNotAfterDays": 0,
    "heartStatus": "string",
    "heartTime": "string",
    // ... other ClusterStatus fields
  }
}
```

---

### 4. TaskRun Status Events

**Subject Format:**
```
ops.clusters.{cluster}.namespaces.{namespace}.taskruns.{taskRunName}.status
```

**Trigger:**
- When TaskRun execution completes or status changes

**Published From:**
- `controllers/taskrun_controller.go`: 
  - When TaskRun execution completes (in `run` function)
  - When TaskRun status changes (in Reconcile)

**Event Data Structure:**
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

### 5. PipelineRun Status Events

**Subject Format:**
```
ops.clusters.{cluster}.namespaces.{namespace}.pipelineruns.{pipelineRunName}.status
```

**Trigger:**
- When PipelineRun execution completes or status changes
- When cross-cluster PipelineRun status synchronization completes

**Published From:**
- `controllers/pipelinerun_controller.go`: 
  - When PipelineRun execution completes (in `run` function)
  - When PipelineRun status changes (in Reconcile)
  - When cross-cluster PipelineRun status synchronization completes (in goroutine)

**Event Data Structure:**
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

### 6. Kubernetes Events

**Subject Format:**

For namespaced resources:
```
ops.clusters.{cluster}.namespaces.{namespace}.{resourceKind}s.{resourceName}.events
```

For Node resources (special format, no namespaces):
```
ops.clusters.{cluster}.nodes.{nodeName}.events
```

**Trigger:**
- When Kubernetes native events (events.k8s.io/v1) are created, updated, or deleted
- Only events from the last 2 minutes are published

**Published From:**
- `controllers/event_controller.go`: Watches Kubernetes Event resources

**Event Data Structure:**
```json
{
  "cluster": "string",
  "type": "string",        // Normal, Warning, etc.
  "reason": "string",      // Event reason
  "eventTime": "string",   // ISO 8601 format
  "from": "string",        // Event source (Manager)
  "message": "string"      // Event message
}
```

**Example Subjects:**

Namespaced resources:
```
ops.clusters.mycluster.namespaces.ops-system.pods.my-pod.events
ops.clusters.mycluster.namespaces.ops-system.deployments.my-deployment.events
```

Node resources (special format):
```
ops.clusters.mycluster.nodes.mynode.events
```

---

### 7. Notification Events

**Subject Format:**
```
ops.notifications.providers.{provider}.channels.{channel}.severities.{severity}
```

Where:
- `{provider}`: Notification provider or system name (e.g., `ksyun`, `ai`, etc.)
- `{channel}`: Notification channel type (e.g., `webhook`, `email`, `sms`, etc.)
- `{severity}`: Severity level (e.g., `info`, `warning`, `error`, `critical`, etc.)

**Trigger:**
- Published when the notification system sends notifications

**Published From:**
- Published via API endpoint `/api/v1/namespaces/{namespace}/events/{event}`
- If the event path starts with `ops.`, it will be used directly as the subject without format transformation

**Publishing Method:**
```bash
curl -X POST http://localhost:80/api/v1/namespaces/ops-system/events/ops.notifications.providers.ksyun.channels.webhook.severities.info \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "provider": "ksyun",
    "channel": "webhook",
    "severity": "info",
    "message": "Notification message",
    "title": "Notification title"
  }'
```

**Event Data Structure:**
```json
{
  "provider": "string",
  "channel": "string",
  "severity": "string",
  "message": "string",
  "title": "string",
  "timestamp": "string",
  // ... other notification-related fields
}
```

**Example Subjects:**
```
ops.notifications.providers.ksyun.channels.webhook.severities.info
ops.notifications.providers.ksyun.channels.webhook.severities.warning
ops.notifications.providers.ksyun.channels.webhook.severities.error
ops.notifications.providers.ksyun.channels.email.severities.critical
ops.notifications.providers.AI.channels.webhook.severities.info
```

**Note:**
- Notification events use an independent subject format without `clusters` and `namespaces` parts
- Used for notification system routing and distribution
- Supports combinations of multiple providers, channels, and severity levels
- **Important**: If the event path starts with `ops.`, the API will use the path directly as the NATS subject without any format transformation

---

### 8. Custom Events (via API)

**Subject Format:**
The API uses different processing methods based on the event path:

1. **Event path starting with `ops.`** (direct delivery):
   ```
   ops.{any path}
   ```
   - Uses the path directly as the NATS subject without any format transformation
   - Suitable for notification events and other independent format events

2. **Event path starting with `nodes.`** (node events):
   ```
   ops.clusters.{cluster}.nodes.{nodeName}.{observation}
   ```
   - Converts to node event format without the `namespaces` part

3. **Standard format** (other paths):
   ```
   ops.clusters.{cluster}.namespaces.{namespace}.{eventName}
   ```
   - Automatically adds cluster and namespace prefixes

**Trigger:**
- Manually published via API endpoint `/api/v1/namespaces/{namespace}/events/{event}`

**Published From:**
- `pkg/server/api.go`: `CreateEvent` function

**Event Data Structure:**
- Custom JSON object defined by the caller

**Usage Examples:**
Standard format event:
```bash
curl -X POST http://localhost:80/api/v1/namespaces/ops-system/events/my-custom-event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "message": "Custom event data",
    "level": "info"
  }'
```

Direct delivery format (starting with `ops.`):
```bash
curl -X POST http://localhost:80/api/v1/namespaces/ops-system/events/ops.notifications.providers.ksyun.channels.webhook.severities.info \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "provider": "ksyun",
    "channel": "webhook",
    "severity": "info",
    "message": "Notification message"
  }'
```

Node event format:
```bash
curl -X POST http://localhost:80/api/v1/namespaces/ops-system/events/nodes.mynode.findings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "message": "Node finding",
    "status": "normal"
  }'
```

---

## Event Subscription Examples

### Subscribe to All Events

```bash
nats --user=app --password=${apppassword} sub "ops.>"
```

### Subscribe to Events in a Specific Namespace

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.ops-system.>"
```

### Subscribe to Host Status Events

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.hosts.*.status"
```

### Subscribe to TaskRun Status Events

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.taskruns.*.status"
```

### Subscribe to PipelineRun Status Events

```bash
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.pipelineruns.*.status"
```

### Subscribe to Kubernetes Events

```bash
# Subscribe to Kubernetes events for all namespaced resources
nats --user=app --password=${apppassword} sub "ops.clusters.*.namespaces.*.*.events"

# Subscribe to node events (special format)
nats --user=app --password=${apppassword} sub "ops.clusters.*.nodes.*.events"
```

### Subscribe to Notification Events

```bash
# Subscribe to all notification events
nats --user=app --password=${apppassword} sub "ops.notifications.>"

# Subscribe to notifications from a specific provider
nats --user=app --password=${apppassword} sub "ops.notifications.providers.ksyun.>"

# Subscribe to notifications from a specific channel
nats --user=app --password=${apppassword} sub "ops.notifications.providers.*.channels.webhook.>"

# Subscribe to notifications with a specific severity level
nats --user=app --password=${apppassword} sub "ops.notifications.providers.*.channels.*.severities.error"
```

## Event Format

All events follow the [CloudEvents](https://cloudevents.io/) standard and include these standard fields:

- `id`: Unique event identifier (UUID)
- `source`: Event source (fixed as `https://github.com/shaowenchen/ops`)
- `type`: Event type (e.g., `Host`, `Cluster`, `PipelineRun`, etc.)
- `specversion`: CloudEvents specification version (`1.0`)
- `time`: Event timestamp (ISO 8601 format)
- `data`: Event data (JSON format)
- `subject`: NATS subject (automatically set)

## Environment Variables

The event system depends on the following environment variables:

- `EVENT_CLUSTER`: Cluster name, used to build event subjects
- `EVENT_ENDPOINT`: NATS server address (e.g., `nats://nats:4222`)

## Notes

1. **Event Deduplication**: NATS JetStream supports message deduplication via `dupe-window` configuration (default 2 minutes)
2. **Event Retention**: Event retention time is determined by Stream configuration (default 24 hours)
3. **Event Ordering**: Events on the same subject are published in chronological order
4. **Async Publishing**: Most events are published asynchronously via goroutines, not blocking the main flow
5. **Event Filtering**: Kubernetes events only publish events from the last 2 minutes to avoid historical event interference

## Related Documentation

- [NATS Configuration](./nats.md)
- [EventHook Usage](./opscontroller.md#eventhook)

