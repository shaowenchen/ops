# EventHooks Usage Guide

## Overview

EventHooks is an event notification mechanism provided by Ops that allows you to forward events to different notification channels based on keyword matching rules. It supports multiple notification types including Webhook, Collaborative Document, Event Forwarding, and Elasticsearch.

## Basic Concepts

### EventHooks Resource

EventHooks is a Kubernetes CRD resource used to define event notification rules:

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: my-eventhook
  namespace: ops-system
spec:
  type: webhook              # Notification type
  subject: ops.clusters.*.namespaces.*.pods.*.events  # Event subject to subscribe
  url: https://example.com/webhook  # Notification target URL
  keywords:                  # Keyword matching rules
    include:
      - "Error"
    matchType: CONTAINS
```

### Field Description

- **type**: Notification type, supported values:
  - `webhook`: HTTP Webhook notification
  - `xiezuo`: Collaborative document notification
  - `event`: Event forwarding (forward to NATS subject)
  - `elasticsearch`: Elasticsearch indexing
- **subject**: Event subject to subscribe, supports wildcards (e.g., `ops.clusters.*.namespaces.*.pods.*.events`)
- **url**: Notification target address, different formats required for different notification types
- **keywords**: Keyword matching configuration
- **options**: Additional configuration options, different notification types support different options

## Keyword Matching

### Match Mode (matchMode)

- **ANY** (default): Matches if any keyword in the list is found
- **ALL**: Matches only if all keywords in the list are found

### Match Type (matchType)

- **CONTAINS** (default): String contains matching
- **EXACT**: Exact matching
- **REGEX**: Regular expression matching

### Examples

#### Example 1: Contains Matching

```yaml
keywords:
  include:
    - "Error"
    - "Failed"
  matchMode: ANY
  matchType: CONTAINS
```

Matches events containing "Error" or "Failed".

#### Example 2: Regular Expression Matching

```yaml
keywords:
  include:
    - "(?=.*(kube|etcd|calico|csi|fluid)).*(BackOff|OOMKilled|Evicted|NetworkNotReady|Unhealthy|Error|Failed|ImagePullBackOff).*"
  matchType: REGEX
```

Using regular expression to match:
- Must contain: any one of `kube`, `etcd`, `calico`, `csi`, or `fluid`
- And must contain: any one of `BackOff`, `OOMKilled`, `Evicted`, etc.

#### Example 3: Exclude Matching

```yaml
keywords:
  include:
    - "Error"
  exclude:
    - "healthcheck"
    - "test"
  matchType: CONTAINS
```

Matches events containing "Error" but not containing "healthcheck" or "test".

## Notification Types

### 1. Webhook Notification (webhook)

Sends event data in JSON format to the specified HTTP Webhook URL.

**Configuration Example:**

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

**Data Format Sent:**

Readable string format of the event (contains all event field information).

### 2. Collaborative Document Notification (xiezuo)

Sends to collaborative document system webhook.

**Configuration Example:**

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

**Data Format:**

If data is already in XiezuoBody format, send directly; otherwise convert to text format.

### 3. Event Forwarding (event)

Forwards events to another NATS subject. Supports wildcard replacement.

**Configuration Example:**

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: EventHooks
metadata:
  name: event-forward
  namespace: ops-system
spec:
  type: event
  subject: ops.clusters.*.namespaces.*.pods.*.events
  # Use wildcard * in URL, automatically replaced with corresponding values from original event subject
  url: ops.clusters.*.namespaces.*.pods.*.alerts
  keywords:
    include:
      - "(?=.*(kube|etcd|calico)).*(BackOff|OOMKilled|Evicted|Error|Failed).*"
    matchType: REGEX
```

**Wildcard Replacement Example:**

- Original event subject: `ops.clusters.cluster1.namespaces.ns1.pods.pod1.events`
- URL template: `ops.clusters.*.namespaces.*.pods.*.alerts`
- Replaced: `ops.clusters.cluster1.namespaces.ns1.pods.pod1.alerts`

**Node Event Support:**

Also supports node event format:
- Original event: `ops.clusters.cluster1.nodes.node1.events`
- URL template: `ops.clusters.*.nodes.*.alerts`
- Replaced: `ops.clusters.cluster1.nodes.node1.alerts`

**Notes:**

- Forwarded events will have new ID and timestamp
- Need to set environment variable `EVENT_ENDPOINT` to specify NATS server address

### 4. Elasticsearch Notification (elasticsearch)

Indexes event data to Elasticsearch.

**Configuration Example:**

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
    username: elastic      # Elasticsearch username (optional)
    password: changeme     # Elasticsearch password (optional)
    index: ops-events      # Override index name from URL (optional)
    id: ""                 # Document ID (optional, auto-generated if not specified)
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

**URL Format:**

- `http://host:port/index/_doc` - Auto-generate document ID
- `http://host:port/index/_doc/doc-id` - Specify document ID
- `http://host:port` - Need to specify index in options

**Date-based Index Names:**

You can use date placeholders in the index name to create time-based indices. Supported placeholders:

- `{date}` or `{YYYY.MM.DD}` -> `2024.01.19`
- `{YYYY-MM-DD}` -> `2024-01-19`
- `{YYYYMMDD}` -> `20240119`
- `{YYYY.MM}` -> `2024.01` (monthly index)
- `{YYYY-MM}` -> `2024-01` (monthly index)
- `{YYYY}` -> `2024` (yearly index)

**Example with Date-based Index:**

```yaml
spec:
  type: elasticsearch
  subject: ops.clusters.*.namespaces.*.pods.*.events
  url: http://elasticsearch:9200
  options:
    index: ops-events-{YYYY.MM.DD}  # Creates index like ops-events-2024.01.19
    username: elastic
    password: changeme
```

This will create daily indices like:
- `ops-events-2024.01.19`
- `ops-events-2024.01.20`
- etc.

**Indexed Document Structure:**

The indexed document contains only the original event data from `event.Data()`. No additional metadata or extension fields are added.

**Example:**

If the original event data is:
```json
{
  "cluster": "cluster1",
  "namespace": "ns1",
  "pod": "pod1",
  "type": "Warning",
  "reason": "BackOff",
  "message": "Event message"
}
```

Then the indexed document will be exactly the same:
```json
{
  "cluster": "cluster1",
  "namespace": "ns1",
  "pod": "pod1",
  "type": "Warning",
  "reason": "BackOff",
  "message": "Event message"
}
```

**Note:** Only the original event data is indexed. No metadata fields (like `@timestamp`, `event_id`, `event_type`, etc.) or extension fields (like `ext_*`) are added to the document.

**Supported Options:**

- `username`: Elasticsearch username (for Basic auth)
- `password`: Elasticsearch password (for Basic auth)
- `index`: Override index name from URL (supports date placeholders like `{YYYY.MM.DD}`)
- `id`: Document ID (if not specified in URL)

## Complete Examples

### Example 1: Send Pod Error Events to Webhook

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

### Example 2: Forward Events to Alert Subject

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

### Example 3: Index Events to Elasticsearch

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

### Example 3b: Index Events to Elasticsearch with Date-based Indices

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
    index: ops-events-{YYYY.MM.DD}  # Creates daily indices like ops-events-2024.01.19
    username: elastic
    password: changeme
  keywords:
    include:
      - "Error"
    matchType: CONTAINS
```

### Example 4: Node Event Processing

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

## Best Practices

1. **Use Regular Expressions for Complex Matching**: When matching multiple conditions, use regular expressions for more precise control.

2. **Use Exclude Rules Appropriately**: Use `exclude` to filter out unwanted events and reduce false positives.

3. **Use Wildcards for Event Forwarding**: In event forwarding scenarios, use wildcards to maintain event subject structure for subsequent processing.

4. **Elasticsearch Index Naming**: Use meaningful index names, consider organizing by date or cluster name.

5. **Monitor EventHooks Status**: Monitor EventHooks trigger status through Prometheus metrics to detect configuration issues promptly.

## Troubleshooting

### View EventHooks Status

```bash
kubectl get eventhooks -n ops-system
kubectl describe eventhooks <name> -n ops-system
```

### View Controller Logs

```bash
kubectl logs -n ops-system deployment/ops-controller-manager | grep eventhook
```

### Check Metrics

```bash
# View EventHooks trigger count
curl http://localhost:8080/metrics | grep ops_controller_eventhooks_status
```

### Common Issues

1. **Events Not Triggering**:
   - Check if `subject` correctly matches event subjects
   - Check if `keywords` configuration is correct
   - View Controller logs for errors

2. **Webhook Send Failure**:
   - Check if URL is accessible
   - Check network connection
   - View error messages in Controller logs

3. **Elasticsearch Index Failure**:
   - Check Elasticsearch connection
   - Verify authentication credentials
   - Check index permissions

## Related Documentation

- [Event System Documentation](./opscontroller-events.md)
- [Metrics Documentation](./opscontroller-metrics.md)
