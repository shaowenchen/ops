## Ops-controller-manager task 对象

### 直接使用 `create` 子命令创建

```bash
/usr/local/bin/opscli create task --name t1 --typeref host --nameref dev1  --filepath ./task/get-osstaus.yaml
```

通过 `--typeref host --nameref dev1` 指定任务执行的主机。

### 使用 yaml 文件创建

如果不指定 `typeref` ，那么任务将在 ops-controller-manager pod 中执行。

下面是一个定时任务，每分钟检查一次 http 状态码，如果不是 200，那么发送通知。

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
    - name: get-status
      content: curl -I -m 10 -o /dev/null -s -w %{http_code} ${url}
    - name: notification
      when: ${steps.get-status.output} != ${expect}
      content: |
        curl -X POST 'https://365.kdocs.cn/woa/api/v1/webhook/send?key=xxx' -H 'content-type: application/json' -d '{ "msgtype": "text", "text": { "content": "${message}" } }'
```

### 查看对象

```bash
kubectl get task t1 -n ops-system

kubectl get task -n ops-system
NAME                             CRONTAB       TYPEREF   NAMEREF   NODENAME   ALL    STARTTIME   RUNSTATUS
alert-http-status-dockermirror   */1 * * * *
```

### 导出结果变量

**默认输出变量：**

每个 step 的输出会自动在后续 step 中可用。可以直接引用或使用路径语法：

- `${output}` 或 `${result}` - 引用前一个 step 的输出（直接引用，推荐）
- `${steps.{stepName}.output}` - 通过名称引用指定 step 的输出（路径引用）

示例（直接引用）：
```yaml
steps:
  - name: get-status
    content: curl -I -m 10 -o /dev/null -s -w %{http_code} ${url}
  - name: notification
    when: ${output} != ${expect}  # 直接引用（推荐）
    content: echo "状态码是 ${output}"
```

或使用路径语法：
```yaml
steps:
  - name: get-status
    content: curl -I -m 10 -o /dev/null -s -w %{http_code} ${url}
  - name: notification
    when: ${steps.get-status.output} != ${expect}  # 路径引用
    content: echo "状态码是 ${steps.get-status.output}"
```

**使用 OPS_RESULT 标记导出：**

如果需要将结果导出供 Pipeline 中的其他任务使用，可在 step 输出中使用 `OPS_RESULT:` 标记：

```yaml
steps:
  - name: build-step
    content: |
      docker build -t myapp:v1.0.0 .
      echo "OPS_RESULT:image=registry.example.com/myapp:v1.0.0"
      echo "OPS_RESULT:tag=v1.0.0"
```

支持的格式：
- `OPS_RESULT:key=value`
- `OPS_RESULT:key:value`
- `OPS_RESULT:{"key":"value"}` (JSON 格式，支持多个结果)

**使用 key:value 格式导出（向后兼容）：**

最后一个 step 的输出如果是 `key:value` 格式，会自动导出：

```yaml
steps:
  - name: build-step
    content: |
      docker build -t myapp:v1.0.0 .
      echo "image:registry.example.com/myapp:v1.0.0"
```
