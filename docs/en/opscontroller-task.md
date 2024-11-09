### **Ops-controller-manager Task Object**

The `Task` object in the Ops Controller defines the operations or tasks that need to be executed, either on specific hosts or within the Ops Controller pod. You can create and manage `Task` objects using `opscli` commands or YAML files.

#### **Create Task Using `opscli` Command**

To create a `Task` object directly using the `create` sub-command:

```bash
/usr/local/bin/opscli create task --name t1 --typeref host --nameref dev1 --filepath ./task/get-osstaus.yaml
```

- **`name`**: The name of the task (e.g., `t1`).
- **`typeref`**: Specifies the type of resource the task will execute on (e.g., `host`).
- **`nameref`**: Specifies the name of the host (e.g., `dev1`).
- **`filepath`**: The path to the YAML file containing the task definition.

#### **Create Task Using YAML File**

Alternatively, you can define the `Task` object in a YAML file and apply it using `kubectl`. Below is an example YAML file for a scheduled task that checks the HTTP status of a URL every minute and sends a notification if the status code is not 200.

```yaml
apiVersion: crd.chenshaowen.com/v1
kind: Task
metadata:
  name: alert-http-status-dockermirror
  namespace: ops-system
spec:
  desc: alert
  crontab: "*/1 * * * *"
  variables:
    url:
      default: http://1.1.1.1:5000/
    expect:
      default: "200"
    message:
      default: ${url} http status is not ${expect}
  steps:
    - name: get status
      content: curl -I -m 10 -o /dev/null -s -w %{http_code} ${url}
    - name: notifaction
      when: ${result} != ${expect}
      content: |
        curl -X POST 'https://xz.wps.cn/api/v1/webhook/send?key=xxx' -H 'content-type: application/json' -d '{ "msgtype": "text", "text": { "content": "${message}" } }'
```

In this YAML:

- **`desc`**: A description of the task (e.g., `alert`).
- **`crontab`**: The cron schedule (e.g., `*/1 * * * *` means every minute).
- **`variables`**: Defines task-specific variables, like `url`, `expect`, and `message`.
- **`steps`**: Defines the individual operations or commands that should be executed:
  - `get status`: Executes a `curl` command to get the HTTP status code.
  - `notifaction`: Sends a notification if the HTTP status code does not match the expected value.

You can apply this file using:

```bash
kubectl apply -f task.yaml
```

#### **View Task Object Status**

To view the status of a specific `Task` object, use the following command:

```bash
kubectl get task t1 -n ops-system
```

To list all tasks in the `ops-system` namespace:

```bash
kubectl get task -n ops-system
```

Example output:

```bash
NAME                             CRONTAB       TYPEREF   NAMEREF   NODENAME   ALL    STARTTIME   RUNSTATUS
alert-http-status-dockermirror   */1 * * * *   host      dev1      node1      true   2024-11-09  successed
```

In the output:

- **`NAME`**: The name of the task.
- **`CRONTAB`**: The cron schedule for the task.
- **`TYPEREF`**: The type of the resource the task will run on (e.g., `host`).
- **`NAMEREF`**: The reference name of the target resource (e.g., `dev1`).
- **`NODENAME`**: The name of the node where the task is running.
- **`ALL`**: Indicates whether the task runs on all nodes.
- **`STARTTIME`**: The time the task was started.
- **`RUNSTATUS`**: The current status of the task (e.g., `successed`).
